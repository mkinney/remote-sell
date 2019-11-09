package controllers

import (
	"github.com/revel/revel"
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

    return c.Render(location, crypto, fiat)
}
