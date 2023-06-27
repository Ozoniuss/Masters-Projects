package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type App struct {
	abstractions *map[string]Abstraction
	state        *procstate.ProcState
	queue        *queue.Queue
}

func NewApp(state *procstate.ProcState, queue *queue.Queue, abstractions *map[string]Abstraction) *App {
	return &App{
		state:        state,
		queue:        queue,
		abstractions: abstractions,
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

		registerId := APP_NNAR + "[" + msg.GetAppWrite().GetRegister() + "]"

		// Must create the nnar abstraction with the given name, if it doesn't
		// exist already.
		if _, ok := (*app.abstractions)[registerId]; !ok {
			nnar := NewNnar(app.state, app.queue, registerId)
			RegisterNnar(app.abstractions, registerId, nnar, app.state, app.queue)
		}

		nnarWrite := pb.Message{
			Type:              pb.Message_NNAR_WRITE,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   registerId,
			NnarWrite: &pb.NnarWrite{
				Value: msg.GetAppWrite().GetValue(),
			},
		}
		trigger(app.state, app.queue, &nnarWrite)

		// Hub sends an APP_WRITE when attempting to write to the register.
	case pb.Message_APP_READ:

		registerId := APP_NNAR + "[" + msg.GetAppRead().GetRegister() + "]"

		// Must create the nnar abstraction with the given name, if it doesn't
		// exist already.
		if _, ok := (*app.abstractions)[registerId]; !ok {
			nnar := NewNnar(app.state, app.queue, registerId)
			RegisterNnar(app.abstractions, registerId, nnar, app.state, app.queue)
		}

		nnarRead := pb.Message{
			Type:              pb.Message_NNAR_READ,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   registerId,
			NnarRead:          &pb.NnarRead{},
		}
		trigger(app.state, app.queue, &nnarRead)

	// These messages come from app.nnar[x] and we have to retrieve x.
	case pb.Message_NNAR_READ_RETURN:
		regsterId := msg.GetFromAbstractionId()[9 : len(msg.GetFromAbstractionId())-1]
		plSend := pb.Message{
			Type:              pb.Message_PL_SEND,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Next(msg.GetToAbstractionId(), "pl"),
			PlSend: &pb.PlSend{
				Destination: &pb.ProcessId{
					Host:  "127.0.0.1",
					Port:  5000,
					Owner: "hub",
				},
				Message: &pb.Message{
					Type:              pb.Message_APP_READ_RETURN,
					FromAbstractionId: msg.GetToAbstractionId(),
					ToAbstractionId:   "hub",
					AppReadReturn: &pb.AppReadReturn{
						Register: regsterId,
						Value:    msg.GetNnarReadReturn().GetValue(),
					},
				},
			},
		}
		trigger(app.state, app.queue, &plSend)

	case pb.Message_NNAR_WRITE_RETURN:
		regsterId := msg.GetFromAbstractionId()[9 : len(msg.GetFromAbstractionId())-1]
		plSend := pb.Message{
			Type:              pb.Message_PL_SEND,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Next(msg.GetToAbstractionId(), "pl"),
			PlSend: &pb.PlSend{
				Destination: &pb.ProcessId{
					Host:  "127.0.0.1",
					Port:  5000,
					Owner: "hub",
				},
				Message: &pb.Message{
					Type:              pb.Message_APP_WRITE_RETURN,
					FromAbstractionId: msg.GetToAbstractionId(),
					ToAbstractionId:   "hub",
					AppWriteReturn: &pb.AppWriteReturn{
						Register: regsterId,
					},
				},
			},
		}
		trigger(app.state, app.queue, &plSend)

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
		return app.Handle(msg.GetPlDeliver().GetMessage())
	}

	return nil
}
