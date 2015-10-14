// Simple wrapper around http://www.currencylayer.com/.

package currency_layer

import (
	"encoding/json"
	"fmt"
	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const cachePath = "~/Dropbox/finance/currency_layer_cache.json"
const endpoint = "http://apilayer.net/api/live"
const cacheDuration = 24 * time.Hour

type cacheItem struct {
	Conversion float64 `json:"conversion"`
	Timestamp  string  `json:"timestamp"`
}

type cacheType struct {
	Conversions map[string]*cacheItem `json:"conversions"`
}

var cache cacheType
var accessKey string

func Init() {
	if util.FileExists(cachePath) {
		util.LoadJSONFileOrDie(cachePath, &cache)
	} else {
		cache.Conversions = make(map[string]*cacheItem)
	}

	body := util.ReadFile("~/Dropbox/finance/currency_layer_key")
	accessKey = strings.Trim(body, "\n\r ")
}

func showCache() {
	for key := range cache.Conversions {
		fmt.Println(key, cache.Conversions[key].Conversion)
	}
}

func writeCache() {
	// showCache()
	// fmt.Println("writing cache")

	bytes, err := json.Marshal(cache)
	if err != nil {
		panic(err)
	}
	util.WriteFile(cachePath, bytes)
}

func (item cacheItem) isConversionFresh() bool {
	takenAt, err := time.Parse(time.RFC3339, item.Timestamp)
	if err != nil {
		panic(err)
	}
	invalidAt := takenAt.Add(cacheDuration)
	return time.Now().Before(invalidAt)
}

func getConversionFromResponse(body []byte, key string) float64 {
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		panic(err)
	}
	fmt.Println(jsonBody)
	ok := jsonBody["success"].(bool)
	if !ok {
		panic("Error converting " + key)
	}
	fmt.Println("Converting: " + key)
	quote := jsonBody["quotes"].(map[string]interface{})[key]
	if quote == nil {
		panic("Quote is nil for " + key)
	}
	return quote.(float64)
}

func getConversion(from string, to string) float64 {
	if from == to {
		return 1
	}
	cacheKey := from + "_" + to
	if cachedConversion, cacheHit := cache.Conversions[cacheKey]; cacheHit {
		if cachedConversion.isConversionFresh() {
			return cachedConversion.Conversion
		}
	}

	values := url.Values{}
	values.Add("access_key", accessKey)
	values.Add("currencies", from+","+to)

	requestUrl := endpoint + "?" + values.Encode()
	resp, err := http.Get(requestUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	conversion := getConversionFromResponse(body, from+to)
	cache.Conversions[cacheKey] = &cacheItem{
		Conversion: conversion,
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	writeCache()
	return conversion
}

func Convert(from money.Money, to string) money.Money {
	factor := getConversion(from.Currency, to)
	return money.New(to, from.Amount*factor)
}

func CanConvert(from money.Money, to string) bool {
	return from.Currency != "BTC"
}
