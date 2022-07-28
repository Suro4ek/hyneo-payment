package qiwi

import (
	"crypto/hmac"
	"crypto/sha256"
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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type handler struct {
	client  *mysql.Client
	log     *logging.Logger
	service order.Service
}

const (
	urlBill = "https://api.qiwi.com/partner/bill/v1/bills/"
)

func NewQiwiHandler(client *mysql.Client, log *logging.Logger, service order.Service) handlers.Handler {
	return &handler{
		client:  client,
		log:     log,
		service: service,
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
	var dto PaymentUpdate
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	fmt.Print("dto ", dto)
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "Qiwi").Find(&method).Error
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var order model.Order
	err = h.client.DB.Model(&model.Order{}).Where("id = ?", dto.Bill.BillId).First(&order).Error
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
	invoiceParameters := ""
	invoiceParameters += dto.Bill.Amount.Currency + "|"
	invoiceParameters += dto.Bill.Amount.Value + "|"
	invoiceParameters += dto.Bill.BillId + "|"
	invoiceParameters += dto.Bill.SiteId + "|"
	invoiceParameters += dto.Bill.Status.Value
	hash_request := hmac.New(sha256.New, []byte(method.SecretKey))
	hash_request.Write([]byte(invoiceParameters))
	sha := hex.EncodeToString(hash_request.Sum(nil))
	if sha == hash {
		go func() {
			h.service.Give(int(order.ID))
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
	var method model.MethodKey
	err := h.client.DB.Model(&model.MethodKey{}).Joins("Method").Where("Method.name = ?", "Qiwi").Find(&method).Error
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
	ord, err := h.service.CreateOrder(dto.Name, item, method.Method.Name, dto.Promo)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var bearer = "Bearer " + method.SecretKey
	expireTime := time.Now().UTC().Add(time.Hour * 72).Format("2006-01-02T15:04:05+00:00")
	h.log.Info("expireTime: ", expireTime)
	bill := CreateBill()
	bill.Amount.Currency = "RUB"
	bill.Amount.Value = fmt.Sprintf("%d", item.Price)
	bill.Comment = "Оплата заказа " + item.Name
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s%d", urlBill, ord.ID), strings.NewReader(bill.toJSON()))
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
	h.log.Info("resp code: ", resp.StatusCode)
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
	ctx.JSON(200, gin.H{
		"status": "ok",
		"payUrl": raw["payUrl"].(string),
	})
}

type Amount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

// CreateBill
// Method creates Bill with default parameters
func CreateBill() *Bill {
	return &Bill{
		Amount:             Amount{},
		ExpirationDateTime: time.Now().UTC().Add(time.Hour * 72).Format("2006-01-02T15:04:05+00:00"),
	}
}

type PaymentUpdate struct {
	Bill    QiwiBill `json:"bill"`
	Version string   `json:"version"`
}

type Bill struct {
	Amount             Amount `json:"amount"`
	Comment            string `json:"comment"`
	ExpirationDateTime string `json:"expirationDateTime"`
}

func (b *Bill) toJSON() string {
	arr, err := json.Marshal(b)
	if err != nil {
		log.Println("Error on toJSON Bill: " + err.Error())
	}
	return string(arr)
}
