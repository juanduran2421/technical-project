package main

import (
	"context"
	"encoding/json"
	"fmt"

	"technical-proyect/shared"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	secretClient *secretsmanager.Client

	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-payments"
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func (req *request) savePayment(ctx context.Context, paymentOutput *shared.PaymentOutput) *events.APIGatewayProxyResponse {
	item, err := attributevalue.MarshalMapWithOptions(paymentOutput, shared.EncodeWithJSONKey)
	if err != nil {
		fmt.Printf("Marshal payment error %v\n", err)

		return shared.NewInvalidRequestError(err, req.Headers)
	}

	_, err = dynamoDbObj.PutItem(ctx,
		&dynamodb.PutItemInput{
			Item:                item,
			TableName:           aws.String(tableName),
			ConditionExpression: aws.String("attribute_not_exists(payment_id)"),
		},
	)
	if err != nil {
		fmt.Printf("Put payment error %v\n", err)

		return shared.NewInternalServerError(req.Headers)
	}

	return shared.NewSuccessResponse(paymentOutput, req.Headers)
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

	paymentOutput, err := shared.MakePaymentRequest(paymentInfo, token)
	if err != nil {
		return shared.PaymentOutput{}, err
	}

	paymentOutput.Username = req.Headers["username"]

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

// HandleRequest handler of the apiGateway request
func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	createUserRequest := &request{
		APIGatewayProxyRequest: &req,
	}

	paymentInfo, err := parseRequest(createUserRequest.Body)
	if err != nil {
		fmt.Printf("Parse payment request error %v\n", err)

		return shared.NewInvalidRequestError(err, req.Headers), nil
	}

	err = shared.ValidatePaymentInfo(paymentInfo)
	if err != nil {
		return shared.NewInvalidRequestError(err, req.Headers), nil
	}

	paymentOutput, err := createUserRequest.madePayment(ctx, paymentInfo)
	if err != nil {
		fmt.Printf("Internal server error %v\n", err)

		return shared.NewInternalServerError(req.Headers), nil
	}

	return createUserRequest.savePayment(ctx, &paymentOutput), nil
}

func main() {
	config, _ := config.LoadDefaultConfig(context.TODO())
	secretClient = secretsmanager.NewFromConfig(config)
	dynamoDbObj = dynamodb.NewFromConfig(config)

	lambda.Start(HandleRequest)
}
