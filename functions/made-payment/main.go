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
	"github.com/juanduran2421/technical-proyect/shared"
)

var (
	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-users"

	ErrInvalidRequestBody = errors.New("invalid json body")
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func (req *request) madePayment(ctx context.Context, paymentInfo *shared.PaymentInput) error {
	err := shared.MadePaymentRequest(paymentInfo)

	return err
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
		return shared.NewInternalServerError(req.Headers), nil
	}

	return shared.NewSuccessResponse(paymentInfo, req.Headers), nil
}

func main() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)

	lambda.Start(HandleRequest)
}
