package main

import (
	"fmt"
	"gopkg.in/readline.v1"
)


const CATEGORY_HOBBIES_AND_LEISURE = 51
const CATEGORY_TRAVEL = 99
const CATEGORY_PRAHA_DECIN_DOPRAVA = 132
const CATEGORY_FOOD = 42
const CATEGORY_SERVICE_CHARGE = 8
const CATEGORY_RENT = 23
const CATEGORY_MISC = 77

var categoryMap map[string]int

func init() {
	categoryMap = map[string]int{
		"hobbies_and_leisure": CATEGORY_HOBBIES_AND_LEISURE,
		"food": CATEGORY_FOOD,
		"travel": CATEGORY_TRAVEL,
		"service_charge": CATEGORY_SERVICE_CHARGE,
		"praha_decin_doprava": CATEGORY_PRAHA_DECIN_DOPRAVA,
		"rent": CATEGORY_RENT,
		"misc": CATEGORY_MISC,
	}
}

func promptCategory() int {
	var items []readline.PrefixCompleterInterface
	for key, _ := range categoryMap {
		items = append(items, readline.PcItem(key))
	}
	var completer = readline.NewPrefixCompleter(items...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "Category? > ",
		AutoComplete: completer,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	line, err := rl.Readline()
	if err != nil {
		return -1
	}
	if result, ok := categoryMap[line]; ok {
		return result
	} else {
		fmt.Println("no such category, bailing")
		return -1
	}
}
