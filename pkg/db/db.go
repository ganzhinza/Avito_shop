package db

import "avito_shop/pkg/structs"

type DB interface {
	getUser(name string) (structs.UserData, error)
	spendCoins(name string) error
	sendCoins(from, to string) error
	addUser(structs.UserData) error
}
