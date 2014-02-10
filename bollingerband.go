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
	bestOption    selectionParameters = selectionParameters{symbol: "None"}
  
)

// Used in analize function to try to choose the best investment option
type selectionParameters struct {
  symbol string
  avg float64
}

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

// Compares the last value against the 2 last to try to determine if tend is bottom
func tendBottom(q []yahoofinance.Quote)bool {
	l := len(q)
	return q[l-1].Close < q[l-2].Close && q[l-1].Close < q[l-3].Close
}

// Compares the last three values to try to determine if price is under bottom band
func priceUnderBottom(q []yahoofinance.Quote)bool {
	l := len(q)
	isBottom := true
	for i := 3; i > 0; i-- {
		isBottom = isBottom && q[l-i].Close <= q[l-i].Bottom 
	}
	return isBottom 
}

// Compare the last three values to try to determine if price is close(+/- 10%) to bottom band
func closeToBottom(q []yahoofinance.Quote)bool {
        l := len(q)
        isClose := true
        for i := 3; i > 0; i-- {
                var tenPercent = q[l-i].Bottom * 10 / 100
                var difference = math.Abs(q[l-i].Close - q[l-1].Bottom)
                isClose = isClose && (tenPercent > difference)
        }
	return isClose
}

// Compares the last value against the 2 last to try to determine if tend is up
func tendUp(q []yahoofinance.Quote)bool {
	l := len(q)
	return q[l-1].Close > q[l-2].Close && q[l-1].Close > q[l-3].Close
}

// Compares the last three values to try to determine if price is over top band
func priceOverTop(q []yahoofinance.Quote)bool {
        l := len(q)
        isUp := true
        for i := 3; i > 0; i-- {
                isUp = isUp && q[l-i].Close >= q[l-i].Top
        }
        return isUp
}

// Compare the last three values to try to determine if price is close(+/- 10%) to top band
func closeToTop(q []yahoofinance.Quote)bool {
	l := len(q)
	isClose := true
	for i := 3; i > 0; i-- {
		var tenPercent = q[l-i].Top * 10 / 100
		var difference = math.Abs(q[l-i].Top - q[l-1].Close)
		isClose = isClose && (tenPercent > difference)
	}
	return isClose
}

// Band size average for the last three values
func bandSizeAvg(q []yahoofinance.Quote)float64 {
	l := len(q)
	sum := 0.0
	for i := 3; i > 0; i-- {
		sum += q[l-i].Top - q[l-i].Bottom
	}
	return sum/3.0
}

// Compare the best option with actual and assign the best between both
func setBestOption(symbol string, q []yahoofinance.Quote) {
	avg := bandSizeAvg(q)
	if (bestOption.symbol == "None" || avg > bestOption.avg) {
		bestOption = selectionParameters{symbol: symbol, avg: avg}
	}
}

/*
  analize try to determine what should be the best option to invest using the following rules:
  1. Do not invest if price is under bottom band and prices tend to bottom, since it can mean that prices will keep the trend.
  2. If price is close to up band but tend to bottom then do not invest since the reboot effect.
  3. Invest if price is over top band and price tend to up, since it can mean that prices will keep the trend.
  4. If price is close to bottom band but tend to up then invest since the reboot effect.
*/
func analize() {
	fmt.Printf("Calculating which stock to invest given our current strategy.....\n")
	for _, symbol := range configuration.Symbols {
		dataset := quotes[symbol]
		if (priceUnderBottom(dataset) && tendBottom(dataset)) { //The first rule is reached so we should not invest
			break
		}
		if (priceOverTop(dataset) && tendUp(dataset)) { //The third rule is reached so is a good option invest
			setBestOption(symbol, dataset)
			break
		}
		if (closeToTop(dataset) && tendBottom(dataset)) { //The second rule is reached so we should not invest
			break
		}
		if (closeToBottom(dataset) && tendUp(dataset)) { //The fourth rule is reached so is a good option invest
			setBestOption(symbol, dataset)
			break
		}
	}
	fmt.Printf("You should invest in: %s\n", bestOption.symbol)
}

func main() {
	configuration = Configuration()
	periods := configuration.Periods
	factor := configuration.Factor
	startDate, endDate := dates(periods)
	quotes = make(map[string][]yahoofinance.Quote)
	for _, symbol := range configuration.Symbols {
		fmt.Printf("Calculating Bollinger Band for %s\n", symbol)
		var r = yahoofinance.HistoricalPrices(symbol, startDate, endDate)
		compute(r, periods, factor)
		draw(symbol, r, 8, 4)
		quotes[symbol] = r
	}
	analize()
}
