// Simple wrapper around http://www.currencylayer.com/.

package currency_layer

import (
	"encoding/json"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const endpoint = "http://apilayer.net/api/live"

var cache map[string]float64 = make(map[string]float64)
var accessKey string

func Init() {
	body := util.ReadFile("~/Dropbox/finance/currency_layer_key")
	accessKey = strings.Trim(body, "\n\r ")
}

func getConversion(from string, to string) float64 {
	if from == to {
		return 1
	}
	cacheKey := from + "_" + to
	if cachedResult, cacheHit := cache[cacheKey]; cacheHit {
		return cachedResult
	}

	values := url.Values{}
	values.Add("access_key", accessKey)
	values.Add("currencies", from + "," + to)

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
	conversion := jsonBody["quotes"].(map[string]interface{})[from + to].(float64)
	cache[cacheKey] = conversion
	return conversion
}

func Convert(from money.Money, to string) money.Money {
	factor := getConversion(from.Currency, to)
	return money.New(to, from.Amount*factor)
}
