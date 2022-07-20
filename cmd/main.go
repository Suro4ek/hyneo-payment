package main

import (
	"context"
	"hyneo-payment/internal/config"
	freekassa "hyneo-payment/internal/free_kassa"
	"hyneo-payment/internal/getpay"
	"hyneo-payment/internal/minecraft"
	"hyneo-payment/internal/qiwi"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"

	"github.com/gin-gonic/gin"
	"github.com/gorcon/rcon"
)

func main() {
	logging.Init()
	log := logging.GetLogger()
	cfg := config.GetConfig()
	client := mysql.NewClient(context.TODO(), 5, cfg.MySQL)
	rcon.Dial("127.0.0.1:16260", "password")
	RunServer(client, &log)
}

func RunServer(client *mysql.Client, log *logging.Logger) {
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

	free_kassaHandler := freekassa.NewFreeKassaHandler(client, log, give)
	free_kassaHandler.Register(r)

	getpay_handler := getpay.NewGetPayHandler(client, log, give)
	getpay_handler.Register(r)

	qiwi_handler := qiwi.NewQiwiHandler(client, log, give)
	qiwi_handler.Register(r)

	r.Run(":8080")
}
