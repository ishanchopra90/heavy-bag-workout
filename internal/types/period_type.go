package types

// PeriodType represents the type of period (work or rest)
type PeriodType int

const (
	PeriodWork PeriodType = iota
	PeriodRest
)
