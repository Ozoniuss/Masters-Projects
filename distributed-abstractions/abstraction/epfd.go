package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"time"
)

func intersect(s1, s2 map[*pb.ProcessId]struct{}) bool {
	for k := range s1 {
		// key is also in s2, meaning they intersect
		if _, ok := s2[k]; ok {
			return true
		}
	}
	return false
}

type Epfd struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	alive         map[*pb.ProcessId]struct{}
	suspected     map[*pb.ProcessId]struct{}
	delay         time.Duration
	delta         time.Duration

	timeoutDone chan struct{}
}

func NewEpfd(state *procstate.ProcState, queue *queue.Queue, abstractionId string) *Epfd {

	epfd := &Epfd{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		alive:         state.GetProcessesAsMap(),
		delta:         1 * time.Second,
		delay:         1 * time.Second,

		timeoutDone: make(chan struct{}, 1),
	}

	// Start the timer right away.
	epfd.timeoutDone <- struct{}{}

	go func() {
		for {
			// Do not start the timer again until the timeout event had been
			// handled.
			<-epfd.timeoutDone
			time.Sleep(epfd.delay)
			trigger(state, queue, &pb.Message{
				Type:              pb.Message_EPFD_TIMEOUT,
				FromAbstractionId: abstractionId,
				ToAbstractionId:   abstractionId,
				EpfdTimeout:       &pb.EpfdTimeout{},
			})
		}
	}()

	return epfd
}

func (epfd *Epfd) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", epfd.abstractionId)
	}

	log.Printf("%s got message: %+v\n\n", epfd.abstractionId, msg)

	switch msg.GetType() {

	case pb.Message_EPFD_TIMEOUT:
		// If alive and suspected intersect, it means we need to increase
		// timeout.
		if intersect(epfd.alive, epfd.suspected) {
			epfd.delay += epfd.delta
		}
		for _, proc := range epfd.state.Processes {
			// process is neither alive nor suspected
			_, ok1 := epfd.alive[proc]
			_, ok2 := epfd.alive[proc]
			if !ok1 && !ok2 {
				epfd.suspected[proc] = struct{}{}
				suspect := pb.Message{
					Type:              pb.Message_EPFD_SUSPECT,
					FromAbstractionId: epfd.abstractionId,
					ToAbstractionId:   epfd.abstractionId,
					EpfdSuspect: &pb.EpfdSuspect{
						Process: proc,
					},
				}
				trigger(epfd.state, epfd.queue, &suspect)
			} else if ok1 && ok2 {
				delete(epfd.suspected, proc)
				restore := pb.Message{
					Type:              pb.Message_EPFD_RESTORE,
					FromAbstractionId: epfd.abstractionId,
					ToAbstractionId:   epfd.abstractionId,
					EpfdRestore: &pb.EpfdRestore{
						Process: proc,
					},
				}
				trigger(epfd.state, epfd.queue, &restore)
			}
			plSend := pb.Message{
				FromAbstractionId: epfd.abstractionId,
				ToAbstractionId:   Next(epfd.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: proc,
					Message: &pb.Message{
						Type:                         pb.Message_EPFD_INTERNAL_HEARTBEAT_REQUEST,
						FromAbstractionId:            epfd.abstractionId,
						ToAbstractionId:              epfd.abstractionId,
						EpfdInternalHeartbeatRequest: &pb.EpfdInternalHeartbeatRequest{},
					},
				},
			}
			trigger(epfd.state, epfd.queue, &plSend)
		}
		epfd.alive = make(map[*pb.ProcessId]struct{}, len(epfd.state.Processes))

		// Can start the timer in the background now.
		epfd.timeoutDone <- struct{}{}

	case pb.Message_PL_DELIVER:
		switch msg.GetPlDeliver().GetMessage().GetType() {

		case pb.Message_EPFD_INTERNAL_HEARTBEAT_REQUEST:
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: epfd.abstractionId,
				ToAbstractionId:   Next(epfd.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: msg.GetPlDeliver().GetSender(),
					Message: &pb.Message{
						Type:                       pb.Message_EPFD_INTERNAL_HEARTBEAT_REPLY,
						SystemId:                   epfd.state.SystemId,
						FromAbstractionId:          epfd.abstractionId,
						ToAbstractionId:            epfd.abstractionId,
						EpfdInternalHeartbeatReply: &pb.EpfdInternalHeartbeatReply{},
					},
				},
			}
			trigger(epfd.state, epfd.queue, &plSend)

		case pb.Message_EPFD_INTERNAL_HEARTBEAT_REPLY:
			epfd.alive[msg.GetPlDeliver().GetSender()] = struct{}{}
		}

	}
	return nil
}
