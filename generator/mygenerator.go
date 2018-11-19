package generator

import (
	"MyLoadGen/lib"
	"context"
	"github.com/astaxie/beego/logs"
	"math"
	"time"
)

var logger = logs.NewLogger(10000) // 创建一个日志记录器，参数为缓冲区的大小

func init() {
	logger.SetLogger("console", "")  // 设置日志记录方式：控制台记录
	logger.SetLevel(logs.LevelDebug) // 设置日志写入缓冲区的等级：Debug级别（最低级别，所以所有log都会输入到缓冲区）
	logger.EnableFuncCallDepth(true) // 输出log时能显示输出文件名和行号（非必须）
}

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

func (mgt *myGenerator) Start() bool {
	panic("implement me")
}

func (mgt *myGenerator) Stop() bool {
	panic("implement me")
}

func (mgt *myGenerator) Status() uint32 {
	panic("implement me")
}

func (mgt *myGenerator) CallCount() int64 {
	panic("implement me")
}

// 初始化载荷发生器。
func (mgt *myGenerator) init() error {
	//	var buf bytes.Buffer
	logger.Informational("Initializing the load generator...")
	//	buf.WriteString("Initializing the load generator...")
	// 载荷的并发量 ≈ 载荷的响应超时时间 / 载荷的发送间隔时间
	var total64 = int64(mgt.timeoutNS)/int64(1e9/mgt.lps) + 1
	if total64 > math.MaxInt32 {
		total64 = math.MaxInt32
	}
	mgt.concurrency = uint32(total64)
	tickets, err := lib.NewGoTickets(mgt.concurrency)
	if err != nil {
		return err
	}
	mgt.tickets = tickets

	//	buf.WriteString(fmt.Sprintf("Done. (concurrency=%d)", mgt.concurrency))
	logger.Informational("Done. (concurrency=%d)", mgt.concurrency)
	return nil
}

type ParamSet struct {
	caller     lib.Caller
	timeoutNS  time.Duration
	lps        uint32
	durationNS time.Duration
	resultCh   chan *lib.CallResult
}

func (set *ParamSet) Check() error {
	return nil
}

func NewGenerator(set ParamSet) (lib.Generator, error) {
	logger.Informational("Initializing a load generator ...")
	if err := set.Check(); err != nil {
		return nil, err
	}

	gen := &myGenerator{
		caller:     set.caller,
		timeoutNS:  set.timeoutNS,
		lps:        set.lps,
		durationNS: set.durationNS,
		status:     STATUS_ORIGIN,
		resultCh:   set.resultCh,
	}

	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}
