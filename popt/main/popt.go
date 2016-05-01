package main

import (
	"os"
	"math/rand"
	"fmt"
	"github.com/MichalPokorny/worthy/yahoo_stock_api"
	"math"
	"strings"
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
	if len(symbols) != len(weights) {
		panic("bad sizes")
	}
	prices := make([][]float64, len(symbols))
	for i, symbol := range symbols {
		prices[i] = getRelativeChanges(symbol)

		if len(prices[i]) != len(prices[0]) {
			panic("bad symbol: " + symbol)
		}
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

func showCurrentAllocation() {
	symbols := []string{
		"AAPL",
		"ATVI",
		"BABA",
		"CAH",
		"CP",
		"DIS",
		"EMN",
		"GOOG",
		"HBI",
		"HRL",
		"IFF",
		"IVZ",
		"IWV",
		"LOW",
		"QQQ",
		"TSLA",
	}
	// TODO: this should be what I held initially...
	weights := []float64{
		54636.79, // AAPL
		33687.72, // ATVI
		70550.92, // BABA
		47845.69, // CAH
		5541.96, // CP
		46334.51, // DIS
		23669.76, // EMN
		106685.11, // GOOG
		67147.23, // HBI
		65008.47, // HRL
		28472.96, // IFF
		59991.44, // IVZ
		55926.77, // IWV
		61834.45, // LOW
		56643.59, // QQQ
		31544.99, // TSLA
	}
	normalize(weights)
	mix := makeMix(symbols, weights)
	metrics := getMetrics(mix)
	fmt.Println("base earnings:", metrics.Earnings, "volatility:", metrics.Volatility)
}

func normalize(x []float64) {
	sum := 0.0
	for j := 0; j < len(x); j++ {
		sum += x[j]
	}
	for j := 0; j < len(x); j++ {
		x[j] /= sum
	}
}

//const SP500 = `MMM ABT ABBV ACN ACE ATVI ADBE ADT AAP AES AET AFL AMG A GAS APD ARG AKAM AA AGN ALXN ALLE ADS ALL GOOGL GOOG MO AMZN AEE AAL AEP AXP AIG AMT AMP ABC AME AMGN APH APC ADI AON APA AIV AAPL AMAT ADM AIZ T ADSK ADP AN AZO AVGO AVB AVY BHI BLL BAC BK BCR BXLT BAX BBT BDX BBBY BBY BIIB BLK HRB BA BWA BXP BSX BMY BRCM CHRW CA CVC COG CAM CPB COF CAH HSIC KMX CCL CAT CBG CBS CELG CNP CTL CERN CF SCHW CHK CVX CMG CB CHD CI XEC CINF CTAS CSCO C CTXS CLX CME CMS COH KO CCE CTSH CL CPGX CMCSA CMA CAG COP CNX ED STZ GLW COST CCI CSRA CSX CMI CVS DHI DHR DRI DVA DE DLPH DAL XRAY DVN DO DFS DISCA DISCK DG DLTR D DOV DOW DPS DTE DD DUK DNB ETFC EMN ETN EBAY ECL EIX EW EA EMC EMR ENDP ESV ETR EOG EQT EFX EQIX EQR ESS EL ES EXC EXPE EXPD ESRX XOM FFIV FB FAST FDX FIS FITB FSLR FE FISV FLIR FLS FLR FMC FTI F BEN FCX FTR GME GPS GRMN GD GE GGP GIS GM GPC GILD GS GT GWW HAL HBI HOG HAR HRS HIG HAS HCA HCP HP HES HPE HD HON HRL HST HPQ HUM HBAN ITW ILMN IR INTC ICE IBM IP IPG IFF INTU ISRG IVZ IRM JEC JBHT JNJ JCI JPM JNPR KSU K KEY GMCR KMB KIM KMI KLAC KSS KHC KR LB LLL LH LRCX LM LEG LEN LVLT LUK LLY LNC LLTC LMT L LOW LYB MTB MAC M MNK MRO MPC MAR MMC MLM MAS MA MAT MKC MCD MHFI MCK MJN WRK MDT MRK MET KORS MCHP MU MSFT MHK TAP MDLZ MON MNST MCO MS MOS MSI MUR MYL NDAQ NOV NAVI NTAP NFLX NWL NFX NEM NWSA NWS NEE NLSN NKE NI NBL JWN NSC NTRS NOC NRG NUE NVDA ORLY OXY OMC OKE ORCL OI PCAR PH PDCO PAYX PYPL PNR PBCT POM PEP PKI PRGO PFE PCG PM PSX PNW PXD PBI PCL PNC RL PPG PPL PX PCP PCLN PFG PG PGR PLD PRU PEG PSA PHM PVH QRVO PWR QCOM DGX RRC RTN O RHT REGN RF RSG RAI RHI ROK COL ROP ROST RCL R CRM SNDK SCG SLB SNI STX SEE SRE SHW SIG SPG SWKS SLG SJM SNA SO LUV SWN SE STJ SWK SPLS SBUX HOT STT SRCL SYK STI SYMC SYF SYY TROW TGT TEL TE TGNA THC TDC TSO TXN TXT HSY TRV TMO TIF TWX TWC TJX TMK TSS TSCO RIG TRIP FOXA FOX TSN TYC USB UA UNP UAL UNH UPS URI UTX UHS UNM URBN VFC VLO VAR VTR VRSN VRSK VZ VRTX VIAB V VNO VMC WMT WBA DIS WM WAT ANTM WFC HCN WDC WU WY WHR WFM WMB WLTW WEC WYN WYNN XEL XRX XLNX XL XYL YHOO YUM ZBH ZION ZTS`;
const SP500 = `MMM ABT ABBV ACN ACE ATVI ADBE ADT AAP AES AET AFL AMG A GAS APD ARG AKAM AA AGN ALXN ALLE ADS ALL GOOGL GOOG MO AMZN AEE AAL AEP AXP AIG AMT AMP ABC AME AMGN APH APC ADI AON APA AIV AAPL AMAT ADM AIZ T ADSK ADP AN AZO AVGO AVB AVY BHI BLL BAC BK BCR BAX BBT BDX BBBY BBY BIIB BLK HRB BA BWA BXP BSX BMY BRCM CHRW CA CVC COG CAM CPB COF CAH HSIC KMX CCL CAT CBG CBS CELG CNP CTL CERN CF SCHW CHK CVX CMG CB CHD CI XEC CINF CTAS CSCO C CTXS CLX CME CMS COH KO CCE CTSH CL CMCSA CMA CAG COP CNX ED STZ GLW COST CCI CSX CMI CVS DHI DHR DRI DVA DE DLPH DAL XRAY DVN DO DFS DISCA DISCK DG DLTR D DOV DOW DPS DTE DD DUK DNB ETFC EMN ETN EBAY ECL EIX EW EA EMC EMR ENDP ESV ETR EOG EQT EFX EQIX EQR ESS EL ES EXC EXPE EXPD ESRX XOM FFIV FB FAST FDX FIS FITB FSLR FE FISV FLIR FLS FLR FMC FTI F BEN FCX FTR GME GPS GRMN GD GE GGP GIS GM GPC GILD GS GT GWW HAL HBI HOG HAR HRS HIG HAS HCA HCP HP HES HD HON HRL HST HPQ HUM HBAN ITW ILMN IR INTC ICE IBM IP IPG IFF INTU ISRG IVZ IRM JEC JBHT JNJ JCI JPM JNPR KSU K KEY GMCR KMB KIM KMI KLAC KSS KR LB LLL LH LRCX LM LEG LEN LVLT LUK LLY LNC LLTC LMT L LOW LYB MTB MAC M MNK MRO MPC MAR MMC MLM MAS MA MAT MKC MCD MHFI MCK MJN MDT MRK MET KORS MCHP MU MSFT MHK TAP MDLZ MON MNST MCO MS MOS MSI MUR MYL NDAQ NOV NAVI NTAP NFLX NWL NFX NEM NWSA NWS NEE NLSN NKE NI NBL JWN NSC NTRS NOC NRG NUE NVDA ORLY OXY OMC OKE ORCL OI PCAR PH PDCO PAYX PNR PBCT POM PEP PKI PRGO PFE PCG PM PSX PNW PXD PBI PCL PNC RL PPG PPL PX PCP PCLN PFG PG PGR PLD PRU PEG PSA PHM PVH QRVO PWR QCOM DGX RRC RTN O RHT REGN RF RSG RAI RHI ROK COL ROP ROST RCL R CRM SNDK SCG SLB SNI STX SEE SRE SHW SIG SPG SWKS SLG SJM SNA SO LUV SWN SE STJ SWK SPLS SBUX HOT STT SRCL SYK STI SYMC SYF SYY TROW TGT TEL TE TGNA THC TDC TSO TXN TXT HSY TRV TMO TIF TWX TWC TJX TMK TSS TSCO RIG TRIP FOXA FOX TSN TYC USB UA UNP UAL UNH UPS URI UTX UHS UNM URBN VFC VLO VAR VTR VRSN VRSK VZ VRTX VIAB V VNO VMC WMT WBA DIS WM WAT ANTM WFC HCN WDC WU WY WHR WFM WMB WLTW WEC WYN WYNN XEL XRX XLNX XL XYL YHOO YUM ZBH ZION ZTS`;

type Result struct {
	Metrics Metrics
	AddedSymbol string
}

func IsParetoOptimal(x Result, others []Result) bool {
	for _, other := range others {
		if other.Metrics.Earnings > x.Metrics.Earnings && other.Metrics.Volatility < x.Metrics.Volatility {
			fmt.Printf("%s has more earnings and less volatility than %s\n", other.AddedSymbol, x.AddedSymbol)
			return false
		}
	}
	fmt.Printf("%s is Pareto-optimal\n", x.AddedSymbol)
	return true
}

func writeParetoOptimalResults(results []Result, path string) {
	paretoOpt := make([]Result, 0)
	for _, result := range(results) {
		if IsParetoOptimal(result, results) {
			paretoOpt = append(paretoOpt, result)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, result := range paretoOpt {
		f.WriteString(fmt.Sprintf("%.3f %.3f %s\n", result.Metrics.Earnings, result.Metrics.Volatility, result.AddedSymbol))
	}
}

func findBestAddedSecurity(priorSymbols []string, priorAllocation []float64, newSecurities []string, amount float64) {
	symbols := append(priorSymbols, newSecurities...)

	showCurrentAllocation()

	results := make([]Result, 0)
	for x, addedSymbol := range symbols {
		weights := make([]float64, len(symbols))
		for j := 0; j < len(priorSymbols); j++ {
			weights[j] = priorAllocation[j]
		}

		fmt.Println(x, addedSymbol)

		// Allocate one new stock.
		weights[x] += amount

		normalize(weights)

		mix := makeMix(symbols, weights)
		metrics := getMetrics(mix)
		results = append(results, Result{
			Metrics: metrics,
			AddedSymbol: addedSymbol,
		})
		// fmt.Println("earnings:", , "volatility:", metrics.Volatility)
	}
	writeParetoOptimalResults(results, "buy-100k.csv")
}

func findOptimalAllocation(symbols []string) {
	results := make([]Result, 0)
	N := 100
	for i := 0; i < N; i++ {
		fmt.Println(i)
		weights := make([]float64, len(symbols))
		for j := 0; j < len(symbols); j++ {
			weights[j] = float64(rand.Intn(1000))
		}
		normalize(weights)
		mix := makeMix(symbols, weights)
		metrics := getMetrics(mix)
		results = append(results, Result{
			Metrics: metrics,
			AddedSymbol: fmt.Sprintf("%v=>%v", symbols, weights),
		})
	}
	writeParetoOptimalResults(results, "sp500-optimal.csv")
	// TODO: initial population: everything=1
}

func main() {
	sp500 := strings.Split(SP500, " ")

	/*
	findOptimalAllocation(sp500)
	*/
	priorSymbols := []string{"AAPL", "ATVI", "BABA", "CAH", "CP", "DIS", "EMN", "GOOG", "HBI", "HRL", "IFF", "IVZ", "IWV", "LOW", "QQQ", "TSLA"}
	priorAllocation := []float64{
		54636.79, // AAPL
		33687.72, // ATVI
		70550.92, // BABA
		47845.69, // CAH
		5541.96, // CP
		46334.51, // DIS
		23669.76, // EMN
		106685.11, // GOOG
		67147.23, // HBI
		65008.47, // HRL
		28472.96, // IFF
		59991.44, // IVZ
		55926.77, // IWV
		61834.45, // LOW
		56643.59, // QQQ
		31544.99, // TSLA
	}
	findBestAddedSecurity(priorSymbols, priorAllocation, sp500, 100000.0)
}
