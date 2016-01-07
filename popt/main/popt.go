package main

import (
	"os"
	"math/rand"
	"fmt"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"math"
)

const startDate = "2015-01-01"
const endDate = "2015-12-31"

func getRelativeChanges(ticker string) []float64 {
	days := yahoo_stock_api.GetHistoricalPrices(ticker, startDate, endDate)
	changes := make([]float64, len(days))
	for i, day := range days {
		changes[i] = day.AdjustedClose / days[0].AdjustedClose
	}
	return changes
}

type Metrics struct {
	Earnings float64
	Volatility float64
	// TODO: Sharpe ratio, ...?
}

func getMetrics(dailyReturns []float64) (metrics Metrics) {
	// TODO: return on *benchmark asset*!
	metrics.Earnings = dailyReturns[len(dailyReturns) - 1]

	averageReturn := 0.0
	for _, r := range dailyReturns {
		averageReturn += r / float64(len(dailyReturns))
	}
	variance := 0.0
	for _, r := range dailyReturns {
		variance += (r - averageReturn) * (r - averageReturn)
	}
	metrics.Volatility = math.Sqrt(variance)
	return metrics
}

func makeMix(symbols []string, weights []float64) []float64 {
	prices := make([][]float64, len(symbols))
	for i, symbol := range symbols {
		prices[i] = getRelativeChanges(symbol)
	}
	// TODO: assert that downloaded trading days are the same
	mix := make([]float64, len(prices[0]))
	for i := 0; i < len(prices[0]); i++ {
		mix[i] = 0
		for j, weight := range weights {
			mix[i] += weight * prices[j][i]
		}
		mix[i] /= mix[0]
	}
	return mix
}

func main() {
	symbols := []string{"AAPL", "GOOG", "MSFT"}
	for _, symbol := range symbols {
		metrics := getMetrics(getRelativeChanges(symbol))
		fmt.Println(symbol, "earnings:", metrics.Earnings, "volatility:", metrics.Volatility)
	}

	f, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < 10000; i++ {
		weights := make([]float64, len(symbols))
		sum := 0.0
		for j := 0; j < len(symbols); j++ {
			weights[j] = float64(rand.Int63n(1000))
			sum += weights[j]
		}
		for j := 0; j < len(symbols); j++ {
			weights[j] /= sum
		}

		mix := makeMix(symbols, weights)
		metrics := getMetrics(mix)
		// fmt.Println("earnings:", , "volatility:", metrics.Volatility)
		f.WriteString(fmt.Sprintf("%.3f %.3f\n", metrics.Earnings, metrics.Volatility))
	}
}
