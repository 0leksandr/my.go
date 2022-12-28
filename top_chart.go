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
		position := len(topChart.Trends)
		for position > 0 {
			if trend.TrendValue() > topChart.Trends[position - 1].TrendValue() {
				position--
			} else {
				break
			}
		}

		if fitsIn { topChart.Trends = append(topChart.Trends, Trend(nil)) }
		copy(topChart.Trends[position+1:], topChart.Trends[position:])
		topChart.Trends[position] = trend
	}
}
