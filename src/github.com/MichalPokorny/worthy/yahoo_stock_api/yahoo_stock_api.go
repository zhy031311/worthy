package yahoo_stock_api

import (
	"errors"
	"fmt"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/stock"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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

func GetTickers(symbols []string) ([]stock.Ticker, error) {
	tickers := make([]stock.Ticker, len(symbols))
	if len(symbols) == 0 {
		return tickers, nil
	}

	values := url.Values{}
	values.Add("s", strings.Join(symbols, "+"))
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
	if len(lines) != len(tickers) {
		return tickers, errors.New("Num of tickers requested != num of tickers received")
	}
	for i, line := range lines {
		if tickers[i], err = parseTicker(line); err != nil {
			return tickers, err
		}
	}
	return tickers, nil
}
