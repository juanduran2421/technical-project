package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MakePaymentRequest make the request to do the payment with the provider
func MakePaymentRequest(paymentInfo *PaymentInput, token string) (PaymentOutput, error) {
	paymentInfo.MerchantRefNum = "merchant 03.24.17_3"
	paymentInfo.SettleWithAuth = false

	url := "https://api.test.paysafe.com/cardpayments/v1/accounts/1002776850/auths/"

	paymentBytes, err := json.Marshal(paymentInfo)
	if err != nil {
		return PaymentOutput{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(paymentBytes))
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PaymentOutput{}, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Println("Close error", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PaymentOutput{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return parseFailedResponse(paymentInfo, body)
	}

	return parseSuccessResponse(body)
}

func parseFailedResponse(paymentInfo *PaymentInput, body []byte) (PaymentOutput, error) {
	output := &RequestFailed{}
	paymentOutput := PaymentOutput{}

	err := json.Unmarshal(body, output)
	if err != nil {
		return PaymentOutput{}, err
	}

	paymentOutput.PaymentID = output.ID
	paymentOutput.Card = paymentInfo.Card
	paymentOutput.Card.Cvv = ""

	if len(paymentOutput.Card.CardNum) > 4 {
		paymentOutput.Card.CardNum = ""
		paymentOutput.Card.LastDigits = paymentOutput.Card.CardNum[len(paymentOutput.Card.CardNum)-4:]
	}
	paymentOutput.BillingDetails = paymentInfo.BillingDetails
	paymentOutput.Profile = paymentInfo.Profile

	paymentOutput.Status = "Error"
	paymentOutput.ErrorMessage = output.Error.Message

	return paymentOutput, nil
}

func parseSuccessResponse(body []byte) (PaymentOutput, error) {
	paymentOutput := PaymentOutput{}

	err := json.Unmarshal(body, &paymentOutput)
	if err != nil {
		return PaymentOutput{}, err
	}

	mapResponse := map[string]interface{}{}
	err = json.Unmarshal(body, &mapResponse)
	if err != nil {
		return PaymentOutput{}, err
	}

	paymentOutput.PaymentID = mapResponse["id"].(string)

	return paymentOutput, nil
}
