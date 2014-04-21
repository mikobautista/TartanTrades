package channelmap

import (
	"github.com/mikobautista/tartantrades/logging"
)

type ChannelMap struct {
	m                 map[interface{}]interface{}
	getChannel        chan getRequest
	putChannel        chan putRequest
	remChannel        chan remRequest
	manipulateChannel chan manipulateRequest
}

type getRequest struct {
	k interface{}
	r chan interface{}
}

type putRequest struct {
	k interface{}
	v interface{}
	r chan interface{}
}

type remRequest struct {
	k interface{}
	r chan interface{}
}

type manipulateRequest struct {
	k interface{}
	f func(interface{}) interface{}
	r chan interface{}
}

var LOG = logging.NewLogger(true)

func NewChannelMap() ChannelMap {
	rv := ChannelMap{
		m:                 make(map[interface{}]interface{}),
		getChannel:        make(chan getRequest, 200),
		putChannel:        make(chan putRequest, 200),
		remChannel:        make(chan remRequest, 200),
		manipulateChannel: make(chan manipulateRequest, 200),
	}
	go rv.listen()
	return rv
}

func (tm *ChannelMap) Get(k interface{}) interface{} {
	r := make(chan interface{})
	tm.getChannel <- getRequest{
		k: k,
		r: r,
	}
	select {
	case v := <-r:
		return v
	}
}

func (tm *ChannelMap) Put(k interface{}, v interface{}) {
	r := make(chan interface{})
	tm.putChannel <- putRequest{
		k: k,
		v: v,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (tm *ChannelMap) Rem(k interface{}) {
	r := make(chan interface{})
	tm.remChannel <- remRequest{
		k: k,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (tm *ChannelMap) ManipulateValue(k interface{}, f func(interface{}) interface{}) {
	r := make(chan interface{})
	tm.manipulateChannel <- manipulateRequest{
		k: k,
		f: f,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (tm *ChannelMap) listen() {
	for {
		select {
		case c := <-tm.getChannel:
			c.r <- tm.m[c.k]
		case c := <-tm.putChannel:
			tm.m[c.k] = c.v
			c.r <- struct{}{}
		case c := <-tm.manipulateChannel:
			v := tm.m[c.k]
			if v != nil {
				tm.m[c.k] = c.f(v)
			}
			c.r <- struct{}{}
		case c := <-tm.remChannel:
			delete(tm.m, c.k)
			c.r <- struct{}{}
		}
	}
}
