package shared

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestSuccessRequest(t *testing.T) {
	c := require.New(t)

	httpmock.Activate()

	defer httpmock.Deactivate()

	paymentOutput := map[string]interface{}{
		"id":     "dummy_id",
		"status": "COMPLETED",
	}

	httpmock.RegisterResponder("POST", "https://api.test.paysafe.com/cardpayments/v1/accounts/1002776850/auths/",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(http.StatusOK, paymentOutput)
			return resp, err
		},
	)

	_, err := MakePaymentRequest(&PaymentInput{
		Card: &CardDetails{},
	}, "token")
	c.NoError(err)
}

func TestParseFailedResponse(t *testing.T) {
	c := require.New(t)

	paymentInput := &PaymentInput{
		Card: &CardDetails{},
	}

	mapInvalidResponse := map[string]interface{}{
		"id": 123,
	}

	mapBytes, err := json.Marshal(mapInvalidResponse)
	c.NoError(err)

	_, err = parseFailedResponse(paymentInput, mapBytes)
	c.Error(err)

	mapInvalidResponse = map[string]interface{}{
		"id": "dummy",
	}

	mapBytes, err = json.Marshal(mapInvalidResponse)
	c.NoError(err)

	_, err = parseFailedResponse(paymentInput, mapBytes)
	c.NoError(err)
}

func TestParseSuccessResponse(t *testing.T) {
	c := require.New(t)

	mapInvalidResponse := map[string]interface{}{
		"id":       "dummy",
		"username": 123,
	}

	mapBytes, err := json.Marshal(mapInvalidResponse)
	c.NoError(err)

	_, err = parseSuccessResponse(mapBytes)
	c.Error(err)

	mapInvalidResponse = map[string]interface{}{
		"id":       "dummy",
		"username": "dummy_user",
	}

	mapBytes, err = json.Marshal(mapInvalidResponse)
	c.NoError(err)

	_, err = parseSuccessResponse(mapBytes)
	c.NoError(err)
}
