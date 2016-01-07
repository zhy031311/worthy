package stock

import (
	"github.com/MichalPokorny/worthy/money"
)

type Ticker struct {
	Symbol string
	Sell   money.Money
}

type TradingDay struct {
	Symbol string
	Date string

	Open float64
	High float64
	Low float64
	Close float64

	AdjustedClose float64
}
