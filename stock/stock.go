package stock

import (
	"github.com/MichalPokorny/worthy/money"
)

type Ticker struct {
	Symbol string
	Sell   money.Money
}
