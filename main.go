package main

import (
	"MyLoadGen/generator"
	"MyLoadGen/lib"
	"MyLoadGen/request"
	"github.com/astaxie/beego"
	"time"
)

// printDetail 代表是否打印详细结果。
var printDetail = true

func main() {

	logger := beego.BeeLogger
	serverAddr := "http://127.0.0.1:8080/json"

	// 初始化载荷发生器。

	pset := generator.NewParamSet(request.NewHttpRequest(serverAddr), 50*time.Millisecond,
		uint32(1000), 10*time.Second, make(chan *lib.CallResult, 50))

	logger.Info("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.Lps, pset.DurationNS)
	gen, err := generator.NewGenerator(pset)
	if err != nil {
		logger.Warn("Load generator initialization failing: %s\n",
			err)
		return
	}

	// 开始！
	logger.Info("Start load generator...")
	gen.Start()

	// 显示结果。
	countMap := make(map[lib.RetCode]int)
	for r := range pset.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			logger.Info("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	logger.Info("RetCode Count:")
	for k, v := range countMap {
		codePlain := lib.GetRetCodePlain(k)
		logger.Info("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	logger.Info("Total: %d.\n", total)
	successCount := countMap[lib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(pset.DurationNS/1e9)
	logger.Info("Loads per second: %d; Treatments per second: %f; Rate: %f.\n", pset.Lps, tps, tps/(float64(pset.Lps)))

}
