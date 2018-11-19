package lib

import "time"

//TODO define the basic struct of request and response

type RawReq struct {
	ID  int64
	Req []byte
}

type RawResp struct {
	ID     int64
	Resp   []byte
	Err    error
	Elapse time.Duration //耗时 单位ns
}

type RetCode struct {
}

type CallResult struct {
	ID     int64         //id
	Req    RawReq        //原生请求
	Resp   RawResp       //原生响应
	Code   RetCode       //响应代码
	Msg    string        //原因简述
	Elapse time.Duration //耗时
}

type Caller interface {
	//构建请求
	BuildReq() RawReq
	//调用
	Caller(req []byte, timeoutNs time.Duration) ([]byte, error)
	//检查响应
	CheckResp(req RawReq, resp RawResp) *CallResult
}

//TODO GoTickets like java Thread Pool or POSIX semaphore
type GoTickets interface {
	//获取一张票
	Take()
	//归还
	Return()
	//是否激活
	Active() bool
	//票总数
	Total() uint32
	Remainder() uint32
}

type myGoTickets struct {
	total    uint32        //票的总数
	ticketCh chan struct{} //票的容器
	active   bool          //票池是否被激活
}
