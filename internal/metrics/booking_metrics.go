package metrics

import "github.com/prometheus/client_golang/prometheus"

type BookingMetrics struct {
	createdCounter prometheus.Counter
	updatedCounter prometheus.Counter
}

func NewBookingMetrics(registry prometheus.Registerer) *BookingMetrics {
	createdCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "booking_created_total",
		Help: "Total number of created bookings.",
	})

	updatedCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "booking_updated_total",
		Help: "Total number of updated bookings.",
	})

	registry.MustRegister(createdCounter, updatedCounter)

	return &BookingMetrics{
		createdCounter: createdCounter,
		updatedCounter: updatedCounter,
	}
}

func (m *BookingMetrics) IncrementCreated() {
	if m == nil {
		return
	}

	m.createdCounter.Inc()
}

func (m *BookingMetrics) IncrementUpdated() {
	if m == nil {
		return
	}

	m.updatedCounter.Inc()
}
