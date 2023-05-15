package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"fastfood/counter"
	"fastfood/orders"

	amqp "github.com/rabbitmq/amqp091-go"
)

var OrdersDB = orders.NewOrdersDB("orders.json")
var Counter = counter.NewCounter("count")

func main() {

	fmt.Println("23")

	// Since the entire application runs within a single connection, it is
	// fine to share the connection. It is recommended that each thread creates
	// its own channel to communicate with the broker, even is sharing the
	// connection.
	conn, err := amqp.Dial("amqp://fast:food@localhost:5672/fastfood")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Flags for worker and waiter. Note that it is considered that the flag
	// was not specified if either is 0, for simplicity.
	//
	// Specifying no worker flag (or having it set to 0) means workers would
	// run in a different program, not a different thread.
	var workerno, waiterno uint

	flag.UintVar(&workerno, "worker", 0, "Specify number of workers")
	flag.UintVar(&waiterno, "waiter", 2, "Specify number of waiters")

	flag.Parse()

	if waiterno == 0 {
		panic("at least one waiter required")
	}

	for i := 0; i < int(workerno); i++ {
		// Run two fast food workers
		go worker(i, conn)
	}

	for i := 0; i < int(waiterno); i++ {
		// Run one waiter.
		go waiter(i, conn)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		ORDER_QUEUE, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		handleOrder(w, r, ctx, ch)
	})
	mux.HandleFunc("/ready", handleReady)
	mux.HandleFunc("/take", handleTakeOrder)
	http.ListenAndServe(":7777", mux)
}

func handleOrder(w http.ResponseWriter, r *http.Request, ctx context.Context, ch *amqp.Channel) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	oid := Counter.Inc()
	order := orders.Order{
		Id:      oid,
		Status:  orders.ORDER_TAKEN,
		Content: string(body),
	}
	OrdersDB.AddOrder(order)

	var serialized []byte
	serialized = binary.BigEndian.AppendUint32(serialized, oid)

	err = ch.PublishWithContext(ctx,
		EXCHANGE,
		ORDER_QUEUE,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         serialized,
		},
	)

	// TODO: nicer error handling
	if err != nil {
		panic("publishing failed")
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "order number: %v", oid)
}

func handleTakeOrder(w http.ResponseWriter, r *http.Request) {

	oidStr := r.URL.Query().Get("order")
	oid, err := strconv.ParseInt(oidStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid order number: %s", oidStr)
		return
	}

	order, err := OrdersDB.TakeOrder(uint32(oid))
	if err != nil {
		switch {
		case errors.Is(err, orders.OrderNotExists):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Order %d does not exist.", uint32(oid))
			return
		case errors.Is(err, orders.OrderNotReady):
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Order %d not ready yet.", uint32(oid))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte{':', '('})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "serving: %s", order.Content)
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	f, ok := w.(http.Flusher)
	if !ok {
		panic("could not convert to flusher")
	}

	io.WriteString(w, "Listening for orders...\n")
	f.Flush()

	io.WriteString(w, OrdersDB.String())
	w.Write([]byte{'\n'})
	f.Flush()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("closed")
			return
		case u := <-OrdersDB.Updates:
			io.WriteString(w, u)
			w.Write([]byte{'\n'})
			f.Flush()
		}
	}
}
