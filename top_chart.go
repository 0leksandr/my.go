package my

type Trend interface {
	TrendValue() float64
}

// TopChart Structure, that holds "top" elements (`Trend`s), inserted into it
type TopChart struct {
	fitsIn func([]Trend, Trend) bool
	Trends []Trend
}
func (*TopChart) New(fitsIn func([]Trend, Trend) bool) *TopChart {
	return &TopChart{
		fitsIn: fitsIn,
		Trends: make([]Trend, 0),
	}
}
func (*TopChart) OfConstSize(size uint) *TopChart {
	return (*TopChart)(nil).New(
		func(trends []Trend, _ Trend) bool {
			return len(trends) < int(size)
		},
	)
}
func (topChart *TopChart) Add(trend Trend) {
	fitsIn := topChart.fitsIn(topChart.Trends, trend)
	l := len(topChart.Trends)
	if fitsIn || l == 0 || trend.TrendValue() > topChart.Trends[l - 1].TrendValue() {
		position := 0
		for _, existingTrend := range topChart.Trends {
			if trend.TrendValue() <= existingTrend.TrendValue() {
				position++
			} else {
				break
			}
		}

		newLen := len(topChart.Trends)
		if fitsIn { newLen++ }
		trends := make([]Trend, 0, newLen)
		trends = append(trends, topChart.Trends[:position]...)
		trends = append(trends, trend)
		trends = append(trends, topChart.Trends[position : newLen-1]...)
		topChart.Trends = trends
	}
}
