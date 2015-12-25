package main

// enconv -L czech -x utf8 o543thgwjfw.csv

import (
	"strconv"
	"strings"
	"os"
	"encoding/csv"
	//"golang.org/x/text/encoding"
//	"golang.org/x/text/transform"
//	"golang.org/x/text/encoding/charmap"
//	"flag"
	"fmt"
)

// windows-1250 encoding

type KBExport struct {
	CreationDate string
	AccountNumber string
	IBAN string
	AccountName string
	OwnAccountName string
	RangeBegin string
	RangeEnd string
	// TODO: predchozi vypis ze dne
	NumberOfItems int
	InitialBalance float64
	// TODO bad name
	TotalMinus float64
	// TODO bad name
	TotalPlus float64
	FinalBalance float64

	Transactions []Transaction
};

type Transaction struct {
	SettlementDate string
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
};

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
			export.AccountNumber = row[1]
			break
		case "IBAN":
			export.IBAN = row[1]
			break
		case "Název účtu":
			export.AccountName = row[1]
			break
		case "Vlastní název účtu":
			export.OwnAccountName = row[1]
			break
		case "Výpis za období":
			export.RangeBegin = row[1]
			// TODO: after this, row[0] = '', row[1] is range end
			break
		case "Počet položek":
			value, err := strconv.Atoi(strings.TrimSpace(row[1]))
			if err != nil {
				panic(err)
			}
			export.NumberOfItems = value
			break
		default:
			if len(row) == 20 && row[0] != "Datum splatnosti" {
				//for i, thing := range row {
				//	fmt.Println(i, thing)
				//}

				transaction := Transaction{
					SettlementDate: row[0],
					OtherBankDeductionDate: row[1],
					Counteraccount: row[2],
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
					AVField1: row[15],
					AVField2: row[16],
					AVField3: row[17],
					AVField4: row[18],
				}
				export.Transactions = append(export.Transactions, transaction)
				break
			}
			fmt.Println(len(row))
			fmt.Println(row)
		}
	}
	return export
}
