package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type App struct {
	abstractions  *map[string]Abstraction
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	stopq         chan struct{}
}

func NewApp(state *procstate.ProcState, queue *queue.Queue, abstractions *map[string]Abstraction, stopq chan struct{}) *App {
	app := &App{
		state:         state,
		queue:         queue,
		abstractions:  abstractions,
		abstractionId: "app",
		stopq:         stopq,
	}

	beb := NewBeb(app.state, app.queue, app.abstractionId+".beb")
	pl := NewPl(app.state, app.queue, app.abstractionId+".pl", abstractions)
	bebpl := NewPl(app.state, app.queue, app.abstractionId+".beb.pl", abstractions)

	// Register all abstractions that are related to the app.
	RegisterAbstraction(app.abstractions, beb.abstractionId, beb)
	RegisterAbstraction(app.abstractions, pl.abstractionId, pl)
	RegisterAbstraction(app.abstractions, bebpl.abstractionId, bebpl)

	return app
}

func (app *App) Handle(msg *pb.Message) error {
	if msg == nil {
		return fmt.Errorf("%s handler received nil message", APP)
	}
	log.Printf("[%s got message]: {%+v}\n\n", APP, msg)
	switch msg.GetType() {

	case pb.Message_PROC_INITIALIZE_SYSTEM:
		app.state.SystemId = msg.SystemId
		app.state.Processes = make([]*pb.ProcessId, 0, len(msg.GetProcInitializeSystem().GetProcesses()))

		for _, pid := range msg.GetProcInitializeSystem().GetProcesses() {
			// Do not register the hub to the process list.
			if pid.Owner == "hub" {
				continue
			}
			app.state.Processes = append(app.state.Processes, pid)
			if pid.Host == app.state.Host && pid.Port == int32(app.state.Port) {
				if app.state.CurrentProcId != nil {
					panic("current process already identified")
				}
				app.state.CurrentProcId = pid
			}
		}

		log.Printf("[%s] system id: %s\n", APP, app.state.SystemId)
		log.Printf("[%s] current process: {%+v}\n", APP, app.state.CurrentProcId)

		for _, p := range app.state.Processes {
			log.Printf("[%s] process %s-%d: {%+v}\n", APP, p.Owner, p.Index, p)
		}
		fmt.Println()

		fmt.Printf("[%s] Initialization Complete.\n\n", APP)

	case pb.Message_PROC_DESTROY_SYSTEM:
		// Stop the queue from processing.
		app.stopq <- struct{}{}

		// Clear all abstractions and the queue.
		newAbstractions := make(map[string]Abstraction)

		// Initialize new abstractions. No need for hthe handshake now.
		app.abstractions = &newAbstractions
		app.state.CurrentProcId = nil
		app.state.Processes = nil

		// Register the app abstraction.
		RegisterAbstraction(app.abstractions, "app", app)

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

	case pb.Message_UC_DECIDE:
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
					Type:              pb.Message_APP_DECIDE,
					FromAbstractionId: msg.GetToAbstractionId(),
					ToAbstractionId:   msg.GetToAbstractionId(),
					AppDecide: &pb.AppDecide{
						Value: msg.GetUcDecide().GetValue(),
					},
				},
			},
		}
		trigger(app.state, app.queue, &plSend)

	case pb.Message_APP_PROPOSE:
		ucid := fmt.Sprintf("app.uc[%s]", msg.GetAppPropose().GetTopic())
		uc := NewUc(app.state, app.queue, app.abstractions, ucid)
		RegisterAbstraction(app.abstractions, ucid, uc)

		ucPropose := pb.Message{
			Type:              pb.Message_UC_PROPOSE,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   ucid,
			UcPropose: &pb.UcPropose{
				Value: msg.GetAppPropose().GetValue(),
			},
		}
		trigger(app.state, app.queue, &ucPropose)

	case pb.Message_PL_DELIVER:
		return app.Handle(msg.GetPlDeliver().GetMessage())
	}

	return nil
}
