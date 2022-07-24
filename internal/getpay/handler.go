package getpay

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

type handler struct {
	client *mysql.Client
	log    *logging.Logger
	Give   give.Give
}

const (
	urlBill = "https://getpay.io/api/pay"
)

func NewGetPayHandler(client *mysql.Client, log *logging.Logger, give give.Give) handlers.Handler {
	return &handler{
		client: client,
		log:    log,
		Give:   give,
	}
}

func (h *handler) Register(r *gin.Engine, auth *gin.RouterGroup) {
	r.POST("/getpay", h.getpay)
}

func (h *handler) getpay(ctx *gin.Context) {
	var dto GetPay
	if err := ctx.ShouldBindQuery(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "GetPay").First(&method).Error
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
	hash := GetMD5Hash(method.Method.PublicKey + ":" + dto.Amount + ":" + dto.Merchant_order_id + ":" + method.Method.SecretKey)
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
	var dto GetPayBill
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "GetPay").First(&method).Error
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
	order := model.Order{
		Username:  dto.Name,
		ItemId:    int(item.ID),
		Method:    method.Name,
		Summa:     item.Price,
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
	hash := GetMD5Hash(method.Method.PublicKey + ":" + fmt.Sprintf("%d", item.Price) + ":" + fmt.Sprintf("%d", order.ID) + ":" + method.Method.SecretKey)
	h.log.Info("hash: ", hash)
	ctx.Redirect(302, fmt.Sprintf("%s?m=%s&oa=%d&o=%d&s=%s", urlBill, method.Method.PublicKey, order.Summa, order.ID, hash))
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
