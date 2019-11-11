package controllers

import (
	"github.com/revel/revel"
	qrcode "github.com/skip2/go-qrcode"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/url"
	"net/http"
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
  //fmt.Println(full_url)
  response, err := http.Get(full_url)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
		return 0
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
    //log.Fatal("Did not get a 200", response.Status)
		fmt.Println(err)
		return 0
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
		return 0
  }
  //fmt.Println(string(responseData))

  m := map[string]CalculateCryptoAmountResponsePartial{}
  err = json.Unmarshal(responseData, &m)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
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
  //fmt.Println(full_url)
  response, err := http.Get(full_url)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
		return SellResponse{}
  }
  defer response.Body.Close()

  if response.StatusCode != 200 {
    //log.Fatal("Did not get a 200", response.Status)
		fmt.Println(err)
		return SellResponse{}
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
		return SellResponse{}
  }
  //fmt.Println(string(responseData))

  sr := SellResponse{}
  err = json.Unmarshal(responseData, &sr)
  if err != nil {
    //log.Fatal(err)
		fmt.Println(err)
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
	fmt.Println("in LocationToSerialNumber... location:", location)
	switch location {
	case "1": // Clackamas Town Center Mall
		retval = "BT300795"
	}
	fmt.Println("leaving LocationToSerialNumber... retval:", retval)
	return retval
}

func (c App) Index() revel.Result {
	greeting := "Remote Sell"
	return c.Render(greeting)
}

// hidden_ fields should be blank incoming
func (c App) RemoteSell(location string, crypto string, fiat float64, hidden_uuid string, hidden_crypto_amount  string, hidden_minutes string, hidden_address string) revel.Result {
	c.Validation.Required(location).Message("Location is a required field.")
	fmt.Printf("%+v", c.Validation)

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
	fmt.Println("BATM url:", batm_url)

	crypto_amount := getCryptoAmount(batm_url, serialNumber, crypto, fiat)
	if crypto_amount < 0.000000001 {
		c.Validation.Error("INTERNAL could not get crypto_amount (code 21).")
	}
	fmt.Println("crypto_amount:", crypto_amount)

	sr := sellCrypto(batm_url, serialNumber, crypto, fiat, crypto_amount)
	if sr.ValidityInMinutes < 1 {
		c.Validation.Error("INTERNAL error processing sell crypto (code 22.")
	}
	fmt.Println("sr:", sr)

	// TODO: when/how to clean up old qr code tmp files?

	// TODO: prob have to deal with insecure https

	prefix := CryptoToPrefix(crypto)
	qrString := fmt.Sprintf("%s%s?amount=%f&label=%s&uuid=%s", prefix, sr.CryptoAddress, crypto_amount, sr.RemoteTransactionID, sr.TransactionUUID)
	fmt.Println(qrString)

	// write the barcode to a temp file based on uuid
	hidden_uuid = sr.RemoteTransactionID
	hidden_crypto_amount = fmt.Sprintf("%f", crypto_amount)
	hidden_minutes = fmt.Sprintf("%d", sr.ValidityInMinutes)
	hidden_address = sr.CryptoAddress
	err := qrcode.WriteFile(qrString, qrcode.Medium, 256, "public/img/rs_" + hidden_uuid + ".png")
	if err != nil {
		fmt.Println(err)
	}

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	return c.Render(location, crypto, fiat, hidden_uuid, hidden_crypto_amount, hidden_minutes, hidden_address)
}
