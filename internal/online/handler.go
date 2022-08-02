package online

import (
	"encoding/json"
	"errors"
	"hyneo-payment/internal/config"
	"hyneo-payment/internal/handlers"
	"hyneo-payment/internal/model"
	"hyneo-payment/pkg/mysql"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	client *mysql.Client
	config *config.Config
}

func NewOnlineHandler(client *mysql.Client, config *config.Config) handlers.Handler {
	return &handler{
		client: client,
		config: config,
	}
}

func (h *handler) Register(router *gin.Engine, auth *gin.RouterGroup) {
	router.GET("/online", h.getOnline)
}

type status struct {
	Players struct {
		Max    int
		Online int
	}
	Version struct {
		Name     string
		Protocol int
	}
	Delay time.Duration
}

func (h *handler) getOnline(ctx *gin.Context) {
	resp, delay, err := bot.PingAndList(h.config.IP + ":25565")
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	var s status
	err = json.Unmarshal(resp, &s)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "bad request",
		})
		return
	}
	s.Delay = delay
	var max model.Online
	err = h.client.DB.Model(&model.Online{}).First(&max).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			max.Max = s.Players.Online
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
	if max.Max < s.Players.Online {
		max.Max = s.Players.Online
		err = h.client.DB.Model(&max).Update("max", s.Players.Online).Error
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
	}
	ctx.JSON(200, gin.H{
		"online": s.Players.Online,
		"slots":  s.Players.Max,
		"max":    max.Max,
	})
}
