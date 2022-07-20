package mysql

import (
	"context"
	"fmt"
	"hyneo-payment/internal/config"
	"hyneo-payment/pkg/utils"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
	error
}

func NewClient(ctx context.Context, maxAttemps int, sc config.MySQL) (client *Client) {
	var pool *gorm.DB
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", sc.User, sc.Password, sc.Host, sc.Port, sc.DB)
	fmt.Println(dsn)
	err = utils.DoWithTries(func() error {
		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
		return nil
	}, maxAttemps, 5*time.Second)

	if err != nil {
		log.Fatal("error to connect in MySQL max attemtps")
	}
	return &Client{
		DB:    pool,
		error: nil,
	}
}
