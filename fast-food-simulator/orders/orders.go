package orders

import (
	"encoding/json"
	"errors"
)

const (
	ORDER_READY     = "ready"
	ORDER_TAKEN     = "taken"
	ORDER_PREPARING = "preparing"
)

var (
	ErrOrderNotExists = errors.New("order does not exist")
	ErrOrderNotReady  = errors.New("order is not ready")
)

type Order struct {
	Id      uint32 `json:"id"`
	Status  string `json:"status"`
	Content string `json:"content"`
}

type Orders map[uint32]Order

// String returns a one-line representation of the orders in json format.
func (o Orders) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// StringFormat returns a more human-readable representation of the orders in
// json format.
func (o Orders) StringFormat() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}
