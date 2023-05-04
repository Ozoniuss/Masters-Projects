package abstraction

import (
	"hw/queue"
	procstate "hw/state"
)

type Nnar struct {
	state *procstate.ProcState
	queue *queue.Queue
}
