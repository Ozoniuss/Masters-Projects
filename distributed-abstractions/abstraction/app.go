package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type App struct {
	state *procstate.ProcState
	queue *queue.Queue
}

func NewApp(state *procstate.ProcState, queue *queue.Queue) *App {
	return &App{
		state: state,
		queue: queue,
	}
}

func (app *App) Handle(msg *pb.Message) error {
	if msg == nil {
		return fmt.Errorf("%s handler received nil message", APP)
	}
	log.Printf("[%s got message]: %+v\n\n", APP, msg)
	switch msg.GetType() {

	// Hub sends an APP_WRITE when attempting to write to the register.
	case pb.Message_APP_WRITE:
		registerId := msg.GetAppWrite().GetRegister()
		nnarWrite := pb.Message{
			Type:              pb.Message_NNAR_READ,
			FromAbstractionId: msg.ToAbstractionId,
			ToAbstractionId:   APP_NNAR + "[" + registerId + "]",
			NnarWrite: &pb.NnarWrite{
				Value: msg.GetAppWrite().GetValue(),
			},
		}
		trigger(app.state, app.queue, &nnarWrite)

	// When receiving an app_broadcast from the hub, start a beb broadcast.
	case pb.Message_APP_BROADCAST:

		beb := pb.Message{
			Type:              pb.Message_BEB_BROADCAST,
			FromAbstractionId: msg.ToAbstractionId,
			ToAbstractionId:   APP_BEB,
			SystemId:          app.state.SystemId,
			MessageUuid:       msg.MessageUuid,
			BebBroadcast: &pb.BebBroadcast{
				Message: &pb.Message{
					Type: pb.Message_APP_VALUE,
					AppValue: &pb.AppValue{
						Value: msg.AppBroadcast.Value,
					},

					// ???
					FromAbstractionId: "app",
					ToAbstractionId:   "app",
				},
			},
		}

		trigger(app.state, app.queue, &beb)

	case pb.Message_BEB_DELIVER:

		// Send the message back to hub.
		plsend := pb.Message{
			Type:              pb.Message_PL_SEND,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   APP_PL,
			SystemId:          app.state.SystemId,
			MessageUuid:       msg.GetMessageUuid(),
			PlSend: &pb.PlSend{
				Message: msg.GetBebDeliver().GetMessage(),
				Destination: &pb.ProcessId{
					Host:  "127.0.0.1",
					Port:  5000,
					Owner: "hub",
				},
			},
		}
		trigger(app.state, app.queue, &plsend)

	case pb.Message_PL_DELIVER:
		return app.Handle(msg.PlDeliver.Message)
	}

	return nil
}
