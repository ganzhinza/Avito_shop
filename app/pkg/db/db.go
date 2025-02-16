package db

import "avito_shop/pkg/structs"

type Interface interface {
	GetUserWithHistory(name string) (structs.UserWithHistory, error)
	SendCoins(from string, transferInfo structs.CoinsSend) error
	BuyItem(name, item string) error
	GetUser(name string) (bool, structs.User, error)
	AddUser(structs.User) error
}
