package abstraction

import (
	"fmt"
	pb "hw/protobuf"
	"hw/queue"
	"hw/state"
)

func trigger(state *state.ProcState, queue *queue.Queue, msg *pb.Message) {
	queue.AddWithMsg(msg, *state.Logg, fmt.Sprintf("%s/%s triggers : {%v} \n",
		state.SystemId, state.Name(), msg.ToString()))
}
