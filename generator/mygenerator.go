package generator

import (
	"MyLoadGen/lib"
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math"
	"sync/atomic"
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

func (mgt *myGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(&mgt.status, STATUS_STARTED, STATUS_STOPPING) {
		return false
	}
	mgt.cancelFunc()
	for {
		if atomic.LoadUint32(&mgt.status) == STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return false
}

func (mgt *myGenerator) Status() uint32 {
	return atomic.LoadUint32(&mgt.status)
}

func (mgt *myGenerator) CallCount() int64 {
	return atomic.LoadInt64(&mgt.callCount)
}

// 初始化载荷发生器。
func (mgt *myGenerator) init() error {

	logger.Informational("Initializing the load generator...")

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

func (set *ParamSet) Check() error {
	return nil
}

func NewGenerator(set ParamSet) (lib.Generator, error) {
	logger.Informational("Initializing a load generator ...")
	if err := set.Check(); err != nil {
		return nil, err
	}

	gen := &myGenerator{
		caller:     set.Caller,
		timeoutNS:  set.TimeoutNS,
		lps:        set.Lps,
		durationNS: set.DurationNS,
		status:     STATUS_ORIGIN,
		resultCh:   set.ResultCh,
	}

	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}

func (mgt *myGenerator) prepareToStop(ctxErr error) {
	logger.Informational("Prepare to stop load generator (case:%s)...", ctxErr)
	atomic.CompareAndSwapUint32(&mgt.status, STATUS_STARTED, STATUS_STOPPING)
	logger.Informational("Closing result channel...")
	close(mgt.resultCh)
	atomic.StoreUint32(&mgt.status, STATUS_STOPPED)
}

func (mgt *myGenerator) genLoad(throttle <-chan time.Time) {

	for {
		select {
		case <-mgt.ctx.Done():
			mgt.prepareToStop(mgt.ctx.Err())
			return
		default:

		}

		if mgt.lps > 0 {
			mgt.asyncCall()
			select {
			case <-throttle:
			case <-mgt.ctx.Done():
				mgt.prepareToStop(mgt.ctx.Err())
				return
			}
		}
	}
}

// asyncSend 会异步地调用承受方接口。
func (mgt *myGenerator) asyncCall() {
	mgt.tickets.Take()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Call Panic! (error: %s)", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic! (clue: %#v)", p)
				}
				logger.Emergency(errMsg)
				result := &lib.CallResult{
					ID:   -1,
					Code: lib.RET_CODE_FATAL_CALL,
					Msg:  errMsg}
				mgt.sendResult(result)
			}
			mgt.tickets.Return()
		}()
		rawReq := mgt.caller.BuildReq()
		// 调用状态：0-未调用或调用中；1-调用完成；2-调用超时。
		var callStatus uint32
		timer := time.AfterFunc(mgt.timeoutNS, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return
			}
			result := &lib.CallResult{
				ID:     rawReq.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_WARNING_CALL_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! (expected: < %v)", mgt.timeoutNS),
				Elapse: mgt.timeoutNS,
			}
			mgt.sendResult(result)
		})
		rawResp := mgt.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			return
		}
		timer.Stop()
		var result *lib.CallResult
		if rawResp.Err != nil {
			result = &lib.CallResult{
				ID:     rawResp.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_ERROR_CALL,
				Msg:    rawResp.Err.Error(),
				Elapse: rawResp.Elapse}
		} else {
			result = mgt.caller.CheckResp(rawReq, *rawResp)
			result.Elapse = rawResp.Elapse
		}
		mgt.sendResult(result)
	}()
}

func (mgt *myGenerator) printIgnoredResult(result *lib.CallResult, cause string) {
	resultMsg := fmt.Sprintf(
		"ID=%d, Code=%d, Msg=%s, Elapse=%v",
		result.ID, result.Code, result.Msg, result.Elapse)
	logger.Warning("Ignored result: %s. (cause: %s)\n", resultMsg, cause)
}

//发送调用结果
func (mgt *myGenerator) sendResult(result *lib.CallResult) bool {
	if atomic.LoadUint32(&mgt.status) != STATUS_STARTED {
		mgt.printIgnoredResult(result, "stopped load generator")
		return false
	}

	select {
	case mgt.resultCh <- result:
		return true
	default:
		mgt.printIgnoredResult(result, "result channel is full")
		return false
	}
}

func (mgt *myGenerator) callOne(req *lib.RawReq) *lib.RawResp {
	atomic.AddInt64(&mgt.callCount, 1)
	if req == nil {
		return &lib.RawResp{
			ID:  -1,
			Err: errors.New("invalid raw request"),
		}
	}

	start := time.Now().UnixNano()
	resp, err := mgt.caller.Caller(req.Req, mgt.timeoutNS)
	end := time.Now().UnixNano()
	elapsedTime := time.Duration(end - start)
	var rawResp lib.RawResp
	if err != nil {
		errMsg := fmt.Sprintf("Sync Call Error:%s.", err)
		rawResp = lib.RawResp{
			ID:     rawResp.ID,
			Err:    errors.New(errMsg),
			Elapse: elapsedTime,
		}

	} else {
		rawResp = lib.RawResp{
			ID:     rawResp.ID,
			Resp:   resp,
			Elapse: elapsedTime,
		}
	}

	return &rawResp
}

// Start 会启动载荷发生器。
func (mgt *myGenerator) Start() bool {
	logger.Informational("Starting load generator...")
	// 检查是否具备可启动的状态，顺便设置状态为正在启动
	if !atomic.CompareAndSwapUint32(
		&mgt.status, STATUS_ORIGIN, STATUS_STARTING) {
		if !atomic.CompareAndSwapUint32(
			&mgt.status, STATUS_STOPPED, STATUS_STARTING) {
			return false
		}
	}

	// 设定节流阀。
	var throttle <-chan time.Time
	if mgt.lps > 0 {
		interval := time.Duration(1e9 / mgt.lps)
		logger.Informational("Setting throttle (%v)...", interval)
		throttle = time.Tick(interval)
	}

	// 初始化上下文和取消函数。
	mgt.ctx, mgt.cancelFunc = context.WithTimeout(
		context.Background(), mgt.durationNS)

	// 初始化调用计数。
	mgt.callCount = 0

	// 设置状态为已启动。
	atomic.StoreUint32(&mgt.status, STATUS_STARTED)

	go func() {
		// 生成并发送载荷。
		logger.Informational("Generating loads...")
		mgt.genLoad(throttle)
		logger.Informational("Stopped. (call count: %d)", mgt.callCount)
	}()
	return true
}
