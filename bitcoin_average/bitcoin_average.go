// Simple wrapper around https://bitcoinaverage.com/api/.

package bitcoin_average

import (
	"encoding/json"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/util"
	"io/ioutil"
	"net/http"
	"time"
)

const cachePath = "~/dropbox/finance/bitcoin_average_cache.json"
const cacheDuration = 1 * time.Hour

type cacheItem struct {
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
}

var cache struct {
	Tickers map[string]*cacheItem `json:"tickers"`
}

func initCache() {
	if cache.Tickers == nil {
		if util.FileExists(cachePath) {
			util.LoadJSONFileOrDie(cachePath, &cache)
		} else {
			cache.Tickers = make(map[string]*cacheItem)
		}
	}
}

func writeCache() {
	bytes, err := json.Marshal(cache)
	if err != nil {
		panic(err)
	}
	util.WriteFile(cachePath, bytes)
}

func (item cacheItem) isTickerFresh() bool {
	takenAt, err := time.Parse(time.RFC3339, item.Timestamp)
	if err != nil {
		panic(err)
	}
	invalidAt := takenAt.Add(cacheDuration)
	isFresh := time.Now().Before(invalidAt)
	return isFresh
}

func (item cacheItem) isConversionFresh() bool {
	takenAt, err := time.Parse(time.RFC3339, item.Timestamp)
	if err != nil {
		panic(err)
	}
	invalidAt := takenAt.Add(cacheDuration)
	return time.Now().Before(invalidAt)
}

const endpoint = "https://api.bitcoinaverage.com/ticker/global/"

func getConversion(from string, to string) float64 {
	initCache()

	if from != "BTC" {
		panic("Can only convert from BTC.")
	}

	if cachedTicker, cacheHit := cache.Tickers[to]; cacheHit {
		if cachedTicker.isConversionFresh() {
			return cachedTicker.Price
		}
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
	cache.Tickers[to] = &cacheItem{
		Price:     conversion,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	writeCache()
	return conversion
}

func Convert(from money.Money, to string) money.Money {
	factor := getConversion(from.Currency, to)
	return money.New(to, from.Amount*factor)
}

func CanConvert(from money.Money, to string) bool {
	return from.Currency == "BTC"
}
