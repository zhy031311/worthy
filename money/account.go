package money

import (
	"github.com/MichalPokorny/worthy/util"
)

type AccountEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`

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

type accountsFileData struct {
	Accounts []AccountEntry `json:"accounts"`
}

var ACCOUNTS_JSON_PATH = "~/dropbox/finance/accounts.json"

func LoadAccounts() []AccountEntry {
	var data accountsFileData
	util.LoadJSONFileOrDie(ACCOUNTS_JSON_PATH, &data)
	return data.Accounts
}
