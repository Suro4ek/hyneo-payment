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

func (r *rcon) Give(order model.Order) error {
	r.log.Info("give ", "order ", order)
	var item model.Item
	err := r.Client.DB.Model(&model.Item{}).Preload("Server").Where("id = ?", order.Item.ID).Find(&item).Error
	if err != nil {
		return err
	}
	con, err := rcon2.Dial(item.Server.Ip+":"+item.Server.Port, item.Server.Password)
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
