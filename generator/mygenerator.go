package generator

import (
	"MyLoadGen/lib"
	"time"
	"context"
)

//TODO define the struct of generator and it`s status

const (
	STATUS_ORIGIN   uint32 = 0 //原始状态
	STATUS_STARTING uint32 = 1 //正在启动
	STATUS_STARTED  uint32 = 2 //已经启动
	STATUS_STOPPING uint32 = 3 //正在关闭
	STATUS_STOPPED  uint32 = 4 //已经关闭

)

type myGenerator struct {
	caller      lib.Caller           //调用器
	timeoutNS   time.Duration        //超时时间 ns
	lps         uint32               //每秒载荷量
	durationNS  time.Duration        //负载持续时间 ns
	concurrency uint32               //载荷并发量
	tickets     lib.GoTickets        //goroutine 票池
	ctx         context.Context      //上下文
	cancelFunc  context.CancelFunc   //取消函数
	callCount   int64                //调用计数
	status      uint32               //状态
	resultCh    chan *lib.CallResult //调用结果通道
}

func NewGenerator(
	caller lib.Caller,
	timeoutNS time.Duration,
	lps uint32,
	duration time.Duration,
	resultCh chan *lib.CallResult) (lib.Generator, error) {

}
