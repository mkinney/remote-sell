package controllers

import (
	"github.com/revel/revel"
	qrcode "github.com/skip2/go-qrcode"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/url"
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

func getCryptoAmount(batm_url string, serial_number string, crypto_currency string, fiat_amount float64) float64 {
  v := url.Values{}
  v.Set("serial_number", serial_number)
  v.Add("fiat_currency", "USD")
  v.Add("crypto_currency", crypto_currency)
  v.Add("fiat_amount", fmt.Sprintf("%f", fiat_amount))

  full_url := batm_url + "/calculate_crypto_amount" + "?" + v.Encode()
  revel.AppLog.Info(full_url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
  response, err := client.Get(full_url)
  if err != nil {
		revel.AppLog.Error("Error in Get on getCryptoAmount", err)
		return 0
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
		revel.AppLog.Error("Did not get a 200 from getCryptoAmount", err)
		return 0
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
		revel.AppLog.Error("Error in response to getCryptoAmount", err)
		return 0
  }
	revel.AppLog.Debug(string(responseData))

  m := map[string]CalculateCryptoAmountResponsePartial{}
  err = json.Unmarshal(responseData, &m)
  if err != nil {
		revel.AppLog.Error("Error in ummarshal in getCryptoAmount", err)
		return 0
  }
  return m[crypto_currency].CryptoAmount
}

func sellCrypto(batm_url string, serial_number string, crypto_currency string, fiat_amount float64, crypto_amount float64) SellResponse {
  v := url.Values{}
  v.Set("serial_number", serial_number)
  v.Add("fiat_currency", "USD")
  v.Add("crypto_currency", crypto_currency)
  v.Add("fiat_amount", fmt.Sprintf("%f", fiat_amount))
  v.Add("crypto_amount", fmt.Sprintf("%.f", crypto_amount))

  full_url := batm_url + "/sell_crypto" + "?" + v.Encode()
  revel.AppLog.Debug(full_url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
  response, err := client.Get(full_url)
  if err != nil {
		revel.AppLog.Error("Error in sellCrypto on Get", err)
		return SellResponse{}
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
		revel.AppLog.Error("response from sellCrypto was not 200", err)
		return SellResponse{}
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
		revel.AppLog.Error("Could not read response in sellCrypto", err)
		return SellResponse{}
  }
	revel.AppLog.Debug("reponse from sellCrypto", responseData)

  sr := SellResponse{}
  err = json.Unmarshal(responseData, &sr)
  if err != nil {
		revel.AppLog.Error("Error in unmarshal of sellCrypto", err)
		return SellResponse{}
  }
  return sr
}

// the start of a BTC barcode is "bitcoin:"
// don't know what the other types are
func CryptoToPrefix(crypto string) string {
	retval := ""
	switch crypto {
	case "BTC":
		retval = "bitcoin:"
	case "LTC":
		retval = "litecoin:"
	}
	return retval
}

func LocationToSerialNumber(location string) string {
	retval := ""
	revel.AppLog.Debug("in LocationToSerialNumber... location:", location)
	switch location {
	case "1": // Clackamas Town Center Mall
		retval = "BT300795"
	case "2": // Clackamas Test
		retval = "BT102781" // test
	}
	revel.AppLog.Debug("leaving LocationToSerialNumber... retval:", retval)
	return retval
}

func (c App) Index() revel.Result {
	greeting := "Remote Sell"
	return c.Render(greeting)
}

// hidden_ fields should be blank incoming
func (c App) RemoteSell(location string, crypto string, fiat float64, hidden_uuid string, hidden_crypto_amount  string, hidden_minutes string, hidden_address string) revel.Result {
	c.Validation.Required(location).Message("Location is a required field.")
	//revel.AppLog.Debug("%+v", c.Validation)

	c.Validation.Required(crypto).Message("Crypto is a required field.")
	c.Validation.MinSize(crypto, 3).Message("The value for crypto is not long enough.")

	c.Validation.RangeFloat(fiat, 20.0, 3000.0).Message("Can only sell between $20 and $3000.")

	// lookup the serial number from the location
	serialNumber := LocationToSerialNumber(location)
	if serialNumber == "" {
		e := fmt.Sprintf("INTERNAL Invalid location:%s (code 19)", location)
		c.Validation.Error(e)
	}

	batm_url := revel.Config.StringDefault("batm_url", "")
	if batm_url == "" {
		c.Validation.Error("INTERNAL could not get a batm_url (code 20).")
	}
	revel.AppLog.Debug("BATM url:", batm_url)

	// do basic field validation before web service calls
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	crypto_amount := getCryptoAmount(batm_url, serialNumber, crypto, fiat)
	if crypto_amount < 0.000000001 {
		c.Validation.Error("INTERNAL could not get crypto_amount (code 21).")
	}
	revel.AppLog.Debug("crypto_amount:", crypto_amount)

	// I do not like the repeat of the validation error check, but it is best
	// to fail early before trying the sell_crypto call
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	sr := sellCrypto(batm_url, serialNumber, crypto, fiat, crypto_amount)
	if sr.ValidityInMinutes < 1 {
		c.Validation.Error("INTERNAL error processing sell crypto (code 22).")
	}
	revel.AppLog.Debug("sr:", sr)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	prefix := CryptoToPrefix(crypto)
	qrString := fmt.Sprintf("%s%s?amount=%f&label=%s&uuid=%s", prefix, sr.CryptoAddress, crypto_amount, sr.RemoteTransactionID, sr.TransactionUUID)
	revel.AppLog.Debug(qrString)

	// write the barcode to a temp file based on uuid
	hidden_uuid = sr.RemoteTransactionID
	hidden_crypto_amount = fmt.Sprintf("%f", crypto_amount)
	hidden_minutes = fmt.Sprintf("%d", sr.ValidityInMinutes)
	hidden_address = sr.CryptoAddress
	err := qrcode.WriteFile(qrString, qrcode.Medium, 256, "public/img/rs_" + hidden_uuid + ".png")
	if err != nil {
		revel.AppLog.Error("Could not write qrcode", err)
	}

	return c.Render(location, crypto, fiat, hidden_uuid, hidden_crypto_amount, hidden_minutes, hidden_address)
}
