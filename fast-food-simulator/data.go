package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

const (
	ORDER_QUEUE = "orders"
	EXCHANGE    = ""
	READY_QUEUE = "ready"
)

const (
	ORDER_READY     = "ready"
	ORDER_TAKEN     = "taken"
	ORDER_PREPARING = "preparing"
)

type order struct {
	Id      uint32 `json:"id"`
	Status  string `json:"status"`
	Content string `json:"content"`
}

type orders struct {
	// Thread-safe orders list
	mutex *sync.RWMutex

	// id is the key, which is redundant but used for fast reads. I know I
	// could've dropped the id from order but who the fuck cares
	OrderList map[uint32]order `json:"orders"`

	counter uint32
}

/*
	Note that panics below are used for debugging purposes.
*/

func (all *orders) String() string {
	all.mutex.RLock()
	defer all.mutex.RUnlock()
	data, err := json.MarshalIndent(all.OrderList, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

// string is used internally within mutexes to avoid deadlocks
func (all *orders) string() string {
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (all *orders) addOrder(content string) uint32 {
	all.mutex.Lock()
	defer all.mutex.Unlock()
	all.counter++
	all.OrderList[all.counter] = order{
		Id:      all.counter,
		Status:  ORDER_TAKEN,
		Content: content,
	}
	Updates <- all.string()
	return all.counter
}

func (all *orders) RemoveOrder(o order) {
	all.mutex.Lock()
	defer all.mutex.Unlock()
	if _, ok := all.OrderList[o.Id]; ok {
		panic(fmt.Sprintf("no such order: %d", o.Id))
	}
	delete(all.OrderList, o.Id)
	Updates <- all.string()
}

func (all *orders) ListOrders() []order {
	all.mutex.RLock()
	defer all.mutex.RUnlock()
	var orders = make([]order, 0, len(all.OrderList))
	for _, o := range all.OrderList {
		orders = append(orders, o)
	}
	return orders
}

func (all *orders) ChangeOrderStatus(oid uint32, status string) {
	// Technically, this doesn't require a lock, but remains to be seen.
	all.mutex.Lock()
	defer all.mutex.Unlock()
	order := all.OrderList[oid]
	order.Status = status
	all.OrderList[oid] = order
	Updates <- all.string()
}

func NewOrders() orders {
	return orders{
		mutex:     &sync.RWMutex{},
		OrderList: make(map[uint32]order, 1000),
	}
}
