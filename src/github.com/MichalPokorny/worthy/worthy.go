package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MichalPokorny/worthy/free_currency_converter"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/portfolio"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
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
		converted := free_currency_converter.Convert(item, target)
		if converted.Currency != target {
			panic("conversion fail")
		}
		total += converted.Amount
	}
	return money.New(target, total)
}

func StartsWith(s string, prefix string) bool {
	return s[0:len(prefix)] == prefix
}

func expand(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if StartsWith(path, "~/") {
		path = dir + "/" + path[2:]
	}
	return path
}

func LoadPortfolio() (portfolio.Portfolio, []money.Money) {
	path := expand("~/.stock-portfolio.json")
	body, err := ioutil.ReadFile(path)
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

func GetBrokerAccountValue() money.Money {
	myPortfolio, myCurrencies := LoadPortfolio()
	myCurrencies = append(myCurrencies, GetValue(myPortfolio))
	return sumMoney(myCurrencies, "CZK")
}

func GetBitcoinValue() money.Money {
	path := expand("~/.btckit/wallet_btc")
	body, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(string(body)), 64)
	if err != nil {
		panic(err)
	}
	bitcoins := money.New("BTC", amount)
	return free_currency_converter.Convert(bitcoins, "CZK")
}

func GetEuroAccountValue() money.Money {
	path := expand("~/.euro-account")
	body, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(string(body)), 64)
	if err != nil {
		panic(err)
	}
	bitcoins := money.New("EUR", amount)
	return free_currency_converter.Convert(bitcoins, "CZK")
}

func main() {
	var mode = flag.String("mode", "", "'broker', 'bitcoin', 'euro_account' or 'table'")
	flag.Parse()

	switch *mode {
	case "broker":
		brokerAccount := GetBrokerAccountValue()
		fmt.Printf("%.2f\n", brokerAccount.Amount)
	case "bitcoin":
		bitcoins := GetBitcoinValue()
		fmt.Printf("%.2f\n", bitcoins.Amount)
	case "euro_account":
		euros := GetEuroAccountValue()
		fmt.Printf("%.2f\n", euros.Amount)
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Append([]string{"Bitcoiny", GetBitcoinValue().String()})
		table.Append([]string{"Akcie", GetBrokerAccountValue().String()})
		table.Append([]string{"EUR účet", GetEuroAccountValue().String()})
		table.Render()
	default:
		panic("bad mode")
	}
}
