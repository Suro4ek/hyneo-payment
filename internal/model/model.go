package model

import "time"

type Method struct {
	ID       uint `gorm:"primarykey"`
	Name     string
	IsActive bool
	Method   MethodKey `gorm:"foreignKey:methodId"`
}

type MethodKey struct {
	ID        uint `gorm:"primarykey"`
	SecretKey string
	PublicKey string
	MethodId  int `gorm:"methodId"`
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

type Server struct {
	ID       uint `gorm:"primarykey"`
	Ip       string
	Port     string
	Password string
	Items    []Item `gorm:"foreignKey:serverId"`
}

type Promo struct {
	ID       uint `gorm:"primarykey"`
	Name     string
	Active   bool
	Count    int
	Discount int
}
