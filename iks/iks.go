package iks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MichalPokorny/worthy/money"
	"github.com/MichalPokorny/worthy/util"
	"gopkg.in/yaml.v2"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var cache struct {
	Prices    map[string]float64 `json:"prices"`
	Timestamp string             `json:"timestamp"`
}

const cachePath = "~/dropbox/finance/iks_cache.json"
const cacheDuration = 24 * time.Hour

func loadCache() {
	if util.FileExists(cachePath) {
		util.LoadJSONFileOrDie(cachePath, &cache)
	}
}

func writeCache() {
	bytes, err := json.Marshal(cache)
	if err != nil {
		panic(err)
	}
	util.WriteFile(cachePath, bytes)
}

func isCacheFresh() bool {
	takenAt, err := time.Parse(time.RFC3339, cache.Timestamp)
	if err != nil {
		panic(err)
	}
	invalidAt := takenAt.Add(cacheDuration)
	return time.Now().Before(invalidAt)
}

func scrapePrices() {
	cache.Prices = make(map[string]float64)

	resp, err := http.Get("http://www.iks-kb.cz/web/fondy_denni_hodnoty.html")
	if err != nil {
		panic(err)
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	containerDiv, ok := scrape.Find(root, scrape.ById("fundView"))
	if !ok {
		panic("cannot find container div")
	}

	rows := scrape.FindAll(containerDiv, scrape.ByTag(atom.Tr))
	for _, row := range rows {
		tds := scrape.FindAll(row, scrape.ByTag(atom.Td))
		if len(tds) < 3 {
			continue
		}

		nameText := scrape.Text(tds[0])

		// Not used:
		// dateText := scrape.Text(tds[1])

		priceText := scrape.Text(tds[3])
		// Replace decimal comma by decimal dot.
		priceText = strings.Replace(priceText, ",", ".", 1)
		// Replace spaces separating thousands by nothing.
		priceText = strings.Replace(priceText, " ", "", -1)
		// \u00a0 = weird space
		priceText = strings.Replace(priceText, "\u00a0", "", -1)

		var err error
		cache.Prices[nameText], err = strconv.ParseFloat(priceText, 64)
		if err != nil {
			panic(err)
		}
	}

	cache.Timestamp = time.Now().Format(time.RFC3339)
}

type Investment struct {
	Invested int                `yaml:"invested"`
	Assets   map[string]float64 `yaml:"assets"`
}

func ParseInvestment(path string) (result Investment) {
	bytes := util.ReadFileBytes(path)
	err := yaml.Unmarshal(bytes, &result)
	if err != nil {
		panic(err)
	}
	return result
}

func GetInvestmentValue(investment Investment) money.Money {
	if cache.Prices == nil {
		loadCache()
		if cache.Prices == nil || !isCacheFresh() {
			scrapePrices()
			writeCache()
		}
	}
	result := money.New("CZK", 0)
	for asset, amount := range investment.Assets {
		price, ok := cache.Prices[asset]
		if !ok {
			panic("unknown asset: " + asset)
		}
		result.Amount += price * amount
	}
	return result
}
