package pgdb

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Order struct {
	ID        int
	Content   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Connect() (*gorm.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	conn, err := gorm.Open(postgres.Open(
		fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, name),
	))
	return conn, err
}
