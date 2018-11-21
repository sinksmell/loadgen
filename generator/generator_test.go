package generator

import (
	"MyLoadGen/lib"
	"testing"
	"time"
)

// printDetail 代表是否打印详细结果。
var printDetail = false

func TestStart(t *testing.T) {

	// 初始化服务器。
	//	server := lib.NewTCPServer()
	//	defer server.Close()
	//	serverAddr := "127.0.0.1:8081"
	//	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	//	err := server.Listen(serverAddr)
	//	if err != nil {
	//		t.Fatalf("TCP Server startup failing! (addr=%s)!\n", serverAddr)
	//		t.FailNow()
	//	}

	serverAddr := "http://127.0.0.1:8080/json"

	// 初始化载荷发生器。
	pset := ParamSet{
		//caller:     lib.NewTCPComm(serverAddr),
		caller:     lib.NewHttpRequest(serverAddr),
		timeoutNS:  50 * time.Millisecond,
		lps:        uint32(1000),
		durationNS: 10 * time.Second,
		resultCh:   make(chan *lib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.timeoutNS, pset.lps, pset.durationNS)
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
	for r := range pset.resultCh {
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
	tps := float64(successCount) / float64(pset.durationNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", pset.lps, tps)

}
