package qiwi

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
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

func NewQiwiHandler(client *mysql.Client, log *logging.Logger, give give.Give) handlers.Handler {
	return &handler{
		client: client,
		log:    log,
		Give:   give,
	}
}

func (h *handler) Register(router *gin.Engine) {
	router.POST("/qiwi", h.qiwi)
}

func (h *handler) qiwi(ctx *gin.Context) {
	if err := ctx.ShouldBindHeader("X-Api-Signature-SHA256"); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	hash := ctx.GetHeader("X-Api-Signature-SHA256")
	if hash == "" {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var dto QiwiBill
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "Qiwi").First(&method).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var order model.Order
	err = h.client.DB.Model(&model.Order{}).Where("id = ?", dto.BillId).First(&order).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	h.log.Info("order: ", order)
	h.log.Info("method: ", method)
	h.log.Info("dto: ", dto)
	invoice_parameters := fmt.Sprintf("%s|%s|%s|%s|%s", dto.Amount_currency, dto.Amount_value, dto.BillId, dto.SiteID, dto.Status_value)
	hash_request := hmac.New(sha256.New, []byte(method.Method.SECRET_KEY))
	hash_request.Write([]byte(invoice_parameters))
	if hmac.Equal([]byte(hash), hash_request.Sum(nil)) {
		go func() {
			h.Give.Give(int(order.ID))
		}()
		ctx.JSON(200, gin.H{
			"status": "ok",
		})
	} else {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
	}
}
