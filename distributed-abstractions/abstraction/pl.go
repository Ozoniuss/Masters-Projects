package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"net"
)

const MAX_MSG_SIZE = 256 * 256 * 256 * 256

// Pl acts both as a perfect link for receiving messages from the hub, as well
// as the other processes.
type Pl struct {
	state *procstate.ProcState
	queue *queue.Queue
	logg  *log.Logger
}

func NewPl(state *procstate.ProcState, queue *queue.Queue, logg *log.Logger) *Pl {
	return &Pl{
		state: state,
		queue: queue,
		logg:  logg,
	}
}

func (app *Pl) Handle(msg *pb.Message) {
	switch msg.GetType() {

	// App.pl is what actually receives network messages.
	case pb.Message_NETWORK_MESSAGE:

		app.logg.Infof("got network message: %+v", msg)
		app.logg.Debugf("%s/%s <- %s : %v", app.state.SystemId, app.state.Name(), msg.FromAbstractionId)
	}

}

func (app_pl *Pl) ReadSocket(lis net.Listener) error {

	for {

		var client net.Conn
		// var err error
		// var gotClient = make(chan struct{}, 1)

		// go func() {
		// 	// Accept incoming messages.
		// 	client, err = lis.Accept()
		// 	if err != nil {
		// 		app_pl.logg.Errorf("could not accept incoming connection: %s", err.Error())
		// 		app_pl.state.Quit <- struct{}{}
		// 		//panic(err)
		// 	} else {
		// 		gotClient <- struct{}{}
		// 	}
		// }()

		// select {
		// case <-app_pl.state.Quit:
		// 	break
		// case <-gotClient:
		// }

		client, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		defer client.Close()
		var received = make([]byte, MAX_MSG_SIZE)
		mlen, err := client.Read(received)

		if err != nil {
			//app_pl.logg.Errorf("could not read incoming message: %s\n", err.Error())
			return fmt.Errorf("could not read incoming message: %w", err)
		}

		app_pl.logg.Tracef("%s-%d : received bytes %v", app_pl.state.CurrentProcId.Owner, app_pl.state.CurrentProcId.Index, received[:mlen])

		var msg pb.Message
		err = pb.UnmarshalMsg(received[:mlen], &msg)
		if err != nil {
			//app_pl.logg.Errorf("could not unmarshal incoming message: %s\n", err.Error())
			//panic(err)
			return fmt.Errorf("could not unmarshal incoming message: %w", err)
		}

		app_pl.logg.Infof("got message %+v", &msg)

		// Upon receiving a network message, we can trigger a private link
		// deliver. Note that triggering the private link deliver acts similar
		// to an internal event -- it's not the response to some other event,
		// but internally the perfect link detected the reception of the
		// message.
		//
		// For this reason, network messages are processed separately and not
		// via the handle method.
		if msg.Type == pb.Message_NETWORK_MESSAGE {
			deliver := pb.Message{
				Type: pb.Message_PL_DELIVER,

				// Can be either app.pl or app.beb.pl, so take it from the
				// message.
				FromAbstractionId: msg.ToAbstractionId,
				ToAbstractionId:   msg.NetworkMessage.Message.ToAbstractionId,
				PlDeliver: &pb.PlDeliver{
					Message: msg.NetworkMessage.Message,
				},
			}
			// Trigger a perfect link deliver.
			trigger(app_pl.state, app_pl.queue, &deliver)
		} else {
			app_pl.logg.Errorf("received message of type %v instead of %v", msg.Type, pb.Message_NETWORK_MESSAGE)
			app_pl.state.Quit <- struct{}{}
		}

		client.Close()
	}
}
