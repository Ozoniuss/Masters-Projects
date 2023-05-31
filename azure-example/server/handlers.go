package main

import (
	"azdemo/pgdb"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func handleTakeOrder(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	oidStr := r.URL.Query().Get("order")
	oid, err := strconv.Atoi(oidStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid order number: %s", oidStr)
		return
	}

	var o pgdb.Order

	// No need for transaction because status is only updated once.

	err = db.Where("id = ?", oid).First(&o).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("order %d does not exist", oid), http.StatusNotFound)
		return
	}

	if o.Status == pgdb.PREPARING {
		http.Error(w, fmt.Sprintf("order %d is still preparing", oid), http.StatusForbidden)
		return
	}

	err = db.Where("id = ?", oid).Delete(&o).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("could not take order: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "here is your order: %s", o.Content)
}
func handlePlaceOrder(w http.ResponseWriter, r *http.Request, ctx context.Context, db *gorm.DB, sender *azservicebus.Sender) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	o := pgdb.Order{
		Status:  pgdb.PREPARING,
		Content: string(body),
	}

	err = db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&o).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("could not create order: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	var oid []byte
	sbMessage := &azservicebus.Message{
		Body: binary.BigEndian.AppendUint32(oid, uint32(o.ID)),
	}
	err = sender.SendMessage(ctx, sbMessage, nil)

	// TODO: cleanup db
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Order not placed."))
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "order number: %d", o.ID)
}

func handleListOrders(w http.ResponseWriter, r *http.Request, ctx context.Context, db *gorm.DB) {
	o := make([]pgdb.Order, 10)
	err := db.Find(&o).Error
	if err != nil {
		http.Error(w, "could not list orders", http.StatusInternalServerError)
		return
	}

	allOrders, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		fmt.Printf("error marshalling orders: %s\n", err.Error())
		http.Error(w, "an error occured", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(allOrders)
}
