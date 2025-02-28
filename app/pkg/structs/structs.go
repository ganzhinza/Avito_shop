package structs

type CoinsSend struct {
	ToUser string
	Amount uint
}

type CoinsRecive struct {
	FromUser string
	Amount   uint
}

type AuthInfo struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Name      string
	Password  string
	Balance   uint
	Inventory []InventoryItem
}

type Item struct {
	Name  string
	Price uint
}

type InventoryItem struct {
	Type     string
	Quantity uint
}

type CoinsHistory struct {
	Recived []CoinsRecive
	Sent    []CoinsSend
}

type UserWithHistory struct {
	Coins        uint
	Inventory    []InventoryItem
	CoinsHistory CoinsHistory
}

func (u *User) AppendItem(item string) User {
	for i := range u.Inventory {
		if u.Inventory[i].Type == item {
			u.Inventory[i].Quantity++
			return *u
		}
	}
	u.Inventory = append(u.Inventory, InventoryItem{Type: item, Quantity: 1})
	return *u
}

type DBerror struct{}

func (db *DBerror) Error() string {
	return "db error"
}

type NotExistsErr struct{}

func (db *NotExistsErr) Error() string {
	return "not exists"
}

type JSONerror struct{}

func (db *JSONerror) Error() string {
	return "JSON error"
}

type NotEnough struct{}

func (db *NotEnough) Error() string {
	return "Not enogh coins"
}
