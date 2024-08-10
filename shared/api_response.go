package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errInternalServerError = errors.New("Internal server error")
)

func NewInvalidRequestError(err error, headers map[string]string) *events.APIGatewayProxyResponse {
	response := map[string]interface{}{
		"code":      1400,
		"http_code": 400,
		"message":   fmt.Sprintf("Invalid request: %s", err.Error()),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errInternalServerError.Error(),
			Headers:    headers,
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       string(jsonResponse),
		Headers:    headers,
	}
}

func NewSuccessResponse(body interface{}, headers map[string]string) *events.APIGatewayProxyResponse {
	jsonResponse, err := json.Marshal(body)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errInternalServerError.Error(),
			Headers:    headers,
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonResponse),
		Headers:    headers,
	}
}

func NewInternalServerError(headers map[string]string) *events.APIGatewayProxyResponse {
	response := map[string]interface{}{
		"code":      1500,
		"http_code": 500,
		"message":   errInternalServerError.Error(),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errInternalServerError.Error(),
			Headers:    headers,
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       string(jsonResponse),
		Headers:    headers,
	}
}
