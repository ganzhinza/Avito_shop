package structs

type CoinsMessage struct {
	toUser string
	amount int
}

type User struct {
	name     string
	password string
}

type UserData struct {
	name     string
	password string
	balance  uint
}
