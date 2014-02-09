package main

// Using plotinum to genrate graphics
import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"fmt"
	"github.com/Aguardientico/yahoofinance"
	"math"
	"time"
)

const (
	GRAPHIC_TITLE = "%s Bollinger Band"
	X_AXIS_LABEL  = "Days"
	Y_AXIS_LABEL  = "Price"
	TOP_LABEL     = "Top"
	PRICE_LABEL   = "Price"
	BOTTOM_LABEL  = "Bottom"
)

var (
	configuration ConfigurationInfo
	quotes        map[string][]yahoofinance.Quote
)

/*
  dates function returns start and end dates to retreive quoute's info.

  Quotes are set only for working days so we need
  to get enough days to process all periods.
  Also to obtain a better approach we need to get double
  length quotes to be more accurate.
*/
func dates(periods int) (startDate, endDate time.Time) {
	var realPeriods int = periods * 2 // Duplicate periods length
	endDate = time.Now()
	for endDate.Weekday() == 6 || endDate.Weekday() == 0 { // If endDate is weekend then we should get the previous working day
		endDate = endDate.Add(time.Hour * -24)
	}
	var duration int = 0
	for startDate = endDate.Add(time.Hour * -24); duration < realPeriods; startDate = startDate.Add(time.Hour * -24) {
		if startDate.Weekday() != 6 && startDate.Weekday() != 0 { //We should only take into account working days
			duration = duration + 1
		}
	}
	startDate = startDate.Add(time.Hour * 24) //Since for behavior is substracting 1 day before verify condition, then we should add 1 day to get the right startDate
	return
}

/*
  Compute moving average, standard deviation and the bollinger upper and lower bands
  This function uses the approach fount at:
  http://stackoverflow.com/questions/14635735/how-to-efficiently-calculate-a-moving-standard-deviation
*/
func compute(quotes []yahoofinance.Quote, period int, factor float64) {
	totalAvg := 0.0
	totalSqr := 0.0

	for i, q := range quotes {
		totalAvg = totalAvg + q.Close
		totalSqr = totalSqr + math.Pow(q.Close, 2)

		if i >= period-1 {
			avg := totalAvg / float64(period)

			stdev := math.Sqrt((totalSqr - math.Pow(totalAvg, 2)/float64(period)) / float64(period))
			quotes[i].Avg = avg
			quotes[i].Top = avg + factor*stdev
			quotes[i].Bottom = avg - factor*stdev

			totalAvg = totalAvg - quotes[i-period+1].Close
			totalSqr = totalSqr - math.Pow(quotes[i-period+1].Close, 2)
		}
	}
}

/*
  draw generates pgn file using a quotes array
  Internally this function uses plotinum:
  http://code.google.com/p/plotinum/
*/
func draw(symbol string, quotes []yahoofinance.Quote, width, height float64) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = fmt.Sprintf(GRAPHIC_TITLE, symbol)
	p.X.Label.Text = X_AXIS_LABEL
	p.Y.Label.Text = Y_AXIS_LABEL

	// Hack to remove no calculated/used records
	temp := make([]yahoofinance.Quote, 0)
	for _, q := range quotes {
		if q.Top != 0 && q.Avg != 0 && q.Bottom != 0 {
			temp = append(temp, q)
		}
	}

	length := len(temp)
	pts1, pts2, pts3 := make(plotter.XYs, length), make(plotter.XYs, length), make(plotter.XYs, length)
	for i, q := range temp {
		j := float64(i)
		pts1[i].X = j
		pts1[i].Y = q.Top
		pts2[i].X = j
		pts2[i].Y = q.Close
		pts3[i].X = j
		pts3[i].Y = q.Bottom
	}

	if err := plotutil.AddLinePoints(p, TOP_LABEL, pts1, PRICE_LABEL, pts2, BOTTOM_LABEL, pts3); err != nil {
		panic(err)
	}

	if err := p.Save(width, height, symbol+".png"); err != nil {
		panic(err)
	}
}

func main() {
	configuration = Configuration()
	periods := configuration.Periods
	factor := configuration.Factor
	startDate, endDate := dates(periods)
	for _, symbol := range configuration.Symbols {
		fmt.Printf("Calculating Bollinger Band for %s\n", symbol)
		var r = yahoofinance.HistoricalPrices(symbol, startDate, endDate)
		compute(r, periods, factor)
		draw(symbol, r, 8, 4)
	}
}
