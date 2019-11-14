/*
    Copyright (C) 2019 Mike Kinney
		See LICENSE.txt
 */

package controllers

import (
	"github.com/revel/revel"
	qrcode "github.com/skip2/go-qrcode"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/url"
	"math"
	"strconv"
	"time"
	"net/http"
	"crypto/tls"
)

type App struct {
	*revel.Controller
}

type CalculateCryptoAmountResponsePartial struct {
  CryptoAmount        float64
  // Note: Ignoring rest of fields
}

type SellResponse struct {
  CashAmount     float64 `json:"cashAmount"`
  CashCurrency   string  `json:"cashCurrency"`
  CryptoAddress  string  `json:"cryptoAddress"`
  CryptoAmount   float64 `json:"cryptoAmount"`
  CryptoCurrency string  `json:"cryptoCurrency"`
  CustomData     struct {} `json:"customData"`
  FixedTransactionFee float64 `json:"fixedTransactionFee"`
  LocalTransactionID  string  `json:"localTransactionId"`
  RemoteTransactionID string  `json:"remoteTransactionId"`
  Status              int     `json:"status"`
  TransactionUUID     string  `json:"transactionUUID"`
  ValidityInMinutes   int     `json:"validityInMinutes"`
}

// rounding from: https://socketloop.com/tutorials/golang-round-float-to-precision-example
func (c App) Round(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

func (c App) RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

func (c App) RoundDown(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	newVal = round / pow
	return
}

func (c App) getCryptoAmount(batm_url string, serial_number string, crypto_currency string, fiat_amount float64) float64 {
  v := url.Values{}
  v.Set("serial_number", serial_number)
  v.Add("fiat_currency", "USD")
  v.Add("crypto_currency", crypto_currency)
  v.Add("fiat_amount", fmt.Sprintf("%f", fiat_amount))

  full_url := batm_url + "/calculate_crypto_amount" + "?" + v.Encode()
  c.Log.Info("in getCryptoAmount", "full_url", full_url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
  response, err := client.Get(full_url)
  if err != nil {
		c.Log.Error("Error in Get on getCryptoAmount", "error", err)
		return 0
  }
  defer response.Body.Close()

	// It looks like it returns 200 even when there is an error.
  if response.StatusCode != 200 {
		c.Log.Error("Did not get a 200 from getCryptoAmount", "error", err)
		return 0
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
		c.Log.Error("Error in response to getCryptoAmount", "error", err)
		return 0
  }
	c.Log.Debug("In getCryptoAmount", "responseData:", string(responseData))

	// Looks like this is how to detect an error.
	if string(responseData) == "ERROR" {
		c.Log.Error("Got ERROR when calling the calculate_crypto_amount service. Double check the BATM configuration for serial_number.", "full_url", full_url)
		return 0
  }

  m := map[string]CalculateCryptoAmountResponsePartial{}
  err = json.Unmarshal(responseData, &m)
  if err != nil {
		c.Log.Error("Error in ummarshal in getCryptoAmount", "error", err)
		return 0
  }
	raw := m[crypto_currency].CryptoAmount
	tmp := fmt.Sprintf("%f", c.RoundUp(raw, 6))
	c.Log.Debug("Rounding", "raw amount:", raw, "tmp", tmp)
	rounded, err := strconv.ParseFloat(tmp, 6)
	if err != nil {
		c.Log.Error("Could not ParseFloat getCryptoAmount", "error:", err)
		return 0
	}
	c.Log.Debug("In getCryptoAmount", "rounded:", rounded)
  return rounded
}

func (c App) sellCrypto(batm_url string, serial_number string, crypto_currency string, fiat_amount float64, crypto_amount float64) SellResponse {
  v := url.Values{}
  v.Set("serial_number", serial_number)
  v.Add("fiat_currency", "USD")
  v.Add("crypto_currency", crypto_currency)
  v.Add("fiat_amount", fmt.Sprintf("%f", fiat_amount))
  v.Add("crypto_amount", fmt.Sprintf("%.f", crypto_amount))

  full_url := batm_url + "/sell_crypto" + "?" + v.Encode()
  c.Log.Debug("in sellCrypto", "full_url", full_url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
  response, err := client.Get(full_url)
  if err != nil {
		c.Log.Error("Error in sellCrypto on Get", "error", err)
		return SellResponse{}
  }
  defer response.Body.Close()

	// It looks like it returns 200 even when there is an error.
  if response.StatusCode != 200 {
		c.Log.Error("response from sellCrypto was not 200", "error", err)
		return SellResponse{}
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
		c.Log.Error("Could not read response in sellCrypto", "error", err)
		return SellResponse{}
  }
	c.Log.Debug("In sellCrypto", "reponse from sellCrypto", string(responseData))

	// Looks like this is how to detect an error.
	if string(responseData) == "ERROR" {
		c.Log.Error("Got ERROR when calling the sell_crypto service. Double check the BATM configuration for serial_number", "full_url", full_url)
		return SellResponse{}
  }

  sr := SellResponse{}
  err = json.Unmarshal(responseData, &sr)
  if err != nil {
		c.Log.Error("Error in unmarshal of sellCrypto", "error", err)
		return SellResponse{}
  }
  return sr
}

// the start of a barcode (for example: BTC is "bitcoin:")
func (c App) CryptoToPrefix(crypto string) string {
	retval := ""
	switch crypto {
	case "BTC":
		retval = "bitcoin:"
	case "LTC":
		retval = "litecoin:"
	}
	return retval
}

func (c App) LocationToSerialNumber(location string) string {
	retval := ""
	switch location {
	case "1": // Clackamas Town Center Mall
		retval = "BT300795"
	case "2": // Clackamas Test
		retval = "BT102781" // test
	}
	return retval
}

func (c App) Index() revel.Result {
	greeting := "Remote Sell"
	return c.Render(greeting)
}

// hidden_ fields should be blank incoming
func (c App) RemoteSell(location string, crypto string, fiat float64, hidden_uuid string, hidden_crypto_amount  string, hidden_minutes string, hidden_address string, hidden_now time.Time) revel.Result {
	c.Validation.Required(location).Message("Location is a required field.")
	//c.Log.Debug("%+v", c.Validation)

	c.Validation.Required(crypto).Message("Crypto is a required field.")
	c.Validation.MinSize(crypto, 3).Message("The value for crypto is not long enough.")

	c.Validation.RangeFloat(fiat, 20.0, 3000.0).Message("Can only sell between $20 and $3000.")

	// lookup the serial number from the location
	serialNumber := c.LocationToSerialNumber(location)
	c.Log.Debug("location", location, "serial number:", serialNumber)
	if serialNumber == "" {
		e := fmt.Sprintf("INTERNAL Invalid location:%s (code 19)", location)
		c.Validation.Error(e)
	}

	batm_url := revel.Config.StringDefault("batm_url", "")
	if batm_url == "" {
		c.Validation.Error("INTERNAL could not get a batm_url (code 20).")
	}
	c.Log.Debug("in RemoteSell", "BATM url:", batm_url)

	// do basic field validation before web service calls
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	crypto_amount := c.getCryptoAmount(batm_url, serialNumber, crypto, fiat)
	if crypto_amount < 0.000000001 {
		c.Validation.Error("INTERNAL could not get crypto_amount (code 21).")
	}
	c.Log.Debug("in RemoteSell", "crypto_amount:", crypto_amount)

	// I do not like the repeat of the validation error check, but it is best
	// to fail early before trying the sell_crypto call
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	sr := c.sellCrypto(batm_url, serialNumber, crypto, fiat, crypto_amount)
	if sr.ValidityInMinutes < 1 {
		c.Validation.Error("INTERNAL error processing sell crypto (code 22).")
	}
	c.Log.Debug("in RemoteSell", "sr:", sr)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	prefix := c.CryptoToPrefix(crypto)
	qrString := fmt.Sprintf("%s%s?amount=%f&label=%s&uuid=%s", prefix, sr.CryptoAddress, crypto_amount, sr.RemoteTransactionID, sr.TransactionUUID)
	c.Log.Debug("in RemoteSell", "qrString", qrString)

	// write the barcode to a temp file based on uuid
	hidden_uuid = sr.RemoteTransactionID
	hidden_crypto_amount = fmt.Sprintf("%f", crypto_amount)
	hidden_minutes = fmt.Sprintf("%d", sr.ValidityInMinutes)
	hidden_address = sr.CryptoAddress
	hidden_now = time.Now()
	err := qrcode.WriteFile(qrString, qrcode.Medium, 256, "public/img/rs_" + hidden_uuid + ".png")
	if err != nil {
		c.Log.Error("Could not write qrcode", "error", err)
	}
	// log all requests to Info
	c.Log.Info("request to remoteSell", "location", location, "crypto", crypto, "fiat", fiat, "hidden_crypto_amount", hidden_crypto_amount, "hidden_minutes", hidden_minutes, "hidden_address", hidden_address, "hidden_uuid", hidden_uuid, "now", hidden_now, "qrString", qrString)

	return c.Render(location, crypto, fiat, hidden_uuid, hidden_crypto_amount, hidden_minutes, hidden_address, hidden_now)
}
