package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	roomdomain "hotelService/internal/domain/room"
	roomdto "hotelService/internal/dto/room"
	"hotelService/internal/service"
)

type RoomHandler struct {
	service *service.RoomService
}

func NewRoomHandler(service *service.RoomService) *RoomHandler {
	return &RoomHandler{service: service}
}

func (h *RoomHandler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/rooms/:id", h.getByID)
	router.POST("/rooms", h.create)
	router.PUT("/rooms/:id", h.update)
	router.DELETE("/rooms/:id", h.delete)
}

// getByID godoc
// @Summary Get room by ID
// @Description Returns room data by unique identifier
// @Tags rooms
// @Produce json
// @Param id path string true "Room UUID"
// @Success 200 {object} room.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /rooms/{id} [get]
func (h *RoomHandler) getByID(c *gin.Context) {
	response, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeRoomError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// create godoc
// @Summary Create room
// @Description Creates a new room
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body room.Request true "Room payload"
// @Success 201 {object} room.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /rooms [post]
func (h *RoomHandler) create(c *gin.Context) {
	var request roomdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// update godoc
// @Summary Update room
// @Description Updates an existing room
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path string true "Room UUID"
// @Param request body room.Request true "Room payload"
// @Success 200 {object} room.Response
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /rooms/{id} [put]
func (h *RoomHandler) update(c *gin.Context) {
	var request roomdto.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// delete godoc
// @Summary Delete room
// @Description Deletes a room by unique identifier
// @Tags rooms
// @Produce json
// @Param id path string true "Room UUID"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /rooms/{id} [delete]
func (h *RoomHandler) delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeRoomError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}

func writeRoomError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, roomdomain.ErrInvalidRoomID),
		errors.Is(err, roomdomain.ErrInvalidFloor),
		errors.Is(err, roomdomain.ErrInvalidCapacity):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, roomdomain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		log.Printf("room handler error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
