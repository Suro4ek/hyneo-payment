package order

import (
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/mysql"
	"time"
)

type Service struct {
	Client *mysql.Client
	Rcon   give.Give
}

func (s *Service) CreateOrder(username string, item model.Item, methodname string, promoString *string) (model.Order, error) {
	var promo model.Promo
	if promoString != nil {
		err := s.Client.DB.Model(&model.Promo{}).Where("name = ?", promoString).First(&promo).Error
		if err != nil {
			promo.Discount = 0
		}
	} else {
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
	err := s.Client.DB.Create(&order).Error
	if err != nil {
		return model.Order{}, err
	}
	return order, nil
}

func (s *Service) Give(orderid int) {
	var order model.Order
	err := s.Client.DB.Model(&model.Order{}).Where("id = ?", orderid).First(&order).Error
	if err != nil {
		return
	}
	err = s.Client.DB.Model(&order).Update("status", "Оплачен").Error
	if err != nil {
		return
	}
	if order.Promo != nil {
		var promo *model.Promo
		err = s.Client.DB.Model(&model.Promo{}).Where("id = ?", order.Promo.ID).First(&promo).Error
		if err != nil {
			return
		}
		if promo.Count != -1 {
			promo.Count--
			err = s.Client.DB.Model(&promo).Update("count", promo.Count).Error
			if err != nil {
				return
			}
		}
	}
	s.Rcon.Give(order)
}

func (s *Service) getPrice(username string, item model.Item, discount int) int {
	if item.Doplata {
		var order model.Order
		err := s.Client.DB.Model(&model.Order{}).Preload("Item").Where(&model.Order{Username: username,
			Item: model.Item{CategoryId: item.CategoryId, Doplata: true}}).Last(&order).Error
		if err != nil {
			return item.Price
		}
		if order.Item.Price > item.Price {
			return 0
		}
		price := item.Price - order.Item.Price
		if price < 0 {
			return 0
		}
		return item.Price - order.Item.Price
	} else if discount != 0 {
		return item.Price - (item.Price*discount)/100
	}
	return item.Price
}
