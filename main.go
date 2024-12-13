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
	Payer_cpf   string `json:"payer_cpf"`
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

// {
// 	"checkout_id": "aaa",
// 	"payment_type": {
// 			"payer_email": "aaaa@gmail.com",
// 			"payer_cpf": "08789662938",
// 			"amount": 1,
// 			"method": "Pix"
// 	},
// 	"method": "Checkout"
// }

type Checkout struct {
	Checkout_id string `json:"checkout_id"`
	Card        Card
	Pix         Pix
}

type PaymentResultPix struct {
	Code uint64 `json:"code"`
	Data struct {
		Error     string `json:"error"`
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
		Error       string `json:"error"`
		PaymentId   uint64 `json:"payment_id"`
		TotalAmount int64  `json:"total_amount"`
		Increase    int64  `json:"increase"`
	} `json:"data"`
}
type PaymentResultCheckout struct {
	Code uint64 `json:"code"`
	Data struct {
		QrCode struct {
			Base64  string `json:"base64"`
			Literal string `json:"literal"`
		} `json:"qr_code"`
		Error       string `json:"error"`
		PaymentId   uint64 `json:"payment_id"`
		TotalAmount int64  `json:"total_amount"`
		Increase    int64  `json:"increase"`
	} `json:"data"`
}

type PaymentResult struct {
	Pix      PaymentResultPix
	Card     PaymentResultCard
	Checkout PaymentResultCheckout
}

func call(payments Payments, req_body []byte) []byte {
	req, err := http.NewRequest(http.MethodPost, payments.api+"/payment/create", bytes.NewBuffer(req_body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create the request: %v", err)
		panic(1)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization-key", "Bearer "+payments.auth)
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
	return bd
}

func (payments *Payments) Create(data interface{}) (PaymentResult, error) {
	switch data := data.(type) {
	case Pix:
		req_body, err := json.Marshal(struct {
			Pix
			Method string `json:"method"`
		}{
			Method: "Pix",
			Pix:    data,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshe the response: %v", err)
			panic(1)
		}
		bd := call(*payments, req_body)
		var rt PaymentResultPix
		err = json.Unmarshal(bd, &rt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse the response: %v", err)
			panic(1)
		}
		return PaymentResult{rt, PaymentResultCard{}, PaymentResultCheckout{}}, nil
	case Card:
		req_body, err := json.Marshal(struct {
			Card
			Method string `json:"method"`
		}{
			Method: "Card",
			Card:   data,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshe the response: %v\n", err)
			panic(1)
		}
		bd := call(*payments, req_body)
		var rt PaymentResultCard
		err = json.Unmarshal(bd, &rt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse the response: %v\n", err)
			panic(1)
		}
		return PaymentResult{PaymentResultPix{}, rt, PaymentResultCheckout{}}, nil
	case Checkout:
		var req_body []byte
		var err error
		if data.Card.Number == "" {
			req_body, err = json.Marshal(
				struct {
					Checkout_id  string `json:"checkout_id"`
					Method       string `json:"method"`
					Payment_type struct {
						Pix
						Method string `json:"method"`
					} `json:"payment_type"`
				}{
					Checkout_id: data.Checkout_id,
					Method:      "Checkout",
					Payment_type: struct {
						Pix
						Method string `json:"method"`
					}{
						Pix:    data.Pix,
						Method: "Pix",
					},
				})
		} else if data.Pix.Amount == 0 {
			req_body, err = json.Marshal(
				struct {
					Checkout_id  string `json:"checkout_id"`
					Method       string `json:"method"`
					Payment_type struct {
						Card
						Method string `json:"method"`
					} `json:"payment_type"`
				}{
					Checkout_id: data.Checkout_id,
					Method:      "Checkout",
					Payment_type: struct {
						Card
						Method string `json:"method"`
					}{
						Card:   data.Card,
						Method: "Card",
					},
				})
		} else {
			fmt.Fprintln(os.Stderr, "incorrect request")
			panic(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshe the response: %v", err)
			panic(1)
		}
		bd := call(*payments, req_body)
		fmt.Printf("%s\n", bd)
		var rt PaymentResultCheckout
		err = json.Unmarshal(bd, &rt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse the response: %v", err)
			panic(1)
		}
		return PaymentResult{PaymentResultPix{}, PaymentResultCard{}, rt}, nil
	default:
		return PaymentResult{}, errors.New("not a suported type")
	}

}

type Client struct {
	Payments Payments
	auth     string
}

func (c *Client) Login(auth string) {
	*c = Client{
		Payments: Payments{
			api:  "http://127.0.0.1:8080/api",
			auth: auth,
		},
		auth: auth,
	}
}
