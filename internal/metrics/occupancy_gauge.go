package metrics

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"hotelService/internal/repository"
)

func RegisterOccupancyGauge(
	registry prometheus.Registerer,
	bookingRepository *repository.BookingRepository,
	roomRepository *repository.RoomRepository,
) {
	registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "room_occupancy_percent",
			Help: "Current hotel occupancy percentage.",
		},
		func() float64 {
			ctx := context.Background()

			totalRooms, err := roomRepository.CountAll(ctx)
			if err != nil {
				log.Printf("collect occupancy total rooms: %v", err)
				return 0
			}
			if totalRooms == 0 {
				return 0
			}

			activeBookings, err := bookingRepository.CountActiveBookingsNow(ctx)
			if err != nil {
				log.Printf("collect occupancy active bookings: %v", err)
				return 0
			}

			return (float64(activeBookings) / float64(totalRooms)) * 100
		},
	))
}
