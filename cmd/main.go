package main

import (
	"context"
	"hyneo-payment/internal/config"
	freekassa "hyneo-payment/internal/free_kassa"
	"hyneo-payment/internal/getpay"
	middleware2 "hyneo-payment/internal/middleware"
	"hyneo-payment/internal/minecraft"
	"hyneo-payment/internal/model"
	"hyneo-payment/internal/online"
	"hyneo-payment/internal/order"
	"hyneo-payment/internal/qiwi"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func main() {
	logging.Init()
	log := logging.GetLogger()
	cfg := config.GetConfig()
	client := mysql.NewClient(context.TODO(), 5, cfg.MySQL)
	token, _ := generateJwtToken(cfg)
	client.DB.AutoMigrate(&model.Online{})
	log.Info(token)
	ticker := time.NewTicker(time.Hour * 5)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				deleteOldOrders(client)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	RunServer(client, &log, cfg)
}

func deleteOldOrders(client *mysql.Client) {
	client.DB.Exec("DELETE FROM `Order` WHERE `dateIssue` <= ( CURDATE() - INTERVAL 2 DAY ) AND `status`='Ожидает оплаты';")
}

func RunServer(client *mysql.Client, log *logging.Logger, config *config.Config) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	var trusted = make([]string, 0)
	free_kassa := []string{
		"136.243.38.147",
		"136.243.38.149",
		"136.243.38.150", "136.243.38.151", "136.243.38.189", "136.243.38.108",
		"46.174.50.247",
	}
	qiwi_trust := []string{
		"79.142.16.0/20",
		"195.189.100.0/22",
		"91.232.230.0/23",
		"91.213.51.0/24",
	}
	getpay_trust := []string{
		"45.135.33.28", "109.248.166.7", "45.133.223.184", "45.94.229.98",
	}
	trusted = append(trusted, free_kassa...)
	trusted = append(trusted, qiwi_trust...)
	trusted = append(trusted, getpay_trust...)
	err := r.SetTrustedProxies(trusted)
	if err != nil {
		log.Error(err)
	}
	give := minecraft.NewGive(client, log)
	orderService := order.Service{
		Client: client,
		Rcon:   give,
	}
	middleware := middleware2.NewMiddleware(config)
	auth := r.Group("/bill", middleware.Auth())
	freeKassahandler := freekassa.NewFreeKassaHandler(client, log, orderService)
	freeKassahandler.Register(r, auth)

	getpayHandler := getpay.NewGetPayHandler(client, log, orderService)
	getpayHandler.Register(r, auth)

	qiwiHandler := qiwi.NewQiwiHandler(client, log, orderService)
	qiwiHandler.Register(r, auth)

	onlineHandler := online.NewOnlineHandler(client, config)
	onlineHandler.Register(r, auth)

	if err := os.Mkdir("images", os.ModePerm); err != nil {
		log.Error(err)
	}
	r.StaticFS("/images", http.Dir("images"))
	r.Run(":8080")
}

func generateJwtToken(cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	return token.SignedString([]byte(cfg.SECRET))
}
