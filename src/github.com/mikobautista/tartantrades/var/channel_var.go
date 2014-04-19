package tvar

import (
    "github.com/cmu440/tribbler/logger"
)

type Tvar struct {
    v                 interface{}
    getChannel        chan getRequest
    setChannel        chan setRequest
    manipulateChannel chan manipulateRequest
}

type getRequest struct {
    r chan interface{}
}

type setRequest struct {
    v   interface{}
    r   chan interface{}
}

type manipulateRequest struct {
    f   func(interface{}) interface{}
    r   chan interface{}
}

var LOG = logger.NewLogger(true)

func NewTvar() Tvar {
    rv := Tvar{
        v:                 make(map[interface{}]interface{}),
        getChannel:        make(chan getRequest, 200),
        setChannel:        make(chan setRequest, 200),
        manipulateChannel: make(chan manipulateRequest, 200),
    }
    go rv.listen()
    return rv
}

func (tm *Tvar) Get() interface{} {
    r := make(chan interface{})
    tm.getChannel <- getRequest{
        r: r,
    }
    select {
    case v := <-r:
        return v
    }
}

func (tm *Tvar) Set(v interface{}) {
    r := make(chan interface{})
    tm.setChannel <- setRequest{
        v:  v,
        r:  r,
    }
    select {
    case _ = <-r:
        return
    }
}

func (tm *Tvar) ManipulateValue(f func(interface{}) interface{}) {
    r := make(chan interface{})
    tm.manipulateChannel <- manipulateRequest{
        f:  f,
        r:  r,
    }
    select {
    case _ = <-r:
        return
    }
}

func (tm *Tvar) listen() {
    for {
        select {
        case c := <-tm.getChannel:
            c.r <- tm.v
        case c := <-tm.setChannel:
            tm.v = c.v
            c.r <- struct{}{}
        case c := <-tm.manipulateChannel:
            tm.v = c.f(tm.v)
            c.r <- struct{}{}
        }
    }
}
