package blockchain_info_api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type walletInfo struct {
	Hash160 string `json:"hash160"`
	Address string `json:"address"`
	NTx int `json:"n_tx"`
	NUnredeemed int `json:"n_unredeemed"`
	TotalReceived int `json:"total_received"` // satoshi?
	TotalSent int `json:"total_sent"` // satoshi?
	FinalBalance int `json:"final_balance"` // satoshi?
	// TODO: plus txs, array of last transactions
}

func getEndpoint(address string) string {
	return "https://blockchain.info/address/" + address + "?format=json"
}

func GetBalance(address string) float64 {
	resp, err := http.Get(getEndpoint(address))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var info walletInfo
	if err := json.Unmarshal(body, &info); err != nil {
		panic(err)
	}
	return float64(0.00000001) * float64(info.FinalBalance)
}
