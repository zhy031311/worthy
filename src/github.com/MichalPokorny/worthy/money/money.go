package money

import "fmt"

type Money struct {
	Currency string
	Amount   float64
}

func New(currency string, amount float64) Money {
	result := Money{}
	result.Currency = currency
	result.Amount = amount
	return result
}

func (money Money) String() string {
	return fmt.Sprintf("%.2f %s", money.Amount, money.Currency)
}
