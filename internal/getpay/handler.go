package getpay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/internal/order"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	client  *mysql.Client
	log     *logging.Logger
	service order.Service
}

const (
	urlBill = "https://getpay.io/api/pay"
)

func NewGetPayHandler(client *mysql.Client, log *logging.Logger, service order.Service) handlers.Handler {
	return &handler{
		client:  client,
		log:     log,
		service: service,
	}
}

func (h *handler) Register(r *gin.Engine, auth *gin.RouterGroup) {
	r.POST("/getpay", h.getpay)
	auth.POST("/getpay", h.bill)
}

func (h *handler) getpay(ctx *gin.Context) {
	var dto GetPay
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "GetPay").Find(&method).Error
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
	hash := GetMD5Hash(dto.MerchantID + ":" + dto.Amount + ":" + dto.Merchant_order_id + ":" + method.SecretKey)
	if hash != dto.SIGN {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	go func() {
		h.service.Give(int(order.ID))
	}()
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *handler) bill(ctx *gin.Context) {
	var dto GetPayBill
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "GetPay").Find(&method).Error
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?secret=%s&wallet=%s&order=%d&resultUrl=%s&backUrl=%s&comment=%s&sum=%d",
		urlBill,
		method.SecretKey,
		method.PublicKey,
		ord.ID,
		"https://api.hyneo.ru/getpay/",
		"https://hyneo.ru/",
		"Оплата заказа "+item.Name,
		ord.Summa,
	), nil)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "not generate request",
		})
		return
	}

	// add authorization header to the req
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		log.Println("Error while reading the response bytes:", err)
		return
	}
	h.log.Info("body: ", string(body))
	var raw map[string]interface{}
	if er := json.Unmarshal(body, &raw); er != nil {
		log.Println("Error while unmarshaling the response:", er)
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	h.log.Info("raw: ", raw)
	ctx.JSON(200, gin.H{
		"status": "ok",
		"payUrl": raw["redirectUrl"].(string),
	})
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
