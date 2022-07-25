package qiwi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hyneo-payment/internal/give"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/logging"
	"hyneo-payment/pkg/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type handler struct {
	client *mysql.Client
	log    *logging.Logger
	Give   give.Give
}

const (
	urlBill = "https://api.qiwi.com/partner/bill/v1/bills/"
)

func NewQiwiHandler(client *mysql.Client, log *logging.Logger, give give.Give) handlers.Handler {
	return &handler{
		client: client,
		log:    log,
		Give:   give,
	}
}

func (h *handler) Register(router *gin.Engine, auth *gin.RouterGroup) {
	router.POST("/qiwi", h.qiwi)
	auth.POST("/qiwi", h.bill)
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
	invoice_parameters := fmt.Sprintf("%s|%s|%s|%s|%s", dto.Amount_currency, dto.Amount_value, dto.BillId, dto.SiteID, dto.Status_value)
	hash_request := hmac.New(sha256.New, []byte(method.Method.SecretKey))
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

func (h *handler) bill(ctx *gin.Context) {
	var dto QIWIPay
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "error form data",
		})
		return
	}
	var method model.Method
	err := h.client.DB.Model(&model.Method{}).Preload("MethodKey").Where("name = ?", "Qiwi").Find(&method).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "error get method",
		})
		return
	}
	// var methodKey model.MethodKey
	// err = h.client.DB.Model(&model.MethodKey{}).Where("methodId = ?", method.ID).First(&methodKey).Error
	// if err != nil {
	// 	ctx.AbortWithStatusJSON(400, gin.H{
	// 		"error": "error get method key",
	// 	})
	// 	return
	// }
	var item model.Item
	err = h.client.DB.Model(&model.Item{}).Where("id = ?", dto.Item_id).First(&item).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "error get item",
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
			"error": "error create order",
		})
		return
	}
	var bearer = "Bearer " + method.Method.SecretKey
	var jsonData = []byte(`{
		"amount":{
			"value": "` + fmt.Sprintf("%d", price) + `",
			"value": "RUB",
		},
		"expirationDateTime": "` + time.Now().Add(time.Hour*72).Format("2025-12-10T09:02:00+03:00") + `",}`)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s%d", urlBill, item.ID), bytes.NewBuffer(jsonData))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "error create request",
		})
		return
	}
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
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
			"error": "error unmarshal",
		})
		return
	}
	h.log.Info("raw: ", raw)
	ctx.Redirect(302, raw["payUrl"].(string))
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
