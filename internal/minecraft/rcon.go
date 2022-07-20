package minecraft

import (
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/mysql"
	"strings"

	rcon2 "github.com/gorcon/rcon"
)

type rcon struct {
	Client *mysql.Client
}

func NewGive(client *mysql.Client) give.Give {
	return &rcon{
		Client: client,
	}
}

func (r *rcon) Give(orderId int) error {
	var order model.Order
	err := r.Client.DB.Model(&model.Order{}).Where("id = ?", orderId).First(&order).Error
	if err != nil {
		return err
	}
	var item model.Item
	err = r.Client.DB.Model(&model.Item{}).Where("id = ?", order.ItemId).First(&item).Error
	if err != nil {
		return err
	}
	var server model.Server
	err = r.Client.DB.Model(&model.Server{}).Where("id = ?", item.ServerId).First(&server).Error
	if err != nil {
		return err
	}
	con, err := rcon2.Dial(server.Ip+":"+server.Port, server.Password)
	if err != nil {
		return err
	}
	defer con.Close()
	_, err = con.Execute(strings.Replace(item.Command, "{user}", order.Username, -1))
	if err != nil {
		return err
	}
	order.Status = "success"
	err = r.Client.DB.Save(&order).Error
	if err != nil {
		return err
	}
	return nil
}
