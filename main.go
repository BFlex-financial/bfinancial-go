package bfinancial

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Payments struct {
	api  string
	auth string
}
type Pix struct {
	Amount      uint64 `json:"amount"`
	Payer_email string `json:"payer_email"`
}
type Card struct {
	Number           string  `json:"number"`
	Amount           float64 `json:"amount"`
	Cvv              string  `json:"cvv"`
	Payer_email      string  `json:"payer_email"`
	Payer_name       string  `json:"payer_name"`
	Payer_cpf        string  `json:"payer_cpf"`
	Expiration_year  uint64  `json:"expiration_year"`
	Expiration_month uint64  `json:"expiration_month"`
}

type PaymentResultPix struct {
	Code uint64 `json:"code"`
	Data struct {
		PaymentId uint64 `json:"payment_id"`
		QrCode    struct {
			Base64  string `json:"base64"`
			Literal string `json:"literal"`
		} `json:"qr_code"`
	} `json:"data"`
}
type PaymentResultCard struct {
	Code uint64 `json:"code"`
	Data struct {
		PaymentId   uint64 `json:"payment_id"`
		TotalAmount int64  `json:"total_amount"`
		Increase    int64  `json:"increase"`
	} `json:"data"`
}

type PaymentResult struct {
	Pix  PaymentResultPix
	Card PaymentResultCard
}

func (payments *Payments) Create(data interface{}) (PaymentResult, error) {
	pix, pix_ok := data.(Pix)
	card, card_ok := data.(Card)

	if !pix_ok && !card_ok {
		return PaymentResult{}, errors.New("not a suported type")
	}
	if pix_ok {
		req_body, err := json.Marshal(struct {
			Pix
			Method string `json:"method"`
		}{
			Method: "Pix",
			Pix:    pix,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshe the response: %v", err)
			panic(1)
		}
		req, err := http.NewRequest(http.MethodPost, payments.api+"/payment/create", bytes.NewBuffer(req_body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot create the request: %v", err)
			panic(1)
		}
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Authorization-key", payments.auth)
		client := http.Client{
			Timeout: time.Second * 30,
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot send the request: %v", err)
			panic(1)
		}
		bd, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the response: %v", err)
			panic(1)
		}
		var rt PaymentResultPix
		err = json.Unmarshal(bd, &rt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse the response: %v", err)
			panic(1)
		}
		return PaymentResult{rt, PaymentResultCard{}}, nil

	} else {
		req_body, err := json.Marshal(struct {
			Card
			Method string `json:"method"`
		}{
			Method: "Card",
			Card:   card,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshe the response: %v\n", err)
			panic(1)
		}
		req, err := http.NewRequest(http.MethodPost, payments.api+"/payment/create", bytes.NewBuffer(req_body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot create the request: %v\n", err)
			panic(1)
		}
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Authorization-key", payments.auth)
		client := http.Client{
			Timeout: time.Second * 30,
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot send the request: %v\n", err)
			panic(1)
		}
		if res.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "error request, status code: %v\n", res.StatusCode)
			panic(1)
		}
		bd, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the response: %v\n", err)
			panic(1)
		}
		var rt PaymentResultCard
		fmt.Printf("%s \n", res.Status)
		err = json.Unmarshal(bd, &rt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse the response: %v\n", err)
			panic(1)
		}
		return PaymentResult{PaymentResultPix{}, rt}, nil
	}
}

type Client struct {
	Payments Payments
	auth     string
}

func (c *Client) Login(auth string) {
	*c = Client{
		Payments: Payments{
			api:  "http://127.0.0.1:8080",
			auth: auth,
		},
		auth: auth,
	}
}
