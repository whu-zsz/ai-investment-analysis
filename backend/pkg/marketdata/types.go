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

type StockDetail struct {
	Symbol         string
	Name           string
	Market         string
	LastPrice      float64
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	PrevClose      float64
	ChangeAmount   float64
	ChangePercent  float64
	Volume         float64
	Turnover       float64
	VolumeRatio    float64
	TurnoverRate   float64
	Amplitude      float64
	LimitUp        float64
	LimitDown      float64
	AveragePrice   float64
	TotalShares    float64
	FloatShares    float64
	TotalMarketCap float64
	FloatMarketCap float64
	Industry       string
	Region         string
	Concepts       []string
	Source         string
	FetchedAt      time.Time
}

type KlineBar struct {
	Symbol        string
	Period        string
	AdjustType    string
	BarTime       time.Time
	Open          float64
	Close         float64
	High          float64
	Low           float64
	Volume        float64
	Amount        float64
	Amplitude     float64
	ChangePercent float64
	ChangeAmount  float64
	TurnoverRate  float64
	Source        string
}
