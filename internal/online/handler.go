package online

import (
	"errors"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/mysql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/millkhan/mcstatusgo"
	"gorm.io/gorm"
)

type handler struct {
	client *mysql.Client
}

func NewOnlineHandler(client *mysql.Client) handlers.Handler {
	return &handler{
		client: client,
	}
}

func (h *handler) Register(router *gin.Engine, auth *gin.RouterGroup) {
	router.GET("/online", h.getOnline)
}

func (h *handler) getOnline(ctx *gin.Context) {
	server, err := mcstatusgo.Status("mc.hyneo.ru", 25565, 10*time.Second, 5*time.Second)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var max model.Online
	err = h.client.DB.Model(&model.Online{}).First(&max).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			max.Max = server.Players.Online
			err = h.client.DB.Save(&max).Error
			if err != nil {
				ctx.AbortWithStatusJSON(400, gin.H{
					"error": "bad request",
				})
				return
			}
		} else {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
	}
	if max.Max < server.Players.Online {
		max.Max = server.Players.Online
		err = h.client.DB.Save(&max).Error
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
	}
	ctx.JSON(200, gin.H{
		"online": server.Players.Online,
		"slots":  server.Players.Max,
		"max":    max.Max,
	})
}
