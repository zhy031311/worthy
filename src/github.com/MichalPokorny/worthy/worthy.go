package main

import (
	"encoding/json"
	"fmt"
	"github.com/MichalPokorny/worthy/free_currency_converter"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/portfolio"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"io/ioutil"
	"os/user"
)

func GetValue(portfolio portfolio.Portfolio) money.Money {
	symbols := portfolio.GetStockSymbols()
	stockValues := map[string]float64{}
	tickers, err := yahoo_stock_api.GetTickers(symbols)
	if err != nil {
		panic(err)
	}
	for _, ticker := range tickers {
		if ticker.Sell.Currency != "USD" {
			panic("stock not selling in USD")
		}
		// TODO: assert no duplicates
		stockValues[ticker.Symbol] = ticker.Sell.Amount
	}
	total := 0.0
	for _, entry := range portfolio {
		total += float64(entry.Amount) * stockValues[entry.Ticker]
	}
	result := money.Money{}
	result.Currency = "USD"
	result.Amount = total
	return result
}

func sumMoney(inputs []money.Money, target string) money.Money {
	total := 0.0
	for _, item := range inputs {
		converted := free_currency_converter.Convert(item, target)
		if converted.Currency != target {
			panic("conversion fail")
		}
		total += converted.Amount
	}
	return money.New(target, total)
}

func LoadPortfolio() (portfolio.Portfolio, []money.Money) {
	usr, _ := user.Current()
	body, err := ioutil.ReadFile(usr.HomeDir + "/.stock-portfolio.json")
	if err != nil {
		panic(err)
	}
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		panic(err)
	}
	stocks := jsonBody["stocks"].(map[string]interface{})
	currencies := jsonBody["currencies"].(map[string]interface{})

	outPortfolio := make(portfolio.Portfolio, 0)
	for ticker, amount := range stocks {
		outPortfolio = append(outPortfolio, portfolio.NewEntry(ticker, int(amount.(float64))))
	}
	outCurrencies := make([]money.Money, 0)
	for ticker, amount := range currencies {
		outCurrencies = append(outCurrencies, money.New(ticker, amount.(float64)))
	}
	return outPortfolio, outCurrencies
}

func main() {
	myPortfolio, myCurrencies := LoadPortfolio()
	myCurrencies = append(myCurrencies, GetValue(myPortfolio))

	total := sumMoney(myCurrencies, "CZK")
	fmt.Printf("%.2f\n", total.Amount)
}
