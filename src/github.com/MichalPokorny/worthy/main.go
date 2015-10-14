package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MichalPokorny/worthy/bitcoin_average"
	"github.com/MichalPokorny/worthy/currency_layer"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/portfolio"
	"github.com/MichalPokorny/worthy/util"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"github.com/olekukonko/tablewriter"
	"os"
)

func getValueOfStocks(portfolio portfolio.Portfolio) money.Money {
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
		total += float64(entry.Amount) * stockValues[entry.Symbol]
	}
	result := money.Money{}
	result.Currency = "USD"
	result.Amount = total
	return result
}

func sumMoney(inputs []money.Money, target string) money.Money {
	total := 0.0
	for _, item := range inputs {
		var converted money.Money = item

		// TODO: multi-step conversion
		if bitcoin_average.CanConvert(converted, target) {
			converted = bitcoin_average.Convert(converted, target)
		} else if currency_layer.CanConvert(converted, target) {
			converted = currency_layer.Convert(converted, target)
		}
		if converted.Currency != target {
			panic("conversion fail")
		}
		total += converted.Amount
	}
	return money.New(target, total)
}

func loadPortfolio(path string) (portfolio.Portfolio, []money.Money) {
	body := util.ReadFileBytes(path)
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		panic(err)
	}

	stocks := make(map[string]interface{})
	if stocksField, ok := jsonBody["stocks"]; ok {
		stocks = stocksField.(map[string]interface{})
	}

	currencies := make(map[string]interface{})
	if currenciesField, ok := jsonBody["currencies"]; ok {
		currencies = currenciesField.(map[string]interface{})
	}

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

func getAccountValue(account money.AccountEntry) money.Money {
	stocks, currencies := loadPortfolio(account.Path)
	currencies = append(currencies, getValueOfStocks(stocks))
	return sumMoney(currencies, "CZK")
}

func main() {
	var mode = flag.String("mode", "", "'broker', 'broker', 'bitcoin' or 'table'")
	flag.Parse()

	currency_layer.Init()
	yahoo_stock_api.Init()

	var accounts []money.AccountEntry = money.LoadAccounts()

	for _, account := range accounts {
		if account.Code == *mode {
			value := getAccountValue(account)
			fmt.Printf("%.2f\n", value.Amount)
			return
		}
	}

	if *mode == "table" {
		table := tablewriter.NewWriter(os.Stdout)
		for _, account := range accounts {
			table.Append([]string{account.Name, getAccountValue(account).String()})
		}
		table.Render()
		return
	}

	panic("bad mode")
}
