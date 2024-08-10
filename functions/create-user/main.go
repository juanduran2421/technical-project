package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"technical-proyect/shared"
)

var (
	dynamoDbObj = &dynamodb.Client{}
	tableName   = "technical-test-users"

	emailRegexp = regexp.MustCompile("\\A[a-zA-Z0-9\\.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*\\z")

	ErrInvalidRequestBody      = errors.New("invalid json body")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidUserAlreadyExits = errors.New("user already exits")
)

type request struct {
	*events.APIGatewayProxyRequest
	err error
}

func validateParams(user *shared.UserModelAuth) error {
	password := user.Password
	if password == "" || len(password) < 16 {
		return ErrInvalidPassword
	}

	email := user.Username
	if !emailRegexp.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// EncodeWithJSONKey use the tag in the struct to encode the item
func EncodeWithJSONKey(eo *attributevalue.EncoderOptions) {
	eo.TagKey = "json"
}

func (req *request) saveUser(ctx context.Context, userInput *shared.UserModelAuth) *events.APIGatewayProxyResponse {
	checksum := sha256.Sum256([]byte(userInput.Password))
	encodedPassword := base64.StdEncoding.EncodeToString(checksum[:])

	user := &shared.UserModelAuth{
		Username: userInput.Username,
		Password: encodedPassword,
	}

	item, err := attributevalue.MarshalMapWithOptions(user, EncodeWithJSONKey)
	if err != nil {
		fmt.Println("MarshalMapWithOptionsError", err)

		return shared.NewInternalServerError(req.Headers)
	}

	_, err = dynamoDbObj.PutItem(ctx,
		&dynamodb.PutItemInput{
			Item:                item,
			TableName:           aws.String(tableName),
			ConditionExpression: aws.String("attribute_not_exists(username)"),
		},
	)

	var ccFailed *types.ConditionalCheckFailedException
	if errors.As(err, &ccFailed) {
		return shared.NewInvalidRequestError(ErrInvalidUserAlreadyExits, req.Headers)

	}

	if err != nil {
		fmt.Println("PutItemError", err)

		return shared.NewInternalServerError(req.Headers)
	}

	return nil
}

func parseRequest(body string) (*shared.UserModelAuth, error) {
	userModel := &shared.UserModelAuth{}

	err := json.Unmarshal([]byte(body), userModel)
	if err != nil {
		return nil, err
	}

	return userModel, nil
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	createUserRequest := &request{
		APIGatewayProxyRequest: &req,
	}

	user, err := parseRequest(createUserRequest.Body)
	if err != nil {
		fmt.Println("ParseRequestError", err)

		return shared.NewInvalidRequestError(ErrInvalidRequestBody, req.Headers), nil
	}

	err = validateParams(user)
	if err != nil {
		return shared.NewInvalidRequestError(err, req.Headers), nil
	}

	response := createUserRequest.saveUser(ctx, user)
	if response != nil {
		return response, nil
	}

	return shared.NewSuccessResponse(user, req.Headers), nil
}

func main() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)

	lambda.Start(HandleRequest)
}
