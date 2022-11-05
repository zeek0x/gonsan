package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/zeek0x/gonsan"
)

type count struct{}

var _ gonsan.GenServer = (*count)(nil)

func StartMonitor(start int) *gonsan.Process {
	p, _ := gonsan.StartMonitor(&count{}, []any{start}, start)
	return p
}

func (e *count) Init(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Init Error. args: %v", args)
	}
	start := args[0].(int)
	return start, nil
}

func (e *count) HandleCall(request any, state any) (any, any, error) {
	return state, state, nil
}

func (e *count) HandleCast(request any, state any) (any, error) {
	a := state.(int)
	b := request.(int)
	newState := a + b
	return newState, nil
}

func (e *count) HandleInfo(info any, state any) (any, error) {
	fmt.Printf("unknown message: %v\n", info)
	return state, nil
}

func (e *count) Terminate(reason any, state any) {
	fmt.Printf("result: %d\n", state)
}

func main() {
	p := StartMonitor(10)
	fmt.Println(gonsan.Call(p, nil))
	gonsan.Cast(p, 15)
	fmt.Println(gonsan.Call(p, nil))
	p.Mailbox <- "howdy"
	fmt.Println(p.CheckMonitor())
	fmt.Println(gonsan.Call(p, nil))
	gonsan.Cast(p, -5)
	gonsan.Stop(p, errors.New("let's call it a day"))
	time.Sleep(time.Second * 1)
	fmt.Println(p.CheckMonitor())
}
