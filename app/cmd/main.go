package main

import (
	"avito_shop/pkg/api"
	"avito_shop/pkg/db/pgsql"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	api *api.API
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	err = PushSchema(pool)
	if err != nil {
		fmt.Println("create error", err)
		return
	}

	err = InitializeData(pool)
	if err != nil {
		fmt.Println("init error", err)
		return
	}

	srv := new(server)

	pgdb := pgsql.New(pool)
	srv.api = api.New(&pgdb, []byte("my-secret-key"))

	http.ListenAndServe(":8080", srv.api.Router())

}

func PushSchema(db *pgxpool.Pool) error {
	ctx := context.Background()

	_, err := db.Exec(ctx, string(createQuery))
	if err != nil {
		return err
	}
	return nil
}

func InitializeData(db *pgxpool.Pool) error {
	ctx := context.Background()
	var err error

	_, err = db.Exec(ctx, string(seedQuery))
	if err != nil {
		return err
	}
	return nil
}

var seedQuery = `INSERT INTO items(name, price)
VALUES  ('t-shirt', 80),
        ('cup', 20),
        ('book', 50),
        ('pen', 10),
        ('powerbank', 200),
        ('hoody', 300),
        ('umbrella', 200),
        ('socks', 10),
        ('wallet', 50),
        ('pink-hoody', 500);
`

var createQuery = `
CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name VARCHAR(32),
    password VARCHAR(32),
    balance DECIMAL(6, 0),
    items JSONB,
    CONSTRAINT idx_users_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS items(
    id SERIAL PRIMARY KEY,
    name VARCHAR(32),
    price DECIMAL(5, 0),
    CONSTRAINT idx_items_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS operations(
    id SERIAL PRIMARY KEY,
    sender VARCHAR(32),
    reciver VARCHAR(32),
    amount DECIMAL(5, 0)
);

CREATE INDEX IF NOT EXISTS idx_sender ON operations USING HASH(sender);
CREATE INDEX IF NOT EXISTS idx_reciver ON operations USING HASH(reciver);
`
