package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/MichalPokorny/worthy/bitcoin_average"
	"github.com/MichalPokorny/worthy/currency_layer"
	"github.com/MichalPokorny/worthy/homebank"
	"github.com/MichalPokorny/worthy/iks"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/portfolio"
	"github.com/MichalPokorny/worthy/util"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"time"
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
	return money.Money{Currency: "USD", Amount: total}
}

func convert(input money.Money, target string) money.Money {
	// TODO: multi-step conversion
	if bitcoin_average.CanConvert(input, target) {
		return bitcoin_average.Convert(input, target)
	} else if currency_layer.CanConvert(input, target) {
		return currency_layer.Convert(input, target)
	} else {
		panic("cannot convert " + input.Currency + " to " + target)
	}
}

func sumMoney(inputs []money.Money, target string) money.Money {
	total := money.New(target, 0)
	for _, item := range inputs {
		converted := convert(item, target)
		if converted.Currency != target {
			panic("conversion fail")
		}
		total.Add(converted)
	}
	return total
}

func loadPortfolio(path string) (portfolio.Portfolio, []money.Money) {
	jsonBody := make(map[string]interface{})
	util.LoadJSONFileOrDie(path, &jsonBody)

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
	if account.Path != nil {
		stocks, currencies := loadPortfolio(*account.Path)
		currencies = append(currencies, getValueOfStocks(stocks))
		return sumMoney(currencies, "CZK")
	} else if account.Value != nil {
		return convert(*account.Value, "CZK")
	} else if account.HomebankPath != nil {
		result, err := homebank.ParseHomebankFile(util.ExpandPath(*account.HomebankPath))
		if err != nil {
			panic(err)
		}
		total := float64(0)
		for _, account := range result.Accounts {
			if account.Flags&homebank.FLAG_CLOSED == homebank.FLAG_CLOSED {
				// Skip closed accounts.
				continue
			}
			if account.Type == homebank.ASSETS_ACCOUNT {
				// We manage assets accounts ourselves.
				continue
			}
			total += result.GetAccountBalance(account.Key)
		}
		return money.New("CZK", total)
	} else if account.IksPath != nil {
		iks.ScrapePrices() // TODO: lazy
		investment := iks.ParseInvestment(*account.IksPath)
		return iks.GetInvestmentValue(investment)
	} else {
		panic("Account has no Path, Value, HomebankPath or IksPath")
	}
}

func getTotalNetWorth(accounts []money.AccountEntry) money.Money {
	total := money.New("CZK", 0)
	for _, account := range accounts {
		total.Add(getAccountValue(account))
	}
	return total
}

func logNetWorth(accountsFile *money.AccountsFileData) {
	// TODO: move to something standardized. RFC3339?

	values := make([]string, len(accountsFile.CsvOrder))
	for i, fieldName := range accountsFile.CsvOrder {
		var value string

		switch fieldName {
		case "_datetime":
			timeLayout := "2006-01-02 15:04:05"
			value = time.Now().Format(timeLayout)

		case "_timestamp":
			value = strconv.FormatInt(time.Now().Unix(), 10)
			break

		case "_sum":
			sumValue := getTotalNetWorth(accountsFile.Accounts)
			value = strconv.FormatFloat(sumValue.Amount, 'f', 2, 64)
			break

		default:
			matchingAccount := findAccount(accountsFile.Accounts, fieldName)
			if matchingAccount == nil {
				panic("no such account: " + fieldName)
			}
			accountValue := getAccountValue(*matchingAccount)
			value = strconv.FormatFloat(accountValue.Amount, 'f', 2, 64)
			break
		}

		values[i] = value
	}

	csvPath := util.ExpandPath(accountsFile.CsvPath)
	file, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	w := csv.NewWriter(file)
	if err := w.Write(values); err != nil {
		panic(err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}

	file.Close()
}

func findAccount(accounts []money.AccountEntry, key string) *money.AccountEntry {
	for _, account := range accounts {
		if account.Code == key {
			return &account
		}
	}
	return nil
}

func main() {
	var mode = flag.String("mode", "table", "name of any account or 'table' or 'silent'")
	var logToCsv = flag.Bool("log_to_csv", false, "log to csv, false by default")
	flag.Parse()

	currency_layer.Init()
	yahoo_stock_api.Init()

	accountsFile := money.LoadAccounts()
	var accounts []money.AccountEntry = accountsFile.Accounts

	if *logToCsv {
		logNetWorth(&accountsFile)
	}

	switch *mode {
	case "silent":
		break

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		for _, account := range accounts {
			table.Append([]string{account.Name, getAccountValue(account).String()})
		}

		table.Append([]string{"Celkem", getTotalNetWorth(accountsFile.Accounts).String()})
		table.Render()
		break

	default:
		matchingAccount := findAccount(accounts, *mode)
		if matchingAccount != nil {
			value := getAccountValue(*matchingAccount)
			fmt.Printf("%.2f\n", value.Amount)
		} else {
			panic("bad mode")
		}
	}
}
