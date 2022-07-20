package model

import "time"

type Method struct {
	ID       uint `gorm:"primarykey"`
	Name     string
	IsActive bool
	Method   MethodKey `gorm:"foreignKey:methodId"`
}

type MethodKey struct {
	ID         uint `gorm:"primarykey"`
	SECRET_KEY string
	PUBLIC_KEY string
	MethodId   int `gorm:"methodId"`
}

type Order struct {
	ID        uint `gorm:"primarykey"`
	Username  string
	Status    string
	Summa     int
	DateIssue time.Time
	ItemId    int `gorm:"itemId"`
	//items ????
}

type Item struct {
	ID       uint `gorm:"primarykey"`
	Command  string
	ServerId int   `gorm:"serverId"`
	Order    Order `gorm:"foreignKey:itemId"`
}

type Server struct {
	ID       uint `gorm:"primarykey"`
	Ip       string
	Port     string
	Password string
	Items    []Item `gorm:"foreignKey:serverId"`
}
