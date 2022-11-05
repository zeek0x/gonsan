package gonsan

type Process struct {
	Mailbox chan any
	monitor chan any
}

func monitor(p *Process) {
	p.monitor = make(chan any)
}

func (p *Process) CheckMonitor() (any, bool) {
	select {
	case reason := <-p.monitor:
		return reason, true
	default:
		return nil, false
	}
}
