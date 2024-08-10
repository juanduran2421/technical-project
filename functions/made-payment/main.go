package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"net/http"
	"technical-proyect/shared"
)

var (
	secretClient *secretsmanager.Client

	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-users"

	ErrInvalidRequestBody = errors.New("invalid json body")
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func getPaySafeSecret(ctx context.Context) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String("paysafe-token"),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := secretClient.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(result.SecretString), nil
}

func (req *request) madePayment(ctx context.Context, paymentInfo *shared.PaymentInput) (shared.PaymentOutput, error) {
	token, err := getPaySafeSecret(ctx)
	if err != nil {
		return shared.PaymentOutput{}, err
	}

	paymentOutput, err := shared.MadePaymentRequest(paymentInfo, token)
	if err != nil {
		return shared.PaymentOutput{}, err
	}

	paymentOutput.Username = req.RequestContext.Authorizer["username"].(string)

	return paymentOutput, nil
}

func parseRequest(body string) (*shared.PaymentInput, error) {
	paymentInfo := &shared.PaymentInput{}

	err := json.Unmarshal([]byte(body), paymentInfo)
	if err != nil {
		return nil, err
	}

	return paymentInfo, nil
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	createUserRequest := &request{
		APIGatewayProxyRequest: &req,
	}

	paymentInfo, err := parseRequest(createUserRequest.Body)
	if err != nil {
		fmt.Println("ParseRequestError", err)

		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       ".......",
			Headers:    req.Headers,
		}, nil
	}

	err = shared.ValidatePaymentInfo(paymentInfo)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "dasdasdasd",
			Headers:    req.Headers,
		}, nil
	}

	paymentOutput, err := createUserRequest.madePayment(ctx, paymentInfo)
	if err != nil {
		fmt.Println("InternalServerError", err)

		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "lsalalslaslsl",
			Headers:    req.Headers,
		}, nil
	}

	fmt.Println("paymentOutput", paymentOutput)

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "asdasdasdasdas",
		Headers:    req.Headers,
	}, nil
}

func main() {
	config, _ := config.LoadDefaultConfig(context.TODO())
	secretClient = secretsmanager.NewFromConfig(config)

	lambda.Start(HandleRequest)
}
