package freekassa

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	urlBill = "https://www.free-kassa.ru/merchant/cash.php"
)

type handler struct {
	client *mysql.Client
	log    *logging.Logger
	Give   give.Give
}

func NewFreeKassaHandler(client *mysql.Client, log *logging.Logger, give give.Give) handlers.Handler {
	return &handler{
		client: client,
		log:    log,
		Give:   give,
	}
}

func (h *handler) Register(router *gin.Engine, auth *gin.RouterGroup) {
	router.POST("/free_kassa", h.freekassa)
	auth.POST("/free_kassa", h.bill)
}

func (h *handler) freekassa(ctx *gin.Context) {
	var dto FreeKassa
	if err := ctx.ShouldBindQuery(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "FreeKassa").First(&method).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var order model.Order
	err = h.client.DB.Model(&model.Order{}).Where("id = ?", dto.Merchant_order_id).First(&order).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	order.Status = "Оплачен"
	err = h.client.DB.Save(&order).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	h.log.Info("order: ", order)
	h.log.Info("method: ", method)
	h.log.Info("dto: ", dto)
	hash := GetMD5Hash(method.Method.PublicKey + ":" + dto.Amount + ":" + method.Method.SecretKey + ":" + dto.Merchant_order_id)
	if hash != dto.SIGN {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	go func() {
		h.Give.Give(int(order.ID))
	}()
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *handler) bill(ctx *gin.Context) {
	var dto FreeKassaBill
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "FreeKassa").First(&method).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var item model.Item
	err = h.client.DB.Model(&model.Item{}).Where("id = ?", dto.Item_id).First(&item).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	h.log.Info("item: ", item)
	h.log.Info("method: ", method)
	h.log.Info("dto: ", dto)
	var promo model.Promo
	if dto.Promo != nil {
		_ = h.client.DB.Model(&model.Promo{}).Where("name = ?", dto.Promo).First(&promo).Error
	} else {
		promo.Discount = 0
	}
	price := h.getPrice(dto.Name, item, promo.Discount)
	order := model.Order{
		Username:  dto.Name,
		ItemId:    int(item.ID),
		Method:    method.Name,
		Summa:     price,
		Status:    "Ожидает оплаты",
		DateIssue: time.Now(),
	}
	err = h.client.DB.Create(&order).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	hash := GetMD5Hash(method.Method.PublicKey + ":" + fmt.Sprintf("%d", price) + ":" + method.Method.SecretKey + ":" + fmt.Sprintf("%d", order.ID))
	h.log.Info("hash: ", hash)
	ctx.Redirect(302, fmt.Sprintf("%s?m=%s&oa=%d&o=%d&s=%s", urlBill, method.Method.PublicKey, price, order.ID, hash))
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (h *handler) getPrice(username string, item model.Item, discount int) int {
	if item.Doplata {
		var order model.Order
		err := h.client.DB.Model(&model.Order{}).Preload("Item").Where("username = ? and doplata = true", username).Last(&order).Error
		if err != nil {
			return item.Price
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
