package freekassa

import (
	"crypto/md5"
	"encoding/hex"
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"

	"github.com/gin-gonic/gin"
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

func (h *handler) Register(router *gin.Engine) {
	router.POST("/freekassa", h.freekassa)
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
	h.log.Info("order: ", order)
	h.log.Info("method: ", method)
	h.log.Info("dto: ", dto)
	hash := GetMD5Hash(method.Method.PUBLIC_KEY + ":" + dto.Amount + ":" + method.Method.SECRET_KEY + ":" + dto.Merchant_order_id)
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

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
