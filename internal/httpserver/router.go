package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"hotelService/internal/config"
	"hotelService/internal/handler"
)

func NewRouter(
	cfg config.Config,
	guestHandler *handler.GuestHandler,
	roomHandler *handler.RoomHandler,
	bookingHandler *handler.BookingHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"application": cfg.App.Name,
			"environment": cfg.App.Env,
		})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	guestHandler.RegisterRoutes(router)
	roomHandler.RegisterRoutes(router)
	bookingHandler.RegisterRoutes(router)

	return router
}
