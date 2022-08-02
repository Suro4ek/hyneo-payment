package minecraft

import (
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"strings"

	"github.com/willroberts/minecraft-client"
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
	err := r.Client.DB.Model(&model.Item{}).Preload("Server").Where("id = ?", order.ItemId).Find(&item).Error
	if err != nil {
		r.log.Error("error ", err)
		return err
	}
	r.log.Info("item ", "item ", item)
	var server model.Server
	r.log.Info("serverID", *item.ServerId)
	var serverId uint = item.Server.ID
	err = r.Client.DB.Model(&model.Server{}).Where("id = ?", serverId).Find(&server).Error
	if err != nil {
		r.log.Error("error ", err)
		return err
	}
	r.log.Info("server ", server)
	client, err := minecraft.NewClient(server.Ip + ":25567")
	if err != nil {
		r.log.Error("error ", err)
	}
	defer client.Close()

	// Send some commands.
	if err := client.Authenticate("uSIhRYaaMkHzZijrcOPSJofHJr80udKc"); err != nil {
		r.log.Error("error ", err)
	}
	_, err = client.SendCommand(strings.Replace(item.Command, "{user}", order.Username, -1))
	if err != nil {
		r.log.Error("error ", err)
	}
	err = r.Client.DB.Model(&order).Update("status", "Выдано").Error
	if err != nil {
		return err
	}
	return nil
}
