package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func MadePaymentRequest(paymentInfo *PaymentInput) error {
	paymentInfo.MerchantRefNum = "merchant 03.24.17_3"
	paymentInfo.SettleWithAuth = true

	url := "https://api.test.paysafe.com/cardpayments/v1/accounts/1002776850/auths/"
	fmt.Println("URL:>", url)

	paymentBytes, err := json.Marshal(paymentInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(paymentBytes))
	req.Header.Set("Authorization", "Basic dGVzdF9qdWFuZHVyYW4yNDIxOkItcWEyLTAtNjZiNjc5NzItMC0zMDJjMDIxNDBkZTY2NTNhMGU2OTlkODg0MDEyMzFkMWFiNmRjODUxNzY1YzE4OGIwMjE0NTNkYjUxNmEyNWUxYWI0ODZjNmZmZjMwYjUwODAzMGNjNDMyYWQwNA==")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Println("Close error", err)
		}
	}()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}
