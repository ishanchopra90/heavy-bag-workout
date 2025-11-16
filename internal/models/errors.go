package models

import "errors"

var (
	ErrInvalidWorkDuration = errors.New("work duration must be greater than 0")
	ErrInvalidRestDuration = errors.New("rest duration cannot be negative")
	ErrInvalidTotalRounds  = errors.New("total rounds must be greater than 0")
)
