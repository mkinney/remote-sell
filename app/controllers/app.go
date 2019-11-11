package controllers

import (
	"github.com/revel/revel"
	qrcode "github.com/skip2/go-qrcode"
	"fmt"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	greeting := "Remote Sell"
	return c.Render(greeting)
}

func (c App) RemoteSell(location string, crypto string, fiat float64) revel.Result {
	c.Validation.Required(location).Message("Location is a required field.")

	c.Validation.Required(crypto).Message("Crypto is a required field.")
	c.Validation.MinSize(crypto, 3).Message("The value for crypto is not long enough.")

	c.Validation.RangeFloat(fiat, 20.0, 3000.0).Message("Can only sell between $20 and $3000.")


	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	// TODO: simulate calculate_crypto_amount
	// TODO: simulate sell_crypto

	// TODO: locationToSerialNumber
	// TODO: call calculate_crypto_amount web service
	// TODO: call sell_crypto with all values
	// TODO: create temp file for qr code
	// TODO: when/how to clean up old qr code tmp files?
	// TODO: do we care that anyone can see the qr.png file(s)?
	err := qrcode.WriteFile("https://example.org", qrcode.Medium, 256, "public/img/qr.png")
	if err != nil {
		fmt.Println(err)
	}

	return c.Render(location, crypto, fiat)
}
