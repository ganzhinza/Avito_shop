package pgsql

import (
	"avito_shop/pkg/structs"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) DB {
	return DB{pool: pool}
}

func (db *DB) GetUserWithHistory(name string) (structs.UserWithHistory, error) {
	var userWithHistory structs.UserWithHistory

	user, err := db.GetUser(name)
	if err != nil {
		return structs.UserWithHistory{}, err
	}

	userWithHistory.Coins = user.Balance
	userWithHistory.Inventory = user.Inventory

	userWithHistory.CoinsHistory.Recived, err = db.getReciveHistory(name)
	if err != nil {
		return structs.UserWithHistory{}, err
	}

	userWithHistory.CoinsHistory.Sent, err = db.getSentHistory(name)
	if err != nil {
		return structs.UserWithHistory{}, err
	}

	return userWithHistory, nil
}

func (db *DB) SendCoins(sender string, transferInfo structs.CoinsSend) error {
	err := db.checkTransferPosibility(sender, transferInfo)
	if err != nil {
		return err
	}

	err = db.sendCoins(sender, transferInfo)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) BuyItem(userName, itemName string) error {
	user, item, err := db.checkPurchasePossibility(userName, itemName)
	if err != nil {
		return err
	}

	err = db.makePurchase(user, item)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetUser(name string) (bool, structs.User, error) {
	ctx := context.Background()
	row := db.pool.QueryRow(ctx, `SELECT name, password, balance, items FROM users WHERE name = $1`, name)

	var user structs.User
	err := row.Scan(&user.Name, &user.Password, &user.Balance, &user.Inventory)
	if err == pgx.ErrNoRows {
		return false, structs.User{}, nil
	}
	if err != nil {
		return false, structs.User{}, err
	}

	return true, user, nil
}

func (db *DB) AddUser(user structs.User) error {
	_, err := db.pool.Exec(context.Background(), `INSERT INTO users (name, password, balance) VALUES ($1, $2, $3)`, user.Name, user.Password, user.Balance)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetItem(name string) (structs.Item, error) {
	ctx := context.Background()
	row := db.pool.QueryRow(ctx, `SELECT name, price FROM items WHERE name = $1`, name)

	var item structs.Item
	err := row.Scan(&item.Name, &item.Price)
	if err != nil {
		return item, err
	}

	return item, nil
}

func (db *DB) getSentHistory(sender string) ([]structs.CoinsSend, error) {
	ctx := context.Background()
	coinsSend := make([]structs.CoinsSend, 0)
	rows, err := db.pool.Query(ctx, `SELECT reciver, SUM(amount) FROM operations WHERE sender = $1 GROUP BY reciver`, sender)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var op structs.CoinsSend
		err := rows.Scan(&op.ToUser, &op.Amount)
		if err != nil {
			return nil, nil
		}
		coinsSend = append(coinsSend, op)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return coinsSend, nil
}

func (db *DB) getReciveHistory(reciver string) ([]structs.CoinsRecive, error) {
	ctx := context.Background()
	coinsRecived := make([]structs.CoinsRecive, 0)
	rows, err := db.pool.Query(ctx, `SELECT sender, SUM(amount) FROM operations WHERE reciver = $1 GROUP BY sender`, reciver)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var op structs.CoinsRecive
		err := rows.Scan(&op.FromUser, &op.Amount)
		if err != nil {
			return nil, nil
		}
		coinsRecived = append(coinsRecived, op)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return coinsRecived, nil
}

func (db *DB) checkPurchasePossibility(userName, itemName string) (structs.User, structs.Item, error) {
	user, err := db.GetUser(userName)
	if err != nil {
		return structs.User{}, structs.Item{}, err
	}

	item, err := db.GetItem(itemName)
	if err != nil {
		return structs.User{}, structs.Item{}, err
	}
	if user.Balance < item.Price {
		return structs.User{}, structs.Item{}, fmt.Errorf("not enough coins")
	}
	return user, item, nil
}

func (db *DB) makePurchase(user structs.User, item structs.Item) error {
	ctx := context.Background()

	user = user.AppendItem(item.Name)
	JSONItems, err := json.Marshal(user.Inventory)
	if err != nil {
		return err
	}
	user.Balance -= item.Price

	_, err = db.pool.Exec(ctx, "UPDATE TABLE users SET balance = $1, items = $2 WHERE name = $3", user.Balance, JSONItems, user.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) checkTransferPosibility(sender string, transferInfo structs.CoinsSend) error {
	user, err := db.GetUser(sender)
	if err != nil {
		return err
	}
	_, err = db.GetUser(transferInfo.ToUser)
	if err != nil {
		return err
	}

	if user.Balance < transferInfo.Amount {
		return fmt.Errorf("not enough coins")
	}

	return nil
}

func (db *DB) sendCoins(sender string, transferInfo structs.CoinsSend) error {
	ctx := context.Background()
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "UPDATE TABLE users SET balance = balance + $1 WHERE name = $2", transferInfo.Amount, transferInfo.ToUser)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE TABLE users SET balance = balance - $1 WHERE name = $2", transferInfo.Amount, sender)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO operations (sender, reciver, amount) VALUES ($1, $2, $3)", sender, transferInfo.ToUser, transferInfo.Amount)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
