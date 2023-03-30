package abstraction

import (
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type AppBeb struct {
	state *procstate.ProcState
	queue *queue.Queue
	logg  log.Logger
}

func NewAppBeb(state *procstate.ProcState, queue *queue.Queue, logg log.Logger) *AppBeb {
	return &AppBeb{
		state: state,
		queue: queue,
		logg:  logg,
	}
}

func (appbeb *AppBeb) Handle(msg *pb.Message) {

	if msg == nil {
		appbeb.logg.Error("appbeb handler received nil message")
		return
	}

	switch msg.GetType() {

	// When receiving beb broadcast, forward to the beb broadcast perfect link.
	case pb.Message_BEB_BROADCAST:

		appbeb.logg.Infof("app got broadcast message: %+v", msg)

		for _, proc := range appbeb.state.Processes {
			plsend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: msg.ToAbstractionId,
				ToAbstractionId:   APP_BEB_PL,
				SystemId:          appbeb.state.SystemId,
				MessageUuid:       msg.MessageUuid,
				PlSend: &pb.PlSend{
					Message:     msg.BebBroadcast.Message,
					Destination: proc,
				},
			}
			trigger(appbeb.state, appbeb.queue, &plsend)
		}
	}
}
