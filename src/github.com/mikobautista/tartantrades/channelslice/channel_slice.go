package channelslice

import (
	"github.com/mikobautista/tartantrades/logging"
)

type ChannelSlice struct {
	s                    []interface{}
	getChannel           chan getRequest
	appendChannel        chan appendRequest
	remChannel           chan remRequest
	flushChannel         chan flushRequest
	containsChannel      chan containsRequest
	getStringListChannel chan getStringListRequest
}

type getRequest struct {
	r chan []interface{}
}

type appendRequest struct {
	v interface{}
	r chan interface{}
}

type remRequest struct {
	v interface{}
	r chan interface{}
}

type flushRequest struct {
	r chan interface{}
}

type containsRequest struct {
	v interface{}
	r chan bool
}

type getStringListRequest struct {
	r chan []string
}

var LOG = logging.NewLogger(false)

func NewChannelSlice() ChannelSlice {
	rv := ChannelSlice{
		s:                    make([]interface{}, 0),
		getChannel:           make(chan getRequest, 200),
		appendChannel:        make(chan appendRequest, 200),
		remChannel:           make(chan remRequest, 200),
		containsChannel:      make(chan containsRequest, 200),
		getStringListChannel: make(chan getStringListRequest, 200),
	}
	go rv.listen()
	return rv
}

func (ts *ChannelSlice) Get() interface{} {
	r := make(chan []interface{})
	ts.getChannel <- getRequest{
		r: r,
	}
	select {
	case v := <-r:
		return v
	}
}

func (ts *ChannelSlice) Size() int {
	x := ts.Get().([]interface{})
	return len(x)
}

func (ts *ChannelSlice) Contains(v interface{}) bool {
	r := make(chan bool)
	ts.containsChannel <- containsRequest{
		v: v,
		r: r,
	}
	select {
	case v := <-r:
		return v
	}
}

func (ts *ChannelSlice) Append(v interface{}) {
	r := make(chan interface{})
	ts.appendChannel <- appendRequest{
		v: v,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (ts *ChannelSlice) Rem(v interface{}) {
	r := make(chan interface{})
	ts.remChannel <- remRequest{
		v: v,
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (ts *ChannelSlice) Flush() {
	r := make(chan interface{})
	ts.flushChannel <- flushRequest{
		r: r,
	}
	select {
	case _ = <-r:
		return
	}
}

func (ts *ChannelSlice) GetStringList() []string {
	r := make(chan []string)
	ts.getStringListChannel <- getStringListRequest{
		r: r,
	}
	select {
	case v := <-r:
		return v
	}
}

func (ts *ChannelSlice) listen() {
	for {
		select {
		case c := <-ts.getChannel:
			c.r <- ts.s
		case c := <-ts.appendChannel:
			ts.s = append(ts.s, c.v)
			c.r <- struct{}{}
		case c := <-ts.remChannel:
			ts.s = remove(ts.s, c.v)
			c.r <- struct{}{}
		case c := <-ts.flushChannel:
			ts.s = nil
			c.r <- struct{}{}
		case c := <-ts.containsChannel:
			c.r <- sliceContains(ts.s, c.v)
		case c := <-ts.getStringListChannel:
			c.r <- convertToStringList(ts.s)
		}
	}
}

//------------------------------------------
//           Helper Functions
//------------------------------------------

func remove(list []interface{}, target interface{}) []interface{} {
	for i := 0; i < len(list); i++ {
		if list[i] == target {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return list
}

func convertToStringList(list []interface{}) []string {
	res := make([]string, 0)
	for i := 0; i < len(list); i++ {
		res = append(res, list[i].(string))
	}
	return res
}

func sliceContains(t []interface{}, v interface{}) bool {
	for _, b := range t {
		if b == v {
			return true
		}
	}
	return false
}
