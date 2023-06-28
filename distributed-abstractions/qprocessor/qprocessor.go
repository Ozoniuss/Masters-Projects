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
}

func NewQueueProcessor(abstractions map[string]abstraction.Abstraction, stopc chan struct{}) *QueueProcessor {
	if stopc == nil {
		stopc = make(chan struct{}, 1)
	}
	return &QueueProcessor{
		abstractions: abstractions,

		stopc: stopc,
	}
}

func (p *QueueProcessor) processMessage(msg *pb.Message) {

	if msg == nil {
		log.Printf("[qprocessor] received nil message while processing\n\n")
		p.Stop()
	}

	// log.Printf("processing message %+v\n\n", msg)

	// The abstraction id that must process the message
	aid := msg.GetToAbstractionId()

	if aid == "" {
		log.Printf("[qprocessor] message has no destination abstraction id\n\n")
		return
	}

	abs, ok := p.abstractions[aid]
	if !ok {
		log.Printf("[qprocessor] no handler registered for abstraction id: %s\n\n", aid)
		return
	}

	log.Printf("[to %s] {%+v}\n\n", msg.ToAbstractionId, msg)

	err := abs.Handle(msg)
	if err != nil {
		log.Printf("[qprocessor] received error while handling message: %s\n\n", err.Error())
		p.Stop()
	}
}

// Start starts the worker thread processing the messages. Can be stopped and
// restarted to process from a different queue.
func (p *QueueProcessor) Start(queue *queue.Queue) {
	go func() {
		for {
			select {
			case <-p.stopc:
				log.Printf("queue cleared\n\n")
				queue.Clear()
			default:
				if queue.Len() == 0 {
					// fmt.Println("reading")
					// time.Sleep(1 * time.Second)
					// //sugar.Infoln("Nothing is happening...")
					// //time.Sleep(1 * time.Second)
				} else {
					// log.Printf("processing...\n\n")
					// Get the message from top of queue.
					msg := queue.Get()
					// Process after reading from the queue since processing may trigger
					// events which are later added to the queue.
					//
					// Note that only one event can be processed at a time.
					p.processMessage(msg)
					// log.Printf("message processed\n\n")
				}
			}
		}
	}()
}

// Stop stops the processing queue.
func (p *QueueProcessor) Stop() {
	p.stopc <- struct{}{}
}
