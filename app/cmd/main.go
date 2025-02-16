package main

import (
	"avito_shop/pkg/api"
	"avito_shop/pkg/db/pgsql"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:rootroot@localhost:5433/films_search")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	err = CreateDatabase(pool)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = InitializeData(pool)
	if err != nil {
		fmt.Println(err)
		return
	}

	var pgdb pgsql.DB
	pgdb = pgsql.New(pool)
	api := api.New(&pgdb)
}

func CreateDatabase(db *pgxpool.Pool) error {
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
