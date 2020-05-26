package mimipay

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	PAY_TYPE_WECHAT = "wechat"
	PAY_TYPE_ALIPAY = "alipay"
)

type MimiPayMakeOrderInput struct {
	UserKey      string  `json:"userKey"`
	Price        float64 `json:"price"`
	PayType      int     `json:"type"`
	OutTradeNo   string  `json:"outTradeNo"`
	OutUsreNo    string  `json:"outUserNo"`
	TradeSubject string  `json:"tradeSubject"`
	NotifyURL    string  `json:"notifyUrl"`
	Timeout      int     `json:"timeout"`
	ReturnURL    string  `json:"returnUrl"`
	Sign         string  `json:"key"`
}

type MimiPayMakeOrderData struct {
	OutTradeNo   string  `json:"outTradeNo"`
	PayNo        string  `json:"payNo"`
	PayType      int     `json:"payType"`
	QrCodeUrl    string  `json:"qrCodeUrl"`
	RealPrice    float64 `json:"realPrice"`
	TradeSubject string  `json:"tradeSubject"`
	Timeout      int     `json:"timeout"`
}

// example
// {
//		"data":{
//			"outTradeNo":"9e019d1d7f0330e",
//			"payNo":"51b43b68199a45d5b3eadf34f607aab8",
//			"payType":2,
//			"price":2.99,
//			"qrCodeUrl":"/image/get?key=99e98803d89bc433815e4afacfc847xx",
//			"realPrice":2.99,
//			"timeout":120,
//			"tradeSubject":"bdy_1"
//		},
//		"status":1,
//		"success":true
//	}
type MimiPayMakeOrderResponse struct {
	MimiPayMakeOrderData `json:"data"`
	Status               int  `json:"status"`
	Success              bool `json:"success"`
}

func mimiPayType(payType string) (r int) {
	if payType == PAY_TYPE_WECHAT {
		return 1
	}

	if payType == PAY_TYPE_ALIPAY {
		return 2
	}

	return
}

func MD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

func MD5WithLowerCase(text string) string {
	s := MD5(text)
	return strings.ToLower(s)
}

type MimiPay struct {
	Endpoint  string
	UserKey   string
	Secret    string
	NotifyURL string
	ReturnURL string
}

func NewMimiPay(userKey string, secret string, notifyURL string) (r *MimiPay) {
	r = &MimiPay{}
	r.UserKey = userKey
	r.Secret = secret
	r.NotifyURL = notifyURL
	r.Endpoint = "https://www.mimipay.cc/api/unified_order?format=json"
	return
}

// https://mimipay.cc/document/pay
// userKey + round(price*100) + type + outTradeNo + notifyUrl + secret
func (this *MimiPay) makeRequestSign(price float64, payType string, orderId string, notifyURL string) (r string) {
	pt := mimiPayType(payType)
	payTypeStr := fmt.Sprintf("%v", pt)
	iPrice := round(price * 100)
	ssPrice := fmt.Sprintf("%v", iPrice)
	ss := this.UserKey + ssPrice + payTypeStr + orderId + notifyURL + this.Secret
	r = MD5WithLowerCase(ss)
	return
}

func (this *MimiPay) MakeOrderToMimiPay(productName string, payType string, orderId string, orderUid string, price float64, timeOutSeconds int, returnURL string) (r *MimiPayMakeOrderResponse, err error) {
	var (
		endpoint  = this.Endpoint
		userKey   = this.UserKey
		notifyURL = this.NotifyURL
		sign      string
	)

	sign = this.makeRequestSign(price, payType, orderId, notifyURL)

	input := &MimiPayMakeOrderInput{}
	input.UserKey = userKey
	input.Price = price
	input.PayType = mimiPayType(payType)
	input.OutTradeNo = orderId
	input.OutUsreNo = orderUid
	input.ReturnURL = returnURL
	input.NotifyURL = notifyURL
	input.Sign = sign
	input.TradeSubject = productName
	input.Timeout = timeOutSeconds

	jsonRequest, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	body := bytes.NewBuffer(jsonRequest)
	request, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")

	log.Printf("mimipay request:\ncurl -H 'Content-Type: application/json' -d '%v' %v \n", string(jsonRequest), endpoint)
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	res, err := httpClient.Do(request)
	if err != nil {
		return
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s := string(b)
		log.Println(s)
		return
	}

	log.Println("mimipay reponse:\n", string(b))
	r = &MimiPayMakeOrderResponse{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		s := string(b)
		log.Println(s)
		if strings.Contains(err.Error(), `invalid character '<' looking for beginning of value`) {
			return
		}
		return
	}
	if !r.Success {
		err = fmt.Errorf("mimi pay return failed, %v", r)
		return
	}
	return
}

func round(x float64) int {
	return int(math.Floor(x + 0.5))
}
