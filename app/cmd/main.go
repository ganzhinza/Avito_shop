package main

import (
	"avito_shop/pkg/api"
	"avito_shop/pkg/db/pgsql"
	"context"
	"fmt"
	"io"
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

	fd, err := os.Open("../pkg/db/dbdata/schema.sql")
	if err != nil {
		return err
	}
	defer fd.Close()

	createQuery, err := io.ReadAll(fd)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, string(createQuery))
	if err != nil {
		return err
	}
	return nil
}

func InitializeData(db *pgxpool.Pool) error {
	ctx := context.Background()

	fd, err := os.Open("../pkg/db/dbdata/data.sql")
	if err != nil {
		return err
	}
	defer fd.Close()

	seedQuery, err := io.ReadAll(fd)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, string(seedQuery))
	if err != nil {
		return err
	}
	return nil
}
