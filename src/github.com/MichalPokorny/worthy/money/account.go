package money

import (
	"github.com/MichalPokorny/worthy/util"
)

type AccountEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type accountsFileData struct {
	Accounts []AccountEntry `json:"accounts"`
}

var ACCOUNTS_JSON_PATH = "~/Dropbox/finance/accounts.json"

func LoadAccounts() []AccountEntry {
	var data accountsFileData
	util.LoadJSONFileOrDie(ACCOUNTS_JSON_PATH, &data)
	return data.Accounts
}