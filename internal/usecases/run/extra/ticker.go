package extra

import "time"

type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type tickerWrapper struct {
	ticker *time.Ticker
}

func NewTicker(d time.Duration) Ticker {
	return &tickerWrapper{
		ticker: time.NewTicker(d),
	}
}

func (tw *tickerWrapper) Chan() <-chan time.Time {
	return tw.ticker.C
}

func (tw *tickerWrapper) Stop() {
	tw.ticker.Stop()
}
