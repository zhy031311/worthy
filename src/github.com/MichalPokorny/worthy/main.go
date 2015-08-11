package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MichalPokorny/worthy/currency_layer"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/portfolio"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"github.com/MichalPokorny/worthy/bitcoin_average"
	"github.com/MichalPokorny/worthy/util"
	"github.com/olekukonko/tablewriter"
	"os"
)

var STOCK_PORTFOLIO_PATH = "~/Dropbox/finance/stock-portfolio.json"
var CHASE_ACCOUNT_PATH = "~/Dropbox/finance/chase_account"
var BITCOIN_PATH = "~/Dropbox/finance/wallet_btc"

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
		converted := currency_layer.Convert(item, target)
		if converted.Currency != target {
			panic("conversion fail")
		}
		total += converted.Amount
	}
	return money.New(target, total)
}

func LoadPortfolio() (portfolio.Portfolio, []money.Money) {
	body := util.ReadFileBytes(STOCK_PORTFOLIO_PATH)
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

func GetBrokerAccountValue() money.Money {
	myPortfolio, myCurrencies := LoadPortfolio()
	myCurrencies = append(myCurrencies, GetValue(myPortfolio))
	return sumMoney(myCurrencies, "CZK")
}

func GetChaseAccountValue() money.Money {
	amount := util.ReadFileFloat64(CHASE_ACCOUNT_PATH)
	dollars := money.New("USD", amount)
	return currency_layer.Convert(dollars, "CZK")
}

func GetBitcoinValue() money.Money {
	amount := util.ReadFileFloat64(BITCOIN_PATH)
	bitcoins := money.New("BTC", amount)
	return bitcoin_average.Convert(bitcoins, "CZK")
}

func main() {
	var mode = flag.String("mode", "", "'broker', 'broker', 'bitcoin' or 'table'")
	flag.Parse()

	currency_layer.Init()

	switch *mode {
	case "broker":
		brokerAccount := GetBrokerAccountValue()
		fmt.Printf("%.2f\n", brokerAccount.Amount)
	case "bitcoin":
		bitcoins := GetBitcoinValue()
		fmt.Printf("%.2f\n", bitcoins.Amount)
	case "chase":
		chase := GetChaseAccountValue()
		fmt.Printf("%.2f\n", chase.Amount)
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Append([]string{"Bitcoiny", GetBitcoinValue().String()})
		table.Append([]string{"Akcie", GetBrokerAccountValue().String()})
		table.Append([]string{"Chase účet", GetChaseAccountValue().String()})
		table.Render()
	default:
		panic("bad mode")
	}
}
