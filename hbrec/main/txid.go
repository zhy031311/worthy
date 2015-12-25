package main

// TODO: This is probably some kind of international standard. Look it up.

import (
	"fmt"
	"strconv"
	"time"
	"strings"
)

func getYMDDate(input string) (ok bool, date time.Time) {
	year, err := strconv.Atoi(input[0:4])
	if err != nil {
		fmt.Println("can't parse year")
		return false, date
	}
	month, err := strconv.Atoi(input[4:6])
	if err != nil {
		fmt.Println("can't parse month")
		return false, date
	}
	day, err := strconv.Atoi(input[6:8])
	if err != nil {
		fmt.Println("can't parse day")
		return false, date
	}
	return true, time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
}

func getDMYDate(input string) (ok bool, date time.Time) {
	day, err := strconv.Atoi(input[0:2])
	if err != nil {
		fmt.Println("can't parse day")
		return false, date
	}
	month, err := strconv.Atoi(input[2:4])
	if err != nil {
		fmt.Println("can't parse month")
		return false, date
	}
	year, err := strconv.Atoi(input[4:8])
	if err != nil {
		fmt.Println("can't parse year")
		return false, date
	}
	return true, time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
}

func InferDateFromFirstPart(firstPart string) (ok bool, date time.Time) {
	parts := strings.Split(firstPart, "-")
	if len(parts) != 2 || len(parts[1]) != 8 {
		//fmt.Println(firstPart)
		//fmt.Println(parts[1])
		//fmt.Println("bad sizes in first part", len(parts), len(parts[1]))
		return false, date
	}
	switch parts[0] {
	case "205": fallthrough
	case "001": fallthrough
	case "000": fallthrough
	case "260": fallthrough

	case "228":  // TODO: not sure about 228
		ok, date = getDMYDate(parts[1])
		//fmt.Println("getDMYDate returned", ok, date)
		return ok, date
	case "120":
		ok, date = getYMDDate(parts[1])
		return ok, date
	default:
		fmt.Println("first part is bad:", parts[0])
		ok = false
	}

	return ok, date
}

func InferDateFromId(id string) (ok bool, date time.Time) {
	parts := strings.Fields(id)
	if len(parts) != 2 && len(parts) != 3 && len(parts) != 4 {
		//fmt.Println(id, "bad part count", len(parts))
		return false, date
	}

	ok, date = InferDateFromFirstPart(parts[0])
	//fmt.Println(id, "=>", ok, date)
	return ok, date

	/*
	year := 2004
	month := time.Month(7)  // January = 1
	day := 5
	return true, time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location())
	*/
}
