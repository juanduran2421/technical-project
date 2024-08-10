package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"technical-proyect/shared"
)

var (
	secretsClient secretsmanageriface.SecretsManagerAPI

	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-users"

	ErrInvalidRequestBody = errors.New("invalid json body")
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func getPaySafeSecret() (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String("paysafe-token"),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := secretsClient.GetSecretValue(input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(result.SecretString), nil
}

func (req *request) madePayment(ctx context.Context, paymentInfo *shared.PaymentInput) error {
	token, err := getPaySafeSecret()
	if err != nil {
		return err
	}

	return shared.MadePaymentRequest(paymentInfo, token)
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

		return shared.NewInvalidRequestError(ErrInvalidRequestBody, req.Headers), nil
	}

	err = shared.ValidatePaymentInfo(paymentInfo)
	if err != nil {
		return shared.NewInvalidRequestError(err, req.Headers), nil
	}

	err = createUserRequest.madePayment(ctx, paymentInfo)
	if err != nil {
		fmt.Println("InternalServerError", err)

		return shared.NewInternalServerError(req.Headers), nil
	}

	return shared.NewSuccessResponse(paymentInfo, req.Headers), nil
}

func main() {
	cfg := &aws.Config{}
	secretsClient = secretsmanager.New(session.Must(session.NewSession(cfg)))

	lambda.Start(HandleRequest)
}
