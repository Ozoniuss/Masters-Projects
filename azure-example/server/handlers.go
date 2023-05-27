package main

import (
	"azdemo/pgdb"
	"context"
	"encoding/binary"
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

	err = db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Where("id = ?", oid).Delete(&o).Error
	if err != nil {
		http.Error(w, fmt.Sprintf("could not take order: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
