package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	bookingdomain "hotelService/internal/domain/booking"
	bookingdto "hotelService/internal/dto/booking"
	"hotelService/internal/service"
)

type BookingHandler struct {
	service *service.BookingService
}

func NewBookingHandler(service *service.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/bookings/:id", h.getByID)
	router.GET("/bookings/available-rooms", h.getAvailableRooms)
	router.POST("/bookings", h.create)
	router.PUT("/bookings/:id", h.update)
	router.DELETE("/bookings/:id", h.delete)
}

// getByID godoc
// @Summary Get booking by ID
// @Description Returns booking with room and guests
// @Tags bookings
// @Produce json
// @Param id path string true "Booking UUID"
// @Success 200 {object} booking.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) getByID(c *gin.Context) {
	response, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// getAvailableRooms godoc
// @Summary Get available rooms
// @Description Returns rooms available in requested time range
// @Tags bookings
// @Produce json
// @Param start query string true "Start time RFC3339" example(2026-04-26T10:00:00Z)
// @Param end query string true "End time RFC3339" example(2026-04-26T12:00:00Z)
// @Success 200 {array} room.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /bookings/available-rooms [get]
func (h *BookingHandler) getAvailableRooms(c *gin.Context) {
	response, err := h.service.FindAvailableRooms(c.Request.Context(), c.Query("start"), c.Query("end"))
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// create godoc
// @Summary Create booking
// @Description Creates booking and links guests to room
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body booking.Request true "Booking payload"
// @Success 201 {object} booking.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) create(c *gin.Context) {
	var request bookingdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// update godoc
// @Summary Update booking
// @Description Updates booking room, guests and time range
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking UUID"
// @Param request body booking.Request true "Booking payload"
// @Success 200 {object} booking.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /bookings/{id} [put]
func (h *BookingHandler) update(c *gin.Context) {
	var request bookingdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// delete godoc
// @Summary Delete booking
// @Description Deletes booking and linked booking_guest rows
// @Tags bookings
// @Produce json
// @Param id path string true "Booking UUID"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /bookings/{id} [delete]
func (h *BookingHandler) delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking deleted successfully"})
}

func writeBookingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, bookingdomain.ErrInvalidBookingID),
		errors.Is(err, bookingdomain.ErrInvalidRoomID),
		errors.Is(err, bookingdomain.ErrInvalidTimeRange),
		errors.Is(err, bookingdomain.ErrRoomCapacityExceeded),
		errors.Is(err, bookingdomain.ErrRoomNotAvailable),
		errors.Is(err, bookingdomain.ErrGuestNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, bookingdomain.ErrNotFound), errors.Is(err, bookingdomain.ErrRoomNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		log.Printf("booking handler error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
