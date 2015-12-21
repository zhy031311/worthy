package money

import (
	"github.com/MichalPokorny/worthy/util"
)

type AccountEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
	HiddenInTable bool `json:"hidden_in_table"`

	// Pick exactly one of the following:

	// Path to file that contains the portfolio in the account.
	Path *string `json:"path"`

	// Value of the account in CZK.
	Value *Money `json:"value"`

	// Path to Homebank file whose non-closed non-asset accounts
	// sum up to this account's value in CZK.
	HomebankPath *string `json:"homebank_path"`

	// Path to YML file describing IKS investment.
	IksPath *string `json:"iks_path"`
}

type AccountsFileData struct {
	Accounts []AccountEntry `json:"accounts"`
	CsvOrder []string       `json:"csv_order"`
	CsvPath  string         `json:"csv_path"`
}

var ACCOUNTS_JSON_PATH = "~/dropbox/finance/accounts.json"

func LoadAccounts() (result AccountsFileData) {
	util.LoadJSONFileOrDie(ACCOUNTS_JSON_PATH, &result)
	return result
}
