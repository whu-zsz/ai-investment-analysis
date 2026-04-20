package marketdata

import "time"

type Quote struct {
	Symbol        string
	Name          string
	Market        string
	SnapshotTime  time.Time
	LastPrice     float64
	ChangeAmount  float64
	ChangePercent float64
	OpenPrice     float64
	HighPrice     float64
	LowPrice      float64
	PrevClose     float64
	Volume        float64
	Turnover      float64
	Source        string
}
