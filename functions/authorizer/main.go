package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/juanduran2421/technical-proyect/shared"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-users"
)

// DecodeWithJSONKey use the tag in the struct to decode the item
func DecodeWithJSONKey(do *attributevalue.DecoderOptions) {
	do.TagKey = "json"
}

func isAuthorized(ctx context.Context, username string, password string) (shared.UserModelAuth, bool) {
	user := shared.UserModelAuth{}

	resp, err := dynamoDbObj.GetItem(ctx,
		&dynamodb.GetItemInput{
			Key: map[string]types.AttributeValue{
				"username": &types.AttributeValueMemberS{
					Value: username,
				},
			},
			TableName: aws.String(tableName)},
	)
	if err != nil {
		fmt.Printf("Cannot Get Item from Table %s: %v\n", tableName, err)

		return shared.UserModelAuth{}, false
	}

	err = attributevalue.UnmarshalMapWithOptions(resp.Item, &user, DecodeWithJSONKey)
	if err != nil {
		fmt.Printf("Unmarshal result %s: %v\n", tableName, err)

		return user, false
	}

	checksum := sha256.Sum256([]byte(password))

	encodedPassword := base64.StdEncoding.EncodeToString(checksum[:])
	if user.Password == encodedPassword {
		return user, true
	}

	return shared.UserModelAuth{}, false
}

func handleRequest(ctx context.Context, apiGatewayRequest events.APIGatewayProxyRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	user, isAuth := isAuthorized(ctx, apiGatewayRequest.Headers["username"], apiGatewayRequest.Headers["password"])
	if isAuth {
		claims := make(map[string]interface{})

		claims["client_id"] = user.ClientID

		return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: true, Context: claims}, nil
	}

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}, nil

}

func main() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)

	lambda.Start(handleRequest)
}
