package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Orders = NewOrders()
var Updates = make(chan string, 1000)

func main() {

	fmt.Println("23")

	for i := 0; i < 2; i++ {
		// Run two fast food workers
		go worker(i)
	}

	for i := 0; i < 2; i++ {
		// Run one waiter.
		go waiter(i)
	}

	conn, err := amqp.Dial("amqp://fast:food@localhost:5672/fastfood")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

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
	http.ListenAndServe(":7777", mux)
}

func handleOrder(w http.ResponseWriter, r *http.Request, ctx context.Context, ch *amqp.Channel) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	fmt.Println(body)

	oid := Orders.addOrder(string(body))

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
	fmt.Fprintf(w, "order number: %d", oid)
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got request: %+v", *r)
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

	io.WriteString(w, Orders.String())
	w.Write([]byte{'\n'})
	f.Flush()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("closed")
			return
		case u := <-Updates:
			io.WriteString(w, u)
			w.Write([]byte{'\n'})
			f.Flush()
		}
	}
}
