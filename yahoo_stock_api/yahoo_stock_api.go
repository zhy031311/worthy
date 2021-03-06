package yahoo_stock_api

import (
	"os"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-yql"
	"database/sql"
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
const cacheDuration = 1 * time.Hour

type cacheItem struct {
	Ticker    stock.Ticker `json:"ticker"`
	Timestamp string       `json:"timestamp"`
}

var cache struct {
	Tickers map[string]*cacheItem `json:"tickers"`
}

var stockHistoryCachePath string = "/home/prvak/dropbox/finance/stock_history"

func SetHistoryCachePath(path string) {
	stockHistoryCachePath = path
}

func init() {
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
				tickers[i] = cachedTicker.Ticker
				needNew = false
			}
		}
		if needNew {
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

func toFloat(str string) float64 {
	x, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return x
}

type priceCache struct {
	Days []stock.TradingDay
}

func GetHistoricalPrices(symbol string, startDate string, endDate string) []stock.TradingDay {
	key := fmt.Sprintf("%s_%s_%s", symbol, startDate, endDate)
	path := stockHistoryCachePath + "/" + key + ".json"
	if _, err := os.Stat(path); err == nil {
		var cachedResult priceCache
		util.LoadJSONFileOrDie(path, &cachedResult)
		return cachedResult.Days
	}

	db, _ := sql.Open("yql", "||store://datatables.org/alltableswithkeys")
	defer db.Close()
	stmt, err := db.Query(
		"select * from yahoo.finance.historicaldata where symbol=? and startDate=? and endDate=?",
		symbol, startDate, endDate)
	if err != nil {
		panic(err)
	}
	days := make([]stock.TradingDay, 0)
	for stmt.Next() {
		var data map[string]interface{}
		stmt.Scan(&data)

		day := stock.TradingDay{
			Symbol: data["Symbol"].(string),
			Date: data["Date"].(string),
			Open: toFloat(data["Open"].(string)),
			High: toFloat(data["High"].(string)),
			Low: toFloat(data["Low"].(string)),
			Close: toFloat(data["Close"].(string)),
			AdjustedClose: toFloat(data["Adj_Close"].(string)),
		}
		days = append(days, day)
	}

	// Reverse the days (earliest first)
	for i := 0; i < len(days) / 2; i++ {
		days[i], days[len(days) - 1 - i] = days[len(days) - 1 - i], days[i]
	}

	cachedResult := priceCache{Days: days}
	util.WriteJSONFileOrDie(path, cachedResult)
	return days
}
