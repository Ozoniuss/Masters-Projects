package abstraction

import (
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type App struct {
	state *procstate.ProcState
	queue *queue.Queue
	logg  log.Logger
}

func NewApp(state *procstate.ProcState, queue *queue.Queue, logg log.Logger) *App {
	return &App{
		state: state,
		queue: queue,
		logg:  logg,
	}
}

func (app *App) Handle(msg *pb.Message) {
	if msg == nil {
		app.logg.Error("app handler received nil message")
		return
	}
	switch msg.GetType() {

	// When receiving an app_broadcast from the hub, start a beb broadcast.
	case pb.Message_APP_BROADCAST:

		app.logg.Infof("app got broadcast message: %+v", msg)

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
				},
			},
		}

		trigger(app.state, app.queue, &beb)

	case pb.Message_PL_DELIVER:
		app.Handle(msg.PlDeliver.Message)
	}

}
