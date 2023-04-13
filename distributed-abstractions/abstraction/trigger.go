package abstraction

import (
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	"hw/state"
)

func trigger(state *state.ProcState, queue *queue.Queue, msg *pb.Message) {
	queue.Add(msg)
	log.Printf("%s/%s triggers : {%+v}\n\n", state.SystemId, state.Name(), msg)
}
