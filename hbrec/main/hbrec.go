package main

import (
	"github.com/MichalPokorny/worthy/homebank"
	"fmt"
	"strings"
)

func main() {
	export := ParseCSVFile("/home/prvak/dropbox/finance/vypisy/437531160267_20140804_20151223.csv")

	accountKey := 2 // "KB bezny"
	accounting, err := homebank.ParseHomebankFile("/home/prvak/dropbox/finance/ucetnictvi.xhb")
	if err != nil {
		panic(err)
	}

	transactionByIdentification := make(map[string]Transaction)
	for _, transaction := range export.Transactions {
		transactionByIdentification[transaction.TransactionIdentification] = transaction
	}

	operations := accounting.GetAccountOperations(accountKey)

	operationsHaveIdentification := make(map[string]bool)

	fmt.Println("In Homebank, but missing in KB:")
	for _, operation := range operations {
		id := operation.Wording

		if homebank.ParseHomebankDate(operation.Date).Before(export.RangeBegin) {
			continue
		}

		// TODO: match with multiple transactions
		isBefore := false
		for _, sep := range []string{";;", " ++ ", ", "} {
			if strings.Contains(operation.Wording, sep) {
				for _, subid := range strings.Split(operation.Wording, sep) {
					operationsHaveIdentification[subid] = true

					ok, date := InferDateFromId(subid)
					if ok && date.Before(export.RangeBegin) {
						isBefore = true
					}

				}
				if !isBefore {
					fmt.Println("multitrans")
				}
				break
			}
		}

		// TODO: multiple operations matchingn with single transaction
		// "(1/2) ..."
		for _, header := range []string{"(1/2) ", "(2/2) "} {
			if len(id) > len(header) &&len(id) > len(header) && id[0:len(header)] == header {
				txid := id[len(header):]
				operationsHaveIdentification[txid] = true

				ok, date := InferDateFromId(txid)
				if ok && date.Before(export.RangeBegin) {
					isBefore = true
				}
				if !isBefore {
					fmt.Println("splittrans")
				}
				break
			}
		}

		if isBefore {
			// multitransaction, one of them is before start of export
			continue
		}

		operationsHaveIdentification[id] = true
		ok, date := InferDateFromId(id)
		if ok && date.Before(export.RangeBegin) {
			// operation took place before export
			continue
		}

		if transaction, ok := transactionByIdentification[id]; ok {
			// NOTE: Inferred date seems to be the date of payment.
			// Settlement is a few days after that.

			// TODO: check that this pairing more or less matches
			// (amount, date)
			_ = transaction
			//fmt.Println("paired OK:", id, transaction)
		} else {
			fmt.Printf("unpaired: %v %v memo=%s %s", homebank.ParseHomebankDate(operation.Date), operation.Amount, operation.Info, operation.Wording)
			if ok {
				fmt.Printf(" inferred date %v", date)
			}
			fmt.Printf("\n")
		}
	}

	fmt.Println("In KB, but missing in Homebank:")
	for _, transaction := range export.Transactions {
		id := transaction.TransactionIdentification

		_, ok := operationsHaveIdentification[id]
		if ok {
			// paired
			continue
		}

		fmt.Printf("unpaired: %v settlement=%v id=%v %v %v\n", transaction.Amount, transaction.SettlementDate.Format(CzechDate), transaction.TransactionIdentification, transaction.SystemDescription, transaction.SenderIdentification)
	}
}
