package yahoo_stock_api

import (
	"errors"
	//"github.com/davecgh/go-spew/spew"
	"github.com/MichalPokorny/worthy/money"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Ticker struct {
	Symbol string
	Sell   money.Money
}

func parseTicker(yahooLine string) (Ticker, error) {
	var ticker Ticker
	var err error
	parts := strings.Split(strings.TrimSpace(yahooLine), ",")
	//spew.Dump(parts)
	// The symbol is quoted in the CSV.
	ticker.Symbol = strings.Replace(parts[0], "\"", "", 2)
	// First try to get a realtime bid price.
	var sellUSD float64
	sellUSD, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return ticker, err
	}
	ticker.Sell.Currency = "USD"
	ticker.Sell.Amount = sellUSD

	return ticker, nil
}

// Pretty much everything but previous close is totally useless.
const giveSymbol = "s0"
const givePreviousClose = "p0"

const endpoint = "http://download.finance.yahoo.com/d/quotes.csv"

func GetTickers(symbols []string) ([]Ticker, error) {
	tickers := make([]Ticker, len(symbols))
	values := url.Values{}
	values.Add("s", strings.Join(symbols, ","))
	values.Add("f", giveSymbol+givePreviousClose)
	values.Add("e", ".csv")

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
