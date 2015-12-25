package main

import (
	"github.com/MichalPokorny/worthy/homebank"
	"fmt"
	//"time"
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
	for _, operation := range operations {
		id := operation.Wording
		transaction, ok := transactionByIdentification[id]
		if ok {
			_ = transaction
			//fmt.Println("paired OK:", id, transaction)
		} else {
			// fmt.Println("unpaired:", operation.Wording, "date=", operation.Date)
			ok, date := InferDateFromId(operation.Wording)
			if ok && date.Before(export.RangeBegin) {
				// Unpaired, but OK.
				//  fmt.Println("unpaired but ok:", operation.Wording, "date=", operation.Date)
			} else {
				 fmt.Println("unpaired:", operation.Wording, "date=", operation.Date)
			}
		}
	}
	fmt.Println(export.RangeBegin)
	fmt.Println(export.RangeEnd)
	//fmt.Println(operations)
}
