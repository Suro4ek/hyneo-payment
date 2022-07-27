package getpay

import (
	"crypto/md5"
	"encoding/hex"
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
	r.POST("/getpay/", h.bill)
}

func (h *handler) getpay(ctx *gin.Context) {
	var dto GetPay
	if err := ctx.ShouldBindQuery(&dto); err != nil {
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
	hash := GetMD5Hash(method.PublicKey + ":" + dto.Amount + ":" + dto.Merchant_order_id + ":" + method.SecretKey)
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
		Method:    method.Method.Name,
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?secret=%s&wallet=%s&order=%d&resultUrl=%s&backUrl=%s&comment=%s&sum=%d",
		urlBill,
		method.SecretKey,
		method.PublicKey,
		order.ID,
		"http://api.hyneo.ru/getpay/",
		"https://hyneo.ru/",
		"Оплата заказа "+item.Name,
		price,
	), nil)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
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
