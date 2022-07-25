package main

import (
	"context"
	"hyneo-payment/internal/config"
	freekassa "hyneo-payment/internal/free_kassa"
	"hyneo-payment/internal/getpay"
	middleware2 "hyneo-payment/internal/middleware"
	"hyneo-payment/internal/minecraft"
	"hyneo-payment/internal/qiwi"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
)

func main() {
	logging.Init()
	log := logging.GetLogger()
	cfg := config.GetConfig()
	client := mysql.NewClient(context.TODO(), 5, cfg.MySQL)
	token, _ := generateJwtToken(cfg)
	log.Info(token)
	RunServer(client, &log, cfg)
}

func RunServer(client *mysql.Client, log *logging.Logger, config *config.Config) {
	r := gin.Default()
	var trusted = make([]string, 0)
	free_kassa := []string{
		"136.243.38.147",
		"136.243.38.149",
		"136.243.38.150", "136.243.38.151", "136.243.38.189", "136.243.38.108",
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
	r.SetTrustedProxies(trusted)
	give := minecraft.NewGive(client)
	middleware := middleware2.NewMiddleware(config)
	auth := r.Group("/bill", middleware.Auth())
	free_kassaHandler := freekassa.NewFreeKassaHandler(client, log, give)
	free_kassaHandler.Register(r, auth)

	getpay_handler := getpay.NewGetPayHandler(client, log, give)
	getpay_handler.Register(r, auth)

	qiwi_handler := qiwi.NewQiwiHandler(client, log, give)
	qiwi_handler.Register(r, auth)

	r.Run(":8080")
}

func generateJwtToken(cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	return token.SignedString([]byte(cfg.SECRET))
}
