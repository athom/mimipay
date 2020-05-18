package mimipay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// example
// {
//		"outUserNo":"123",
//		"price":"2.99",
//		"outTradeNo":"xxxx",
//		"key":"4f113ebb8bc18351dc345a26be6c2xxxx",
//		"realPrice":"2.99"
// }
type mimiPayNotification struct {
	OutTradeNo string  `json:"outTradeNo"`
	OutUserNo  string  `json:"outUserNo"`
	Price      float64 `json:"price"`
	RealPrice  float64 `json:"realPrice"`
	Sign       string  `json:"key"`
}
type mimiPayNotificationFallback struct {
	OutTradeNo string `json:"outTradeNo"`
	OutUserNo  string `json:"outUserNo"`
	Price      string `json:"price"`
	RealPrice  string `json:"realPrice"`
	Sign       string `json:"key"`
}

type MimiPayNotifyResult struct {
	OutTradeNo      string
	OutUserNo       string
	PriceString     string
	RealPriceString string
	PriceFloat      float64
	RealPriceFloat  float64
	Sign            string
}

// https://mimipay.cc/document/pay
// outTradeNo + round(price*100) + round(realPrice*100) + secret
func (this *MimiPay) makeResultSign(price float64, realPrice float64, outTradeNo string) (r string) {
	iPrice := round(price * 100)
	ssPrice := fmt.Sprintf("%v", iPrice)

	iRealPrice := round(realPrice * 100)
	ssRealPrice := fmt.Sprintf("%v", iRealPrice)

	ss := outTradeNo + ssPrice + ssRealPrice + this.Secret
	r = MD5WithLowerCase(ss)
	return
}

func (this *MimiPay) ExtractNotifyData(b []byte) (r *MimiPayNotifyResult, err error) {
	r = &MimiPayNotifyResult{}
	ntf := &mimiPayNotification{}
	err = json.Unmarshal(b, ntf)
	if err != nil {
		ntf2 := &mimiPayNotificationFallback{}
		err = json.Unmarshal(b, ntf2)
		if err != nil {
			return
		}

		r.OutTradeNo = ntf2.OutTradeNo
		r.OutUserNo = ntf2.OutUserNo
		r.PriceString = ntf2.Price
		r.RealPriceString = ntf2.RealPrice
		r.Sign = ntf2.Sign
		r.PriceFloat, err = strconv.ParseFloat(r.PriceString, 64)
		if err != nil {
			return
		}
		r.RealPriceFloat, err = strconv.ParseFloat(r.RealPriceString, 64)
		if err != nil {
			return
		}
	} else {
		r.OutTradeNo = ntf.OutTradeNo
		r.OutUserNo = ntf.OutUserNo
		r.PriceFloat = ntf.Price
		r.RealPriceFloat = ntf.RealPrice
		r.PriceString = fmt.Sprintf("%v", ntf.Price)
		r.RealPriceString = fmt.Sprintf("%v", ntf.RealPrice)
		r.Sign = ntf.Sign
	}

	sign := this.makeResultSign(r.PriceFloat, r.RealPriceFloat, r.OutTradeNo)
	if sign != r.Sign {
		err = fmt.Errorf("notify result get invalid sign: %v", string(b))
		return
	}

	return
}

func (this *MimiPay) GinNotifyData(c *gin.Context) (r *MimiPayNotifyResult, err error) {
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	log.Printf("receive mimipay notify:\n%v", string(b))
	return this.ExtractNotifyData(b)
}
