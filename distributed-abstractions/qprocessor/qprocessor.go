package qprocessor

import (
	"hw/abstraction"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
)

// QueueProcessor processes messages using the provided abstractions.
type QueueProcessor struct {
	abstractions map[string]abstraction.Abstraction
	stopc        chan struct{}
	logg         log.Logger
}

func NewQueueProcessor(abstractions map[string]abstraction.Abstraction, logg log.Logger) *QueueProcessor {
	return &QueueProcessor{
		abstractions: abstractions,
		stopc:        make(chan struct{}, 1),
		logg:         logg,
	}
}

func (p *QueueProcessor) processMessage(msg *pb.Message) {

	// The abstraction id that must process the message
	aid := msg.ToAbstractionId

	if aid == "" {
		p.logg.Errorf("message has no destination abstraction id")
		return
	}

	abs, ok := p.abstractions[aid]
	if !ok {
		p.logg.Errorf("no handler registered for abstraction id: %s", aid)
		return
	}

	abs.Handle(msg)
}

// Start starts the worker thread processing the messages. Can be stopped and
// restarted to process from a different queue.
func (p *QueueProcessor) Start(queue *queue.Queue) {
	go func() {
		for {
			select {
			case <-p.stopc:
				p.logg.Info("queue stopped")
				return
			default:
				if queue.Len() == 0 {
					// fmt.Println("reading")
					// time.Sleep(1 * time.Second)
					// //sugar.Infoln("Nothing is happening...")
					// //time.Sleep(1 * time.Second)
				} else {
					p.logg.Trace("processing...")
					// Get the message from top of queue.
					msg := queue.Get()
					// Process after reading from the queue since processing may trigger
					// events which are later added to the queue.
					//
					// Note that only one event can be processed at a time.
					p.processMessage(msg)
					p.logg.Trace("message processed")
				}
			}
		}
	}()
}

// Stop stops the processing queue.
func (p *QueueProcessor) Stop() {
	p.stopc <- struct{}{}
}
