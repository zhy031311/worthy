package portfolio

type Entry struct {
	Symbol string
	Amount int
}

type Portfolio []Entry

func NewEntry(symbol string, amount int) Entry {
	var entry Entry
	entry.Symbol = symbol
	entry.Amount = amount
	return entry
}

func (portfolio Portfolio) GetStockSymbols() []string {
	symbols := []string{}
	for _, entry := range portfolio {
		symbols = append(symbols, entry.Symbol)
	}
	return symbols
}
