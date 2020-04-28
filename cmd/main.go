package main

import (
	"log"
	"os"

	"github.com/athom/mimipay"
)

func main() {
	userKey := os.Getenv("MIMIPAY_USER_KEY")
	secret := os.Getenv("MIMIPAY_SECRET")
	notifyURL := os.Getenv("MIMIPAY_NOITFY_URL")
	p := mimipay.NewMimiPay(userKey, secret, notifyURL)
	data, err := p.MakeOrderToMimiPay("mimipay demo", mimipay.PAY_TYPE_ALIPAY, "123", "456", 0.01, 300, "")
	if err != nil {
		panic(err)
	}
	log.Printf("qrcode_url: %v", data.QrCodeUrl)
}
