package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"hyneo-payment/internal/config"
	"strings"
)

type middleware struct {
	config *config.Config
}

func NewMiddleware(config *config.Config) *middleware {
	return &middleware{
		config: config,
	}
}

type authHeader struct {
	IDToken string `header:"Authorization"`
}

func (m *middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h1 := authHeader{}
		if err := ctx.ShouldBindHeader(&h1); err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
		idTokenHeader := strings.Split(h1.IDToken, "Bearer ")
		if len(idTokenHeader) != 2 {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
		idToken := idTokenHeader[1]
		token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(m.config.SECRET), nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		} else {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "bad request",
			})
			return
		}
	}
}
