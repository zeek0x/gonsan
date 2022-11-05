package gonsan

type GenServer interface {
	Init(args ...any) (any, error)
	HandleCall(request any, state any) (reply any, newState any, err error)
	HandleCast(request any, state any) (newState any, err error)
	HandleInfo(info any, state any) (newState any, err error)
	Terminate(reason any, state any)
}

type command interface {
	isCommand()
}

type call struct {
	request any
	reply   chan any
}

func (*call) isCommand() {}

var _ command = (*call)(nil)

type cast struct{ request any }

func (*cast) isCommand() {}

var _ command = (*cast)(nil)

type stop struct{ err error }

func (*stop) isCommand() {}

var _ command = (*stop)(nil)

func Start(callback GenServer, args []any, options any) (*Process, error) {
	type startRet struct {
		process *Process
		err     error
	}
	ch := make(chan *startRet)
	defer close(ch)
	go func() {
		state, err := callback.Init(args...)
		if err != nil {
			ch <- &startRet{process: nil, err: err}
		}
		mailbox := make(chan any, 100)
		defer close(mailbox)
		process := &Process{Mailbox: mailbox}
		ch <- &startRet{process: process, err: nil}
		err = loop(process, callback, state)
		if process.monitor != nil {
			process.monitor <- err
		}
	}()
	ret := <-ch
	return ret.process, ret.err
}

func StartMonitor(callback GenServer, args []any, options any) (*Process, error) {
	process, err := Start(callback, args, options)
	if err != nil {
		return nil, err
	}
	monitor(process)
	return process, err
}

func Call(p *Process, request any) any {
	reply := make(chan any)
	defer close(reply)
	p.Mailbox <- &call{request: request, reply: reply}
	return <-reply
}

func Cast(p *Process, request any) {
	p.Mailbox <- &cast{request: request}
}

func Stop(p *Process, err error) {
	p.Mailbox <- &stop{err: err}
}

func loop(p *Process, callback GenServer, state any) (err error) {
	switch message := (<-p.Mailbox).(type) {
	case *call:
		var reply any
		reply, state, err = callback.HandleCall(message.request, state)
		message.reply <- reply
		if err != nil {
			return err
		}
	case *cast:
		state, err = callback.HandleCast(message.request, state)
		if err != nil {
			return err
		}
	case *stop:
		callback.Terminate(message.err, state)
		return message.err
	default:
		state, err = callback.HandleInfo(message, state)
		if err != nil {
			return err
		}
	}
	return loop(p, callback, state)
}
