package main

import (
	"io/ioutil"
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
	if operation.Wording == nil {
		return operationSupplement{
			ids: []string{},
			multipleTransactions: false,
		}
	}

	id := *operation.Wording

	for _, sep := range []string{";;", " ++ ", ", "} {
		if strings.Contains(*operation.Wording, sep) {
			return operationSupplement{
				ids: strings.Split(*operation.Wording, sep),
				multipleTransactions: true,
			}
		}
	}

	// TODO: multiple operations matchingn with single transaction
	// "(1/2) ..."
	for _, header := range []string{"(1/2) ", "(2/2) "} {
		if len(id) > len(header) && len(id) > len(header) && id[0:len(header)] == header {
			txid := id[len(header):]
			return operationSupplement{
				ids: []string{txid},
				splitTransaction: true,
			}
		}
	}

	return operationSupplement{ids: []string{id}}
}

func makePairingOperation(transaction Transaction) *homebank.Operation {
	reconcileTag := " (reconciled by hbrec " + time.Now().Format("2006-01-02") + ")"

	info := transaction.SystemDescription + reconcileTag
	status := homebank.OPERATION_RECONCILED
	date := transaction.SettlementDate
	amount := transaction.Amount

	dateFormat := "01.02.2006"
	s := strings.Split(transaction.AVField4, " ")
	av4D := s[0]
	fmt.Println("AV1: [", transaction.AVField1, "] AV4: [", av4D, "] info: [", info, "] amount: [", amount, "]")
	av4Date, e := time.Parse(dateFormat, av4D)
	if e == nil {
		date = av4Date
	}

	var category *int

	if strings.Contains(info, "Odměna za služby") || strings.Contains(info, "Výběr z bankomatu - poplatek") || strings.Contains(info, "Dotaz na zůstatek v bankomatu") || strings.Contains(transaction.SenderIdentification, "POPLATEK ZA POLOŽKY") {
		i := 8
		category = &i  // service charge
	} else if strings.Contains(transaction.SenderIdentification, "BONUS ZA VÝBĚR ATM KB") {
		i := 8
		category = &i  // service charge
		info = "Bonus za výběr z ATM" + reconcileTag
	} else if strings.Contains(transaction.AVField1, "DAMEJIDLO.CZ") {
		i := 42
		category = &i  // food
		info = "damejidlo.cz" + reconcileTag
	} else if strings.Contains(transaction.AVField1, "STEAMGAMES.COM") {
		i := 51
		category = &i  // hobbies & leisure
		info = "hry na Steamu" + reconcileTag
	} else if strings.Contains(transaction.AVField1, "CAJOVNA") {
		i := 51
		category = &i  // hobbies & leisure
		info = "cajovna (" + transaction.AVField1 + ")" + reconcileTag
	} else if strings.Contains(transaction.SenderIdentification, "DOBITI - O2") {
		i := 11
		category = &i  // dobiti mobilu
		info = "dobiti mobilu" + reconcileTag
	} else if strings.Contains(transaction.AVField1, "CD DECIN HL.N.") && amount <= -120 && amount >= -150 {
		i := 132
		category = &i  // decin <=> praha
		info = "jizdenka z Decina do Prahy" + reconcileTag
	} else if strings.Contains(transaction.AVField1, "CD PRAHA-HOLESOVICE") && amount <= -120 && amount >= -150 {
		i := 132
		category = &i  // decin <=> praha
		info = "jizdenka z Prahy do Decina" + reconcileTag
	} else {
		fmt.Println("! no inferred category !")
		return nil
	}

	return &homebank.Operation{
		// TODO: infer date instead
		Date: homebank.DateToHomebank(date),
		Amount: transaction.Amount,
		Wording: &transaction.TransactionIdentification,
		Info: &info,
		Status: &status,
		Category: category,
	}
}

func loadTransactions() (rangeBegin time.Time, transactions []Transaction) {
	baseDir := "/home/prvak/dropbox/finance/vypisy"
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	rangeBegin = time.Now()
	transactions = make([]Transaction, 0)
	transactionsById := make(map[string]bool)
	for _, file := range files {
		path := baseDir + "/" + file.Name()
		export := ParseCSVFile(path)

		if export.RangeBegin.Before(rangeBegin) {
			rangeBegin = export.RangeBegin
		}

		for _, transaction := range export.Transactions {
			if _, ok := transactionsById[transaction.TransactionIdentification]; !ok {
				transactionsById[transaction.TransactionIdentification] = true
				transactions = append(transactions, transaction)
			}
		}
	}
	return rangeBegin, transactions
}

func main() {
	rangeBegin, transactions := loadTransactions()

	// transaction ID => Homebank operation
	idToHomebank := make(map[string][]homebank.Operation)

	accountKey := 2 // "KB bezny"
	accounting, err := homebank.ParseHomebankFile("/home/prvak/dropbox/finance/ucetnictvi.xhb")
	if err != nil {
		panic(err)
	}

	transactionByIdentification := make(map[string]Transaction)
	for _, transaction := range transactions {
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

		if supplement.isBefore(rangeBegin) {
			// operation took place before export
			isBefore = true
		}

		if homebank.ParseHomebankDate(operation.Date).Before(rangeBegin) {
			// date in Homebank is before export
			isBefore = true
		}

		if isBefore {
			// multitransaction, one of them is before start of export
			continue
		}

		if pairedCount < len(supplement.ids) {
			if len(supplement.ids) == 1 && pairedCount == 0 {
				fmt.Printf("unpaired: %v %v memo=%s %s", homebank.ParseHomebankDate(operation.Date), operation.Amount, *operation.Info, *operation.Wording)
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
	suggestedOperations := make([]homebank.Operation, 0)
	for _, transaction := range transactions {
		id := transaction.TransactionIdentification

		_, ok := operationsHaveIdentification[id]
		if ok {
			// paired
			continue
		}

		//fmt.Printf("unpaired: %v settlement=%v id=[%v] %v %v\n", transaction.Amount, transaction.SettlementDate.Format(CzechDate), transaction.TransactionIdentification, transaction.SystemDescription, transaction.SenderIdentification)
		//fmt.Printf("%+v\n", transaction)
		fmt.Printf("unpaired: %s\n", transaction.HumanString())

		op := makePairingOperation(transaction)
		if op != nil {
			op.Account = accountKey
			suggestedOperations = append(suggestedOperations, *op)
		}
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

	if len(suggestedOperations) > 0 {
		fmt.Println("Suggested additions:")
		for _, op := range suggestedOperations {
			fmt.Printf("%s\n", op.GetXml())
		}
	}

	//homebank.WriteHomebankFile(accounting, "/home/prvak/out.xhb")
}
