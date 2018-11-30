package main

import (
	"MyLoadGen/generator"
	"MyLoadGen/lib"
	"MyLoadGen/request"
	"flag"
	"github.com/astaxie/beego"
	"time"
)

// printDetail 代表是否打印详细结果。
var printDetail = false
var url = flag.String("url", "http://127.0.0.1:8080", "测试地址")
var lps = flag.Int("lps", 1000, "每秒载荷发送量")
var timeOut = flag.String("timeOut", "1000ms", "响应超时时间(单位: ms,s 等)")
var tm = flag.String("t", "10s", "测试时长(单位: s)")

func main() {

	flag.Parse()
	logger := beego.BeeLogger
	//serverAddr := "http://www.baidu.com"
	// 初始化载荷发生器。
	tout, err := time.ParseDuration(*timeOut)

	if err != nil {
		logger.Error("%s\n", err)
		return
	}
	tm, err := time.ParseDuration(*tm)
	if err != nil {
		logger.Error("%s\n", err)
		return
	}
	pset := generator.NewParamSet(
		//request.NewTCPComm(serverAddr),/*TCP 请求*/
		//request.NewPostRequest(serverAddr), //post 请求
		request.NewGetRequest(*url), /*get请求*/
		tout,
		uint32(*lps),
		tm,
		make(chan *lib.CallResult, 100))

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
		logger.Info("  Code plain: %s (%04d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	logger.Info("Total: %d.\n", total)
	successCount := countMap[lib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(pset.DurationNS/1e9)
	logger.Info("Loads per second: %d; Treatments per second: %f; Success Rate: %f.\n", pset.Lps, tps, tps/(float64(pset.Lps)))

}
