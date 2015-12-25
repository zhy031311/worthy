package main

import (
	"math"
	"time"
	"github.com/MichalPokorny/worthy/homebank"
	"fmt"
	"strings"
)

type pairing map[string][]homebank.Operation

type operationSupplement struct {
	ids []string
	multipleTransactions bool
	splitTransaction bool
}

func (supplement operationSupplement) isBefore(cutoff time.Time) bool {
	for _, id := range supplement.ids {
		ok, date := InferDateFromId(id)
		if ok && date.Before(cutoff) {
			// operation took place before export
			return true
		}
	}
	return false
}

func getSupplement(operation homebank.Operation) operationSupplement {
	id := operation.Wording

	for _, sep := range []string{";;", " ++ ", ", "} {
		if strings.Contains(operation.Wording, sep) {
			return operationSupplement{
				ids: strings.Split(operation.Wording, sep),
				multipleTransactions: true,
			}
		}
	}

	// TODO: multiple operations matchingn with single transaction
	// "(1/2) ..."
	for _, header := range []string{"(1/2) ", "(2/2) "} {
		if len(id) > len(header) &&len(id) > len(header) && id[0:len(header)] == header {
			txid := id[len(header):]
			return operationSupplement{
				ids: []string{txid},
				splitTransaction: true,
			}
		}
	}

	return operationSupplement{ids: []string{id}}
}

func main() {
	export := ParseCSVFile("/home/prvak/dropbox/finance/vypisy/437531160267_20140804_20151223.csv")

	// transaction ID => Homebank operation
	idToHomebank := make(map[string][]homebank.Operation)

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
		supplement := getSupplement(operation)
		pairedCount := 0

		for _, id := range supplement.ids {
			operationsHaveIdentification[id] = true
			if transaction, ok := transactionByIdentification[id]; ok {
				// NOTE: Inferred date seems to be the date of payment.
				// Settlement is a few days after that.

				// TODO: check that this pairing more or less matches
				// (amount, date)
				pairedCount++
				_ = transaction
			}

			_, ok := idToHomebank[id]
			if !ok {
				idToHomebank[id] = make([]homebank.Operation, 0)
			}
			idToHomebank[id] = append(idToHomebank[id], operation)
		}

		// TODO: match with multiple transactions
		// TODO: multiple operations matchingn with single transaction
		// "(1/2) ..."

		isBefore := false

		if supplement.isBefore(export.RangeBegin) {
			// operation took place before export
			isBefore = true
		}

		if homebank.ParseHomebankDate(operation.Date).Before(export.RangeBegin) {
			// date in Homebank is before export
			isBefore = true
		}

		if isBefore {
			// multitransaction, one of them is before start of export
			continue
		}

		if pairedCount < len(supplement.ids) {
			if len(supplement.ids) == 1 && pairedCount == 0 {
				fmt.Printf("unpaired: %v %v memo=%s %s", homebank.ParseHomebankDate(operation.Date), operation.Amount, operation.Info, operation.Wording)
				ok, date := InferDateFromId(supplement.ids[0])
				if ok {
					fmt.Printf(" inferred date %v", date)
				}
				fmt.Printf("\n")
			} else {
				fmt.Println("partially paired:", operation, supplement)
			}
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

		//fmt.Printf("unpaired: %v settlement=%v id=[%v] %v %v\n", transaction.Amount, transaction.SettlementDate.Format(CzechDate), transaction.TransactionIdentification, transaction.SystemDescription, transaction.SenderIdentification)
		//fmt.Printf("%+v\n", transaction)
		fmt.Printf("unpaired: %s\n", transaction.HumanString())
	}

	for id, operations := range idToHomebank {
		sum := 0.0
		for _, operation := range operations {
			sum += operation.Amount
		}
		transaction, ok := transactionByIdentification[id]
		if !ok {
			// unpaired
			continue
		}

		diff := math.Abs(sum - transaction.Amount)
		if diff < 1 {
			// looks OK
			continue
		}
		fmt.Println(id, sum, transaction.Amount)
	}
}
