package shared

import (
	"errors"
	"time"
)

var (
	errMissingCardInfo       = errors.New("missing card info")
	errMissingBillingDetails = errors.New("missing billing details")
	errMissingProfile        = errors.New("missing profile")
	errMissingExpirationInfo = errors.New("missing card expiration info")
	errMissingCardNumber     = errors.New("missing card number")
	errInvalidCvv            = errors.New("invalid cvv number")
	errInvalidYearNumber     = errors.New("invalid year number")
	errInvalidMonthNumber    = errors.New("invalid month number")
)

// UserModelAuth basic struct of the users saved
type UserModelAuth struct {
	Username string `json:"username"`
	ClientID string `json:"client_id"`
	Password string `json:"password"`
}

type CardExpiration struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}
type CardDetails struct {
	CardExpiry *CardExpiration `json:"cardExpiry"`
	Cvv        string          `json:"cvv"`
	CardNum    string          `json:"cardNum"`
}

type BillingInfo struct {
	Zip     string `json:"zip"`
	Street  string `json:"street"`
	State   string `json:"state"`
	Country string `json:"country"`
	City    string `json:"city"`
}

type Profile struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
}

type PaymentInput struct {
	MerchantRefNum string       `json:"merchantRefNum"`
	Amount         int          `json:"amount"`
	SettleWithAuth bool         `json:"settleWithAuth"`
	DupCheck       bool         `json:"dupCheck"`
	Card           *CardDetails `json:"card"`
	BillingDetails *BillingInfo `json:"billingDetails"`
	PreAuth        bool         `json:"preAuth"`
	Profile        *Profile     `json:"profile"`
}

type PaymentOutput struct {
	PaymentID      string       `json:"payment_id"`
	Status         string       `json:"status"`
	ErrorMessage   string       `json:"error_message,omitempty"`
	Card           *CardDetails `json:"card"`
	BillingDetails *BillingInfo `json:"billingDetails"`
	Profile        *Profile     `json:"profile"`
}

func ValidatePaymentInfo(input *PaymentInput) error {
	if input.Card == nil {
		return errMissingCardInfo
	}

	if input.BillingDetails == nil {
		return errMissingBillingDetails
	}

	if input.Profile == nil {
		return errMissingProfile
	}

	return validateCardInfo(input.Card)
}

func validateCardInfo(details *CardDetails) error {
	if details.Cvv == "" || len(details.Cvv) != 3 {
		return errInvalidCvv
	}

	if details.CardNum == "" {
		return errMissingCardNumber
	}

	if details.CardExpiry == nil {
		return errMissingExpirationInfo
	}

	nowYear := time.Now().Year()
	if details.CardExpiry.Year < nowYear {
		return errInvalidYearNumber
	}

	if details.CardExpiry.Year == nowYear && details.CardExpiry.Month < int(time.Now().Month()) {
		return errInvalidMonthNumber
	}

	return nil
}
