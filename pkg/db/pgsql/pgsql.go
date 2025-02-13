package pgsql

import (
	"avito_shop/pkg/structs"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

// TODO
func (db *DB) getUser(name string) (structs.UserData, error) {
	return structs.UserData{}, nil
}

// TODO
func (db *DB) spendCoins(name string) error {
	return nil
}

// TODO
func (db *DB) sendCoins(from, to string) error {
	return nil
}

// TODO
func (db *DB) addUser(structs.UserData) error {
	return nil
}
