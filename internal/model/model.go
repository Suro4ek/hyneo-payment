package model

import "time"

type Method struct {
	ID       uint   `gorm:"primarykey;column:id"`
	Name     string `gorm:"column:name"`
	Title    string `gorm:"column:title"`
	IsActive bool   `gorm:"column:isActive"`
}

func (Method) TableName() string {
	return "Method"
}

type MethodKey struct {
	ID         uint   `gorm:"primarykey;column:id"`
	SecretKey  string `gorm:"column:SECRET_KEY"`
	SecretKey2 string `gorm:"column:SECRET_KEY2"`
	PublicKey  string `gorm:"column:PUBLIC_KEY"`
	Methodid   int    `gorm:"column:methodId"`
	Method     Method `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:Methodid"`
}

func (MethodKey) TableName() string {
	return "MethodKey"
}

type Order struct {
	ID        uint      `gorm:"primarykey;column:id"`
	Username  string    `gorm:"column:username"`
	Status    string    `gorm:"column:status"`
	Summa     int       `gorm:"column:summa"`
	DateIssue time.Time `gorm:"column:dateIssue"`
	ItemId    int       `gorm:"column:itemId"`
	Method    string    `gorm:"column:method"`
	Item      Item      `gorm:"foreignKey:ItemId"`
	PromoId   *int      `gorm:"column:promoId"`
	Promo     *Promo    `gorm:"foreignKey:PromoId"`
}

func (Order) TableName() string {
	return "Order"
}

type Item struct {
	ID          uint   `gorm:"primarykey;column:id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Price       int    `gorm:"column:price"`
	FakePrice   int    `gorm:"column:fakePrice"`
	Doplata     bool   `gorm:"column:doplata"`
	Active      bool   `gorm:"column:active"`
	ImageSRC    string `gorm:"column:imageSrc"`
	CategoryId  int    `gorm:"column:categoryId"`
	Command     string `gorm:"column:command"`
	ServerId    int    `gorm:"serverId"`
	Server      Server `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:ServerId"`
}

func (Item) TableName() string {
	return "Item"
}

type Server struct {
	ID       uint   `gorm:"primarykey;column:id"`
	Name     string `gorm:"column:name"`
	Ip       string `gorm:"column:ip"`
	Port     string `gorm:"column:port"`
	Password string `gorm:"column:password"`
	Active   bool   `gorm:"column:active"`
}

func (Server) TableName() string {
	return "Server"
}

type Promo struct {
	ID       uint   `gorm:"primarykey;column:id"`
	Name     string `gorm:"column:name"`
	Active   bool   `gorm:"column:active"`
	Count    int    `gorm:"column:count"`
	Discount int    `gorm:"column:discount"`
}

func (Promo) TableName() string {
	return "Promo"
}

type Online struct {
	ID  uint `gorm:"primarykey;column:id"`
	Max int  `gorm:"column:max"`
}
