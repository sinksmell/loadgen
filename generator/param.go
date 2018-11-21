package generator

import (
	"MyLoadGen/lib"
	"time"
)

type ParamSet struct {
	caller     lib.Caller
	timeoutNS  time.Duration
	lps        uint32
	durationNS time.Duration
	resultCh   chan *lib.CallResult
}
