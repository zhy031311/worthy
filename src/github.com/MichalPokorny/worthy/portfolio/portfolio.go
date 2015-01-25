package portfolio

const (
	money = iota
	stock = iota
)

type Entry struct {
	Ticker string
	Amount int
}

type Portfolio []Entry

func NewEntry(ticker string, amount int) Entry {
	var entry Entry
	//entry.entryType = entryType
	entry.Ticker = ticker
	entry.Amount = amount
	return entry
}

func (portfolio Portfolio) GetStockSymbols() []string {
	symbols := []string{}
	for _, entry := range portfolio {
		symbols = append(symbols, entry.Ticker)
	}
	return symbols
}

