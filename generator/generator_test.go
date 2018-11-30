package generator

import (
	"MyLoadGen/lib"
	"MyLoadGen/request"
	"testing"
	"time"
)

// printDetail 代表是否打印详细结果。
var printDetail = false

func TestStart(t *testing.T) {

	serverAddr := "http://127.0.0.1:8080"

	// 初始化载荷发生器。
	pset := ParamSet{
		//Caller:     lib.NewTCPComm(serverAddr),
		//Caller:     request.NewPostRequest(serverAddr),
		Caller:     request.NewGetRequest(serverAddr),
		TimeoutNS:  500 * time.Millisecond,
		Lps:        uint32(1000),
		DurationNS: 10 * time.Second,
		ResultCh:   make(chan *lib.CallResult, 500),
	}
	t.Logf("Initialize load generator (TimeoutNS=%v, Lps=%d, DurationNS=%v)...",
		pset.TimeoutNS, pset.Lps, pset.DurationNS)
	gen, err := NewGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s\n",
			err)
		t.FailNow()
	}

	// 开始！
	t.Log("Start load generator...")
	gen.Start()

	// 显示结果。
	countMap := make(map[lib.RetCode]int)
	for r := range pset.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	t.Log("RetCode Count:")
	for k, v := range countMap {
		codePlain := lib.GetRetCodePlain(k)
		t.Logf("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	t.Logf("Total: %d.\n", total)
	successCount := countMap[lib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(pset.DurationNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", pset.Lps, tps)

}
