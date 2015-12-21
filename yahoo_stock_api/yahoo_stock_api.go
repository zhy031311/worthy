package yahoo_stock_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/stock"
	"github.com/MichalPokorny/worthy/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const cachePath = "~/dropbox/finance/yahoo_stock_api_cache.json"
const cacheDuration = 10 * time.Minute

type cacheItem struct {
	Ticker    stock.Ticker `json:"ticker"`
	Timestamp string       `json:"timestamp"`
}

type cacheType struct {
	Tickers map[string]*cacheItem `json:"tickers"`
}

var cache cacheType

func Init() {
	if util.FileExists(cachePath) {
		util.LoadJSONFileOrDie(cachePath, &cache)
	} else {
		cache.Tickers = make(map[string]*cacheItem)
	}
}

func writeCache() {
	// showCache()
	// fmt.Println("writing cache")

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

func parseTicker(yahooLine string) (stock.Ticker, error) {
	var ticker stock.Ticker
	var err error
	parts := strings.Split(strings.TrimSpace(yahooLine), ",")
	// The symbol is quoted in the CSV.
	ticker.Symbol = strings.Replace(parts[0], "\"", "", 2)
	var sellUSD float64
	sellUSD, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return ticker, fmt.Errorf("can't parse sell for ticker %s: %r", ticker.Symbol, parts[1])
	}
	ticker.Sell = money.New("USD", sellUSD)
	return ticker, nil
}

// Pretty much everything but previous close is totally useless (happily
// gives 0.0, N/A, 1.0, etc.).
const giveSymbol = "s"
const givePreviousClose = "p"

const endpoint = "http://download.finance.yahoo.com/d/quotes.csv"

func (item cacheItem) isConversionFresh() bool {
	takenAt, err := time.Parse(time.RFC3339, item.Timestamp)
	if err != nil {
		panic(err)
	}
	invalidAt := takenAt.Add(cacheDuration)
	return time.Now().Before(invalidAt)
}

func GetTickers(symbols []string) ([]stock.Ticker, error) {
	tickers := make([]stock.Ticker, len(symbols))
	missedIndices := make([]int, 0)

	for i, symbol := range symbols {
		needNew := true
		if cachedTicker, cacheHit := cache.Tickers[symbol]; cacheHit {
			if cachedTicker.isConversionFresh() {
				// fmt.Println(symbol + ": cache hit")
				tickers[i] = cachedTicker.Ticker
				needNew = false
			}
		}
		if needNew {
			// fmt.Println(symbol + ": cache miss")
			missedIndices = append(missedIndices, i)
		}
	}

	if len(missedIndices) == 0 {
		return tickers, nil
	}

	missedSymbols := make([]string, 0)
	for _, i := range missedIndices {
		missedSymbols = append(missedSymbols, symbols[i])
	}

	values := url.Values{}
	values.Add("s", strings.Join(missedSymbols, "+"))
	values.Add("f", giveSymbol+givePreviousClose)

	requestUrl := endpoint + "?" + values.Encode()

	resp, err := http.Get(requestUrl)
	if err != nil {
		return tickers, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tickers, err
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != len(missedSymbols) {
		return tickers, errors.New("Num of missed symbols != num of tickers received")
	}
	for _, line := range lines {
		var ticker stock.Ticker
		if ticker, err = parseTicker(line); err != nil {
			return tickers, err
		}
		cache.Tickers[ticker.Symbol] = &cacheItem{
			Ticker:    ticker,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		for _, i := range missedIndices {
			if symbols[i] == ticker.Symbol {
				tickers[i] = ticker
			}
		}
	}
	writeCache()
	return tickers, nil
}
