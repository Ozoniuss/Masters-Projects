package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"azdemo/pgdb"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func run() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pgdb.Connect()
	if err != nil {
		return fmt.Errorf("connecting to db: %s", err.Error())
	}

	client, err := azservicebus.NewClientFromConnectionString(os.Getenv("SB_WRITE_ENDPOINT"), nil)
	if err != nil {
		return fmt.Errorf("connecting to service bus: %s", err.Error())
	}
	defer client.Close(ctx)

	sender, err := client.NewSender("orders", nil)
	if err != nil {
		panic(err)
	}
	defer sender.Close(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/take", func(w http.ResponseWriter, r *http.Request) {
		handleTakeOrder(w, r, db)
	})
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		handlePlaceOrder(w, r, ctx, db, sender)
	})

	return http.ListenAndServe(":5566", mux)
}

func main() {

	if err := run(); err != nil {
		fmt.Printf("Error running main: %s", err.Error())
		os.Exit(1)
	}
}
