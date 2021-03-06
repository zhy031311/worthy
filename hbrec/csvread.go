package main

// need to fix windows-1250 encoding
// enconv -L czech -x utf8 o543thgwjfw.csv

import (
	"regexp"
	"fmt"
	"time"
	"strconv"
	"strings"
	"os"
	"encoding/csv"
	"math"
)

type KBExport struct {
	CreationDate string
	AccountNumber string
	IBAN string
	AccountName string
	OwnAccountName string
	RangeBegin time.Time
	RangeEnd time.Time
	// TODO: predchozi vypis ze dne
	NumberOfItems int
	InitialBalance float64
	// TODO bad name
	TotalMinus float64
	// TODO bad name
	TotalPlus float64
	FinalBalance float64

	ExportNumber string
	PreviousExportFromDate string

	Transactions []Transaction
};

type Transaction struct {
	SettlementDate time.Time
	// datum odepsani z jine banky
	OtherBankDeductionDate string
	Counteraccount string
	CounteraccountName string
	Amount float64
	OriginalAmount float64
	OriginalCurrency string
	CurrencyConversion float64
	VariableSymbol string
	ConstantSymbol string
	SpecificSymbol string
	TransactionIdentification string
	SystemDescription string
	SenderIdentification string
	ReceiverIdentification string
	AVField1 string
	AVField2 string
	AVField3 string
	AVField4 string
}

func humanizeString(s string) string {
	re := regexp.MustCompile("( +)")
	return re.ReplaceAllString(s, " ")
}

func (transaction Transaction) HumanString() string {
	amount := transaction.Amount
	settled := transaction.SettlementDate.Format(CzechDate)
	str := fmt.Sprintf("%v %s", amount, settled)

	if transaction.SystemDescription != "" {
		str += " " + transaction.SystemDescription
	}
	if transaction.SenderIdentification != "" {
		str += " " + transaction.SenderIdentification
	}
	if transaction.ReceiverIdentification != "" {
		str += " " + transaction.ReceiverIdentification
	}
	if transaction.AVField1 != "" {
		str += " AV1=" + humanizeString(transaction.AVField1)
	}
	if transaction.AVField2 != "" {
		str += " AV2=" + humanizeString(transaction.AVField2)
	}
	if transaction.AVField3 != "" {
		str += " AV3=" + humanizeString(transaction.AVField3)
	}
	if transaction.AVField4 != "" {
		str += " AV4=" + humanizeString(transaction.AVField4)
	}
	str += " [" + transaction.TransactionIdentification + "]"
	return str
}

const CzechDate = "02.01.2006"

func ParseCzechDate(x string) time.Time {
	time, err := time.Parse(CzechDate, x)
	if err != nil {
		panic(err)
	}
	return time
}

func ParseAmountOrEmpty(x string) float64 {
	if x == "" {
		return 0.0;
	}
	return ParseAmount(x);
}

func ParseAmount(x string) float64 {
	// decimal comma => point
	parsed := strings.Replace(x, ",", ".", 1)
	value, err := strconv.ParseFloat(parsed, 64)
	if err != nil {
		panic("cannot parse " + x)
	}
	return value
}

// csv file assumed to be in UTF-8 encoding, which is not the default
func ParseCSVFile(csvPath string) KBExport {
	file, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(file)
	r.Comma = ';'
	r.FieldsPerRecord = -1  // don't check
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	var export KBExport
	export.Transactions = make([]Transaction, 0)
	for _, row := range records {
		switch row[0] {
		case "Datum vytvoření souboru":
			export.CreationDate = row[1]
			break
		case "Číslo účtu":
			export.AccountNumber = strings.TrimSpace(row[1])
			break
		case "IBAN":
			export.IBAN = row[1]
			break
		case "Název účtu":
			export.AccountName = strings.TrimSpace(row[1])
			break
		case "Vlastní název účtu":
			export.OwnAccountName = row[1]
			break
		case "Počáteční zůstatek":
			if row[1] == "-" {
				export.InitialBalance = math.NaN()
			} else {
				export.InitialBalance = ParseAmount(row[1])
			}
			break
		// NOTE: HARD SPACE OR WHAT?! The following string is weird.
		case "Výpis za období":
			export.RangeBegin = ParseCzechDate(row[1])
			// TODO: after this, row[0] = '', row[1] is range end
			break
		case "":
			export.RangeEnd = ParseCzechDate(row[1])
			// TODO: horrible hack
			break
		case "Počet položek":
			value, err := strconv.Atoi(strings.TrimSpace(row[1]))
			if err != nil {
				panic(err)
			}
			export.NumberOfItems = value
			break
		case "Celkem odepsáno (-)":
			export.TotalMinus = ParseAmount(row[1])
			break
		case "Celkem připsáno (+)":
			export.TotalPlus = ParseAmount(row[1])
			break
		case "Konečný zůstatek":
			export.FinalBalance = ParseAmount(row[1])
			break
		case "Číslo výpisu":
			export.ExportNumber = row[1]
			break

		// NOTE: probably weird whitespace below.
		case "Předchozí výpis ze dne":
			export.PreviousExportFromDate = row[1]
			break
		default:
			if len(row) == 20 && row[0] != "Datum splatnosti" {
				//for i, thing := range row {
				//	fmt.Println(i, thing)
				//}

				transaction := Transaction{
					SettlementDate: ParseCzechDate(row[0]),
					OtherBankDeductionDate: row[1],
					Counteraccount: strings.TrimSpace(row[2]),
					CounteraccountName: row[3],
					Amount: ParseAmount(row[4]),
					OriginalAmount: ParseAmountOrEmpty(row[5]),
					OriginalCurrency: row[6],
					CurrencyConversion: ParseAmountOrEmpty(row[7]),
					VariableSymbol: row[8],
					ConstantSymbol: row[9],
					SpecificSymbol: row[10],
					TransactionIdentification: row[11],
					SystemDescription: row[12],
					SenderIdentification: row[13],
					ReceiverIdentification: row[14],
					AVField1: strings.TrimSpace(row[15]),
					AVField2: strings.TrimSpace(row[16]),
					AVField3: strings.TrimSpace(row[17]),
					AVField4: strings.TrimSpace(row[18]),
				}
				export.Transactions = append(export.Transactions, transaction)
				break
			} else if row[0] == "MojeBanka, export transakční historie" || row[0] == "Datum splatnosti" {
				// Header rows, ignored.
			} else {
				fmt.Println("Unparsed row:", row)
			}
		}
	}
	return export
}
