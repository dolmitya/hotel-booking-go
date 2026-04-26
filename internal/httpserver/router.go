package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"hotelService/internal/config"
	"hotelService/internal/handler"
)

func NewRouter(cfg config.Config, guestHandler *handler.GuestHandler, roomHandler *handler.RoomHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"application": cfg.App.Name,
			"environment": cfg.App.Env,
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	guestHandler.RegisterRoutes(router)
	roomHandler.RegisterRoutes(router)

	return router
}
