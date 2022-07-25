package model

import "time"

type Method struct {
	ID       int       `gorm:"primarykey;column:id"`
	Name     string    `gorm:"column:name"`
	Title    string    `gorm:"column:title"`
	IsActive bool      `gorm:"column:isActive"`
	Method   MethodKey `gorm:"foreignKey:Methodid"`
}

func (Method) TableName() string {
	return "Method"
}

type MethodKey struct {
	ID        int    `gorm:"primarykey;column:id"`
	SecretKey string `gorm:"column:SECRET_KEY"`
	PublicKey string `gorm:"column:PUBLIC_KEY"`
	Methodid  int    `gorm:"column:methodId"`
}

func (MethodKey) TableName() string {
	return "MethodKey"
}

type Order struct {
	ID        uint `gorm:"primarykey"`
	Username  string
	Status    string
	Summa     int
	DateIssue time.Time
	ItemId    int `gorm:"itemId"`
	Method    string
}

func (Order) TableName() string {
	return "Order"
}

type Item struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Command   string
	ServerId  int   `gorm:"serverId"`
	Order     Order `gorm:"foreignKey:itemId"`
	Doplata   bool
	FakePrice int
	Price     int
	Active    bool
}

func (Item) TableName() string {
	return "Item"
}

type Server struct {
	ID       uint `gorm:"primarykey"`
	Ip       string
	Port     string
	Password string
	Items    []Item `gorm:"foreignKey:serverId"`
}

func (Server) TableName() string {
	return "Server"
}

type Promo struct {
	ID       uint `gorm:"primarykey"`
	Name     string
	Active   bool
	Count    int
	Discount int
}

func (Promo) TableName() string {
	return "Promo"
}
