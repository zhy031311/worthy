package money

import "fmt"

type Money struct {
	Currency string `json:"currency"`
	Amount   float64 `json:"amount"`
}

func New(currency string, amount float64) Money {
	result := Money{}
	result.Currency = currency
	result.Amount = amount
	return result
}

func (money Money) String() string {
	if money.Currency == "USD" {
		return fmt.Sprintf("$%.2f", money.Amount)
	}
	return fmt.Sprintf("%.2f %s", money.Amount, money.Currency)
}

func (money *Money) Add(other Money) {
	if money.Currency != other.Currency {
		panic("trying to add incompatible currencies")
	}
	money.Amount += other.Amount
}

func (money *Money) DecreaseBy(other Money) {
	if money.Currency != other.Currency {
		panic("trying to decrease incompatible currencies")
	}
	money.Amount -= other.Amount
}

func (money *Money) AddInterestPercent(percent float64) {
	money.Amount *= 1 + (percent / 100)
}

func (money Money) IsLessThan(other Money) bool {
	if money.Currency != other.Currency {
		panic("trying to compare incompatible currencies")
	}
	return money.Amount < other.Amount
}

func (money Money) IsMoreThan(other Money) bool {
	return other.IsLessThan(money)
}
