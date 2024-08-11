package main

import (
	"context"
	"errors"
	"fmt"

	"technical-proyect/shared"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-payments"

	errPaymentNotFound = errors.New("Payment not found")
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func (req *request) getListOfPayment(ctx context.Context) *events.APIGatewayProxyResponse {
	key := map[string]types.AttributeValue{
		":username": &types.AttributeValueMemberS{
			Value: req.Headers["username"],
		},
	}

	headers := req.Headers
	headers["Content-Type"] = "application/json"

	if req.QueryStringParameters["payment_id"] != "" {
		return req.getSpecificPayment(ctx)
	}

	resp, err := dynamoDbObj.Query(ctx,
		&dynamodb.QueryInput{
			ExpressionAttributeValues: key,
			KeyConditionExpression:    aws.String("username=:username"),
			IndexName:                 aws.String("username-index"),
			TableName:                 aws.String(tableName),
		},
	)
	if err != nil {
		fmt.Printf("Put payment error %v\n", err)

		return shared.NewInternalServerError(headers)
	}

	if len(resp.Items) == 0 {
		return shared.NewNotFoundError(errPaymentNotFound, headers)
	}

	listOfPayments := []shared.PaymentOutput{}

	err = attributevalue.UnmarshalListOfMapsWithOptions(resp.Items, &listOfPayments, shared.DecodeWithJSONKey)
	if err != nil {
		fmt.Printf("Unmarshal result %s: %v\n", tableName, err)

		return shared.NewInternalServerError(headers)
	}

	return shared.NewSuccessResponse(listOfPayments, headers)
}

func (req *request) getSpecificPayment(ctx context.Context) *events.APIGatewayProxyResponse {
	key := map[string]types.AttributeValue{
		"username": &types.AttributeValueMemberS{
			Value: req.Headers["username"],
		},
		"payment_id": &types.AttributeValueMemberS{
			Value: req.QueryStringParameters["payment_id"],
		},
	}

	headers := req.Headers
	headers["Content-Type"] = "application/json"

	resp, err := dynamoDbObj.GetItem(ctx,
		&dynamodb.GetItemInput{
			Key:       key,
			TableName: aws.String(tableName),
		},
	)
	if err != nil {
		fmt.Printf("Get payment error %v\n", err)

		return shared.NewInternalServerError(headers)
	}

	if len(resp.Item) == 0 {
		return shared.NewNotFoundError(errPaymentNotFound, headers)
	}

	payment := &shared.PaymentOutput{}

	err = attributevalue.UnmarshalMapWithOptions(resp.Item, payment, shared.DecodeWithJSONKey)
	if err != nil {
		fmt.Printf("Unmarshal result %s: %v\n", tableName, err)

		return shared.NewInternalServerError(headers)
	}

	return shared.NewSuccessResponse(payment, headers)
}

// HandleRequest handler of the apiGateway request
func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	createUserRequest := &request{
		APIGatewayProxyRequest: &req,
	}

	return createUserRequest.getListOfPayment(ctx), nil
}

func main() {
	config, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(config)

	lambda.Start(HandleRequest)
}
