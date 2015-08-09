// Simple wrapper around https://bitcoinaverage.com/api/.

package bitcoin_average

import (
	"encoding/json"
	"github.com/MichalPokorny/worthy/money"
	"io/ioutil"
	"net/http"
)

const endpoint = "https://api.bitcoinaverage.com/ticker/global/"

func getConversion(from string, to string) float64 {
	if from != "BTC" {
		panic("Can only convert from BTC.")
	}
	requestUrl := endpoint + "/" + to
	resp, err := http.Get(requestUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		panic(err)
	}
	var conversion float64
	if jsonBody["24h_avg"] != nil {
		conversion = jsonBody["24h_avg"].(float64)
	} else {
		conversion = jsonBody["last"].(float64)
	}
	return conversion
}

func Convert(from money.Money, to string) money.Money {
	factor := getConversion(from.Currency, to)
	return money.New(to, from.Amount*factor)
}
