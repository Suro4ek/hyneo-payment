package order

import (
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/mysql"
	"time"
)

type Service struct {
	client *mysql.Client
	rcon   give.Give
}

func (s *Service) createOrder(username string, item model.Item, methodname string, promoString string) (model.Order, error) {
	var promo model.Promo
	err := s.client.DB.Model(&model.Promo{}).Where("name = ?", promoString).First(&promo).Error
	if err != nil {
		promo.Discount = 0
	}
	price := s.getPrice(username, item, promo.Discount)
	order := model.Order{
		Username:  username,
		ItemId:    int(item.ID),
		Method:    methodname,
		Summa:     price,
		Status:    "Ожидает оплаты",
		DateIssue: time.Now(),
	}
	err = s.client.DB.Create(&order).Error
	if err != nil {
		return model.Order{}, err
	}
	return order, nil
}

func (s *Service) Give(orderid int) {
	var order model.Order
	err := s.client.DB.Model(&model.Order{}).Where("id = ?", orderid).First(&order).Error
	if err != nil {
		return
	}
	s.rcon.Give(order)
}

func (s *Service) getPrice(username string, item model.Item, discount int) int {
	if item.Doplata {
		var order model.Order
		err := s.client.DB.Model(&model.Order{}).Preload("Item").Where("username = ? and doplata = true", username).Last(&order).Error
		if err != nil {
			return item.Price
		}
		if order.Item.Price > item.Price {
			return 0
		}
		if order.Item.CategoryId != item.CategoryId {
			return 0
		}
		price := item.Price - order.Summa
		if price < 0 {
			return 0
		}
		return item.Price - order.Summa
	} else if discount != 0 {
		return item.Price - (item.Price*discount)/100
	}
	return item.Price
}
