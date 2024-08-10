package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModelsValidatePaymentInfo(t *testing.T) {
	c := require.New(t)

	paymentInput := &PaymentInput{}
	c.Equal(ValidatePaymentInfo(paymentInput), errMissingCardInfo)

	paymentInput.Card = &CardDetails{}
	c.Equal(ValidatePaymentInfo(paymentInput), errMissingBillingDetails)

	paymentInput.BillingDetails = &BillingInfo{}
	c.Equal(ValidatePaymentInfo(paymentInput), errMissingProfile)

	paymentInput.Profile = &Profile{}
	c.Equal(ValidatePaymentInfo(paymentInput), errInvalidCvv)

	paymentInput.Card.Cvv = "123"
	c.Equal(ValidatePaymentInfo(paymentInput), errMissingCardNumber)

	paymentInput.Card.CardNum = "123xxxxx1234"
	c.Equal(ValidatePaymentInfo(paymentInput), errMissingExpirationInfo)

	paymentInput.Card.CardExpirationDate = &CardExpiration{
		Year: 2021,
	}
	c.Equal(ValidatePaymentInfo(paymentInput), errInvalidYearNumber)

	paymentInput.Card.CardExpirationDate = &CardExpiration{
		Year:  time.Now().Year(),
		Month: int(time.Now().Add(time.Hour * 24 * 33).Month()),
	}
	c.NoError(ValidatePaymentInfo(paymentInput))
}
