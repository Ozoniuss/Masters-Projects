package queue

import (
	"hw/log"
	pb "hw/protobuf"
	"sync"
)

// Queue is a thread-safe queue.
type Queue struct {
	mtx      *sync.RWMutex
	messages []*pb.Message
}

// NewQueue initializes a new queue of messages with given capacity.
func NewQueue(initcap int) *Queue {
	return &Queue{
		mtx:      &sync.RWMutex{},
		messages: make([]*pb.Message, 0, initcap),
	}
}

// Get returns the front element of the queue.
func (q *Queue) Get() *pb.Message {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg
}

// Add adds an element at the bottom of the queue.
func (q *Queue) Add(msg *pb.Message) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.messages = append(q.messages, msg)
}

// AddWithMsg is similar to add, expect it prints a message to the screen.
func (q *Queue) AddWithMsg(msg *pb.Message, show string) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	log.Println(show)
	q.messages = append(q.messages, msg)
}

// Len returns the length of the queue.
func (q *Queue) Len() int {
	q.mtx.RLock()
	defer q.mtx.RUnlock()
	return len(q.messages)
}

// Clear empties the queue.
func (q *Queue) Clear() {
	q.mtx.RLock()
	defer q.mtx.RUnlock()
	q.messages = make([]*pb.Message, 0, 1000)
}
