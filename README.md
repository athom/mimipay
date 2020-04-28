# MimiPay Go SDK

### Install 

```bash
go get github.com/athom/mimipay
```

### Usage

```go
p := mimipay.NewMimiPay(userKey, secret, notifyURL, returnURL)
data, err := p.MakeOrderToMimiPay("mimipay demo", mimipay.PAY_TYPE_ALIPAY, "123", "456", "0.01")
qrcodeURL := data.QrCodeUrl
```

Then use the `qrcodeURL` in your app.

Happy mimipaying!

## License

MimiPay is released under the [WTFPL License](http://www.wtfpl.net/txt/copying).

