package homebank

// TODO: Make the format more complete?

import (
	"encoding/xml"
	"io/ioutil"
)

// Values of the Type field
const ASSETS_ACCOUNT = 3

// Closed account
const FLAG_CLOSED = 2

type Account struct {
	Key     int     `xml:"key,attr"`
	Pos     int     `xml:"pos,attr"`
	Type    int     `xml:"type,attr"`
	Name    string  `xml:"name,attr"`
	Initial float64 `xml:"initial,attr"`
	Minimum int     `xml:"minimum,attr"`
	Flags   int     `xml:"flags,attr"`
}

type Operation struct {
	Date       int     `xml:"date,attr"`
	Amount     float64 `xml:"amount,attr"`
	Account    int     `xml:"account,attr"`
	Paymode    int     `xml:"paymode,attr"`
	DstAccount int     `xml:"dst_account,attr"`
	Status     int     `xml:"st,attr"`
	Category   int     `xml:"category,attr"`
	Wording    string  `xml:"wording,attr"`
	Info       string  `xml:"info,attr"`
}

type HomebankFile struct {
	Version string `xml:"v,attr"`
	D       string `xml:"d,attr"` // TODO: what's this?
	// TODO: <properties>
	// TODO: all other tags in the XML
	Accounts   []Account   `xml:"account"`
	Operations []Operation `xml:"ope"`
}

func (homebank *HomebankFile) GetAccountBalance(accountId int) float64 {
	var balance float64 = -1
	for _, account := range homebank.Accounts {
		if account.Key == accountId {
			balance = account.Initial
		}
	}
	if balance == -1 {
		panic("account not found")
	}

	for _, operation := range homebank.Operations {
		if operation.Account != accountId {
			continue
		}

		balance += operation.Amount
	}
	return balance
}

func ParseHomebankFile(path string) (result HomebankFile, err error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return result, err
	}
	err = xml.Unmarshal(bytes, &result)
	return result, err
}

/*
func main() {
	result, err := ParseHomebankFile("/home/prvak/dropbox/finance/ucetnictvi.xhb")
	if err != nil {
		panic(err)
	}

	total := float64(0)
	for _, account := range(result.Accounts) {
		if account.Flags & FLAG_CLOSED == FLAG_CLOSED {
			// Skip closed accounts.
			continue
		}
		if account.Type == ASSETS_ACCOUNT {
			// We manage assets accounts ourselves.
			continue
		}
		fmt.Println(account)
		fmt.Println(result.GetAccountBalance(account.Key))
		fmt.Println()
	}

	fmt.Printf("Version: %v\n", result.Version)
	fmt.Println(total)
}
*/
