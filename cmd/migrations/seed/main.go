package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/vesselchuckk/go-social/internal/store"
)

func main() {
	conn, err := store.NewPostgresDB()
	if err != nil {

	}
 
	storage := store.NewStorage(conn)

	storage.Seed(storage, sqlx.NewDb(conn, "postgres"))
}
