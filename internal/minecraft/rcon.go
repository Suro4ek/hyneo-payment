package minecraft

import (
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"strings"

	rcon2 "github.com/gorcon/rcon"
)

type rcon struct {
	Client *mysql.Client
	log    *logging.Logger
}

func NewGive(client *mysql.Client, log *logging.Logger) give.Give {
	return &rcon{
		Client: client,
		log:    log,
	}
}

func (r *rcon) Give(orderId int) error {
	var order model.Order
	err := r.Client.DB.Model(&model.Order{}).Preload("Item").Where("id = ?", orderId).First(&order).Error
	if err != nil {
		return err
	}
	r.log.Info("give ", "order ", order)
	var server model.Server
	err = r.Client.DB.Model(&model.Server{}).Where("id = ?", order.Item.ServerId).First(&server).Error
	if err != nil {
		return err
	}
	con, err := rcon2.Dial(server.Ip+":"+server.Port, server.Password)
	if err != nil {
		return err
	}
	defer con.Close()
	_, err = con.Execute(strings.Replace(order.Item.Command, "{user}", order.Username, -1))
	if err != nil {
		return err
	}
	order.Status = "Выдано"
	err = r.Client.DB.Save(&order).Error
	if err != nil {
		return err
	}
	return nil
}
