package free_currency_converter

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"github.com/MichalPokorny/worthy/money"
)

const endpoint = "http://www.freecurrencyconverterapi.com/api/v3/convert"

var cache map[string]float64 = make(map[string]float64)

func getConversion(from string, to string) float64 {
	query := from+"_"+to

	if cachedResult, cacheHit := cache[query]; cacheHit {
		return cachedResult
	}

	values := url.Values{}
	values.Add("q", query)
	values.Add("compact", "y")

	requestUrl := endpoint + "?" + values.Encode()
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
	conversion := jsonBody[query].(map[string]interface{})["val"].(float64)
	cache[query] = conversion
	return conversion
}

func Convert(from money.Money, to string) money.Money {
	factor := getConversion(from.Currency, to)
	return money.New(to, from.Amount * factor)
}
