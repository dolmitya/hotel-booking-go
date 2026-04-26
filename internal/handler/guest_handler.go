package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	guestdomain "hotelService/internal/domain/guest"
	guestdto "hotelService/internal/dto/guest"
	"hotelService/internal/service"
)

type GuestHandler struct {
	service *service.GuestService
}

func NewGuestHandler(service *service.GuestService) *GuestHandler {
	return &GuestHandler{service: service}
}

func (h *GuestHandler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/guests/:id", h.getByID)
	router.POST("/guests", h.create)
	router.PUT("/guests/:id", h.update)
}

// getByID godoc
// @Summary Get guest by ID
// @Description Returns guest data by unique identifier
// @Tags guests
// @Produce json
// @Param id path string true "Guest UUID"
// @Success 200 {object} guest.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /guests/{id} [get]
func (h *GuestHandler) getByID(c *gin.Context) {
	response, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeGuestError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// create godoc
// @Summary Create guest
// @Description Creates a new guest
// @Tags guests
// @Accept json
// @Produce json
// @Param request body guest.Request true "Guest payload"
// @Success 201 {object} guest.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /guests [post]
func (h *GuestHandler) create(c *gin.Context) {
	var request guestdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		writeGuestError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// update godoc
// @Summary Update guest
// @Description Updates an existing guest
// @Tags guests
// @Accept json
// @Produce json
// @Param id path string true "Guest UUID"
// @Param request body guest.Request true "Guest payload"
// @Success 200 {object} guest.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /guests/{id} [put]
func (h *GuestHandler) update(c *gin.Context) {
	var request guestdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		writeGuestError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func writeGuestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, guestdomain.ErrInvalidGuestID), errors.Is(err, guestdomain.ErrInvalidBirthDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, guestdomain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		log.Printf("guest handler error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
