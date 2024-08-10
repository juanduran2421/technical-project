package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

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
