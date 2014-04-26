package channelvar

import (
	"github.com/mikobautista/tartantrades/logging"
)

type ChannelVar struct {
	v                 interface{}
	getChannel        chan getRequest
	setChannel        chan setRequest
	manipulateChannel chan manipulateRequest
	valNil            bool
}

type getRequest struct {
	r chan interface{}
}

type setRequest struct {
	v interface{}
	r chan interface{}
}

type manipulateRequest struct {
	f func(interface{}) interface{}
	r chan interface{}
}

var LOG = logging.NewLogger(true)

func NewChannelVar() ChannelVar {
	rv := ChannelVar{
		v:                 make(map[interface{}]interface{}),
		valNil:            true,
		getChannel:        make(chan getRequest, 200),
		setChannel:        make(chan setRequest, 200),
		manipulateChannel: make(chan manipulateRequest, 200),
	}
	go rv.listen()
	return rv
}

func (tm *ChannelVar) Get() interface{} {
	r := make(chan interface{})
	tm.getChannel <- getRequest{
		r: r,
	}
	select {
	case v := <-r:
		return v
	}
}

func (tm *ChannelVar) IsNil() bool {
	return tm.valNil
}

func (tm *ChannelVar) Set(v interface{}) {
	r := make(chan interface{})
	tm.setChannel <- setRequest{
		v: v,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (tm *ChannelVar) ManipulateValue(f func(interface{}) interface{}) {
	r := make(chan interface{})
	tm.manipulateChannel <- manipulateRequest{
		f: f,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (tm *ChannelVar) listen() {
	for {
		select {
		case c := <-tm.getChannel:
			c.r <- tm.v
		case c := <-tm.setChannel:
			tm.valNil = c.v == nil
			tm.v = c.v
			c.r <- struct{}{}
		case c := <-tm.manipulateChannel:
			tm.v = c.f(tm.v)
			c.r <- struct{}{}
		}
	}
}
