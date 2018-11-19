package lib

import (
	"errors"
	"fmt"
)

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
	//剩余的票数
	Remainder() uint32
}

type myGoTickets struct {
	total    uint32        //票的总数
	ticketCh chan struct{} //票的容器
	active   bool          //票池是否被激活
}

func (gt *myGoTickets) Take() {
	<-gt.ticketCh
}

func (gt *myGoTickets) Return() {
	gt.ticketCh <- struct{}{}
}

func (gt *myGoTickets) Active() bool {
	return gt.active
}

func (gt *myGoTickets) Total() uint32 {
	return gt.total
}

func (gt *myGoTickets) Remainder() uint32 {
	return uint32(len(gt.ticketCh))
}

func (gt *myGoTickets) init(total uint32) bool {
	if gt.active {
		return false
	}

	if total == 0 {
		return false
	}
	ch := make(chan struct{}, total)
	n := int(total)
	for i := 0; i < n; i++ {
		ch <- struct{}{}
	}
	gt.ticketCh = ch
	gt.total = total
	gt.active = true
	return true
}

func NewGoTickets(total uint32) (GoTickets, error) {
	gt := myGoTickets{}
	if !gt.init(total) {
		errMsg := fmt.Sprintf("The gorountine ticket pool inititalized failure! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	return &gt, nil
}
