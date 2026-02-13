package bankstatement

import "time"

type ParsedItem struct {
	Date        time.Time
	Description string
	Debit       float64
	Credit      float64
	Balance     float64
}

type ParseResult struct {
	Items       []ParsedItem
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}

type Parser interface {
	Parse(text string) (*ParseResult, error)
}
