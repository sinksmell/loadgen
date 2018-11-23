package generator

import (
	"MyLoadGen/lib"
	"time"
)

type ParamSet struct {
	Caller     lib.Caller
	TimeoutNS  time.Duration
	Lps        uint32
	DurationNS time.Duration
	ResultCh   chan *lib.CallResult
}

func NewParamSet(caller lib.Caller,
	timeoutNS time.Duration,
	lps uint32,
	durationNS time.Duration,
	resultCh chan *lib.CallResult) ParamSet {

	return ParamSet{
		Caller:     caller,
		TimeoutNS:  timeoutNS,
		Lps:        lps,
		DurationNS: durationNS,
		ResultCh:   resultCh,
	}
}
