package main

import (
	"context"
	"log"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	_ "hotelService/docs"
	"hotelService/internal/config"
	"hotelService/internal/db"
	"hotelService/internal/handler"
	"hotelService/internal/httpserver"
	"hotelService/internal/kafka"
	"hotelService/internal/metrics"
	"hotelService/internal/repository"
	"hotelService/internal/service"
)

const configPathEnv = "APP_CONFIG_PATH"

// @title Hotel Booking API
// @version 1.0
// @description API for managing hotel guests, rooms and bookings.
// @host localhost:8080
// @BasePath /
func main() {
	configPath := os.Getenv(configPathEnv)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	postgres, err := db.NewPostgres(context.Background(), cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer postgres.Close()

	kafkaProducer := kafka.NewProducer(cfg.Kafka)
	defer kafkaProducer.Close()
	if err := kafkaProducer.EnsureTopics(context.Background()); err != nil {
		log.Fatalf("ensure kafka topics: %v", err)
	}

	guestRepository := repository.NewGuestRepository(postgres)
	guestService := service.NewGuestService(guestRepository, kafkaProducer)
	guestHandler := handler.NewGuestHandler(guestService)
	roomRepository := repository.NewRoomRepository(postgres)
	roomService := service.NewRoomService(roomRepository, kafkaProducer)
	roomHandler := handler.NewRoomHandler(roomService)
	bookingRepository := repository.NewBookingRepository(postgres)
	bookingMetrics := metrics.NewBookingMetrics(prometheus.DefaultRegisterer)
	metrics.RegisterOccupancyGauge(prometheus.DefaultRegisterer, bookingRepository, roomRepository)
	bookingService := service.NewBookingService(
		bookingRepository,
		guestRepository,
		roomRepository,
		kafkaProducer,
		bookingMetrics,
	)
	bookingHandler := handler.NewBookingHandler(bookingService)

	router := httpserver.NewRouter(cfg, guestHandler, roomHandler, bookingHandler)

	log.Printf("starting %s on %s", cfg.App.Name, cfg.HTTPAddress())
	if err := router.Run(cfg.HTTPAddress()); err != nil {
		log.Fatalf("run http server: %v", err)
	}
}
