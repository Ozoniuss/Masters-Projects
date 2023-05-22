package orders

import (
	"encoding/json"
	"io"
	"os"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// OrderDB represents the orders database. Fields are exported to allow gob
// encoding without manually implementing GobEncoder and GobDecored. DO NOT
// modify the fields directly!
type OrderDB struct {
	// id is the key, which is redundant but used for fast reads. I know I
	// could've dropped the id from order but who the fuck cares
	file    string
	Updates chan Orders
}

// copyOrders makes a copy of all orders
func copyOrders(orders Orders) Orders {
	new := make(Orders, len(orders))
	for k, v := range orders {
		new[k] = v
	}
	return new
}

/*
	Note that panics below are used for debugging purposes.
*/

// open opens the db for reading.
func (o *OrderDB) open() (*lockedfile.File, Orders) {
	f, err := lockedfile.Open(o.file)
	if err != nil {
		panic(err)
	}
	var orders = make(Orders)
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &orders)
	return f, orders
}

// openfile opens the db for writing.
func (o *OrderDB) openfile() (*lockedfile.File, Orders) {
	f, err := lockedfile.OpenFile(o.file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	var orders = make(Orders)
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &orders)
	return f, orders
}

// flush writes the database to disk.
func (o *OrderDB) flush(w *lockedfile.File, orders Orders) {
	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		panic(err)
	}

	// Ensures that the contents of the file are erased, if opened in RDWR mode.
	w.Truncate(0)
	_, err = w.WriteAt(data, 0)
	if err != nil {
		panic(err)
	}
	if o.Updates != nil {
		o.Updates <- copyOrders(orders)
	}
}

func (o *OrderDB) List() Orders {
	f, orders := o.open()
	defer f.Close()
	return orders
}

func (o *OrderDB) AddOrder(order Order) {
	f, orders := o.openfile()
	defer f.Close()
	orders[order.Id] = order
	o.flush(f, orders)
}

// RemoveOrder removes the order if it exists, or does nothing if it doesn't
// exist.
func (o *OrderDB) RemoveOrder(oid uint32) {
	f, orders := o.openfile()
	defer f.Close()
	delete(orders, oid)
	o.flush(f, orders)
}

func (o *OrderDB) TakeOrder(oid uint32) (Order, error) {
	f, orders := o.openfile()
	defer f.Close()
	retrieved, ok := orders[oid]
	if !ok {
		return Order{}, ErrOrderNotExists
	}
	if retrieved.Status != ORDER_READY {
		return Order{}, ErrOrderNotReady
	}
	delete(orders, oid)
	o.flush(f, orders)
	return retrieved, nil
}

func (o *OrderDB) ChangeOrderStatus(oid uint32, status string) {
	// Technically, this doesn't require a lock, but remains to be seen.
	f, orders := o.openfile()
	defer f.Close()
	order := orders[oid]
	order.Status = status
	orders[oid] = order
	o.flush(f, orders)
}

func NewOrdersDB(file string, c chan Orders) OrderDB {
	return OrderDB{
		file:    file,
		Updates: c,
	}
}
