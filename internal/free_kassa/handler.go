package freekassa

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/internal/order"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
)

const (
	urlBill = "https://pay.freekassa.ru/"
)

type handler struct {
	client  *mysql.Client
	log     *logging.Logger
	service order.Service
}

func NewFreeKassaHandler(client *mysql.Client, log *logging.Logger, service order.Service) handlers.Handler {
	return &handler{
		client:  client,
		log:     log,
		service: service,
	}
}

func (h *handler) Register(router *gin.Engine, auth *gin.RouterGroup) {
	router.POST("/free_kassa", h.freekassa)
	auth.POST("/free_kassa", h.bill)
}

func (h *handler) freekassa(ctx *gin.Context) {
	var dto FreeKassa
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "FreeKassa").Find(&method).Error
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
	h.log.Info("order: ", order)
	h.log.Info("method: ", method)
	h.log.Info("dto: ", dto)
	beforeHash := method.PublicKey + ":" + dto.Amount + ":" + method.SecretKey2 + ":" + dto.Merchant_order_id
	hash := GetMD5Hash(beforeHash)
	if hash != dto.SIGN {
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
	go func() {
		h.service.Give(int(order.ID))
	}()
	ctx.String(200, "YES")
}

func (h *handler) bill(ctx *gin.Context) {
	var dto FreeKassaBill
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "FreeKassa").Find(&method).Error
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
	ord, err := h.service.CreateOrder(dto.Name, item, method.Method.Name, dto.Promo)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	beforeHash := method.PublicKey + ":" + fmt.Sprintf("%d", ord.Summa) + ":" + method.SecretKey + ":RUB:" + fmt.Sprintf("%d", ord.ID)
	hash := GetMD5Hash(beforeHash)
	ctx.JSON(200, gin.H{
		"status": "ok",
		"payUrl": fmt.Sprintf("%s?m=%s&oa=%d&o=%d&s=%s&currency=RUB", urlBill, method.PublicKey, ord.Summa, ord.ID, hash),
	})
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
