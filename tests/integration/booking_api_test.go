package integration_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestAPI_BookingFlow_CreateGetUpdateAvailability(t *testing.T) {
	router := newTestRouter(t)

	guestID := createGuest(t, router)
	secondGuestID := createGuestWithPhone(t, router, "+79991234568")
	roomID := createRoom(t, router)
	secondRoomID := createRoomWithNumber(t, router, "306-"+uuid.NewString()[:8])
	bookingID := createBooking(t, router, []string{guestID}, roomID, bookingWindow(10, 12))

	response := performRequest(t, router, http.MethodGet, "/bookings/"+bookingID, nil, http.StatusOK)
	if response["id"] != bookingID {
		t.Fatalf("expected booking id %s, got %+v", bookingID, response)
	}

	updateBody := map[string]any{
		"guestIds":  []string{guestID, secondGuestID},
		"roomId":    secondRoomID,
		"startTime": bookingTime(13),
		"endTime":   bookingTime(15),
	}
	performRequest(t, router, http.MethodPut, "/bookings/"+bookingID, updateBody, http.StatusOK)

	availableRooms := performArrayRequest(
		t,
		router,
		http.MethodGet,
		"/bookings/available-rooms?start="+bookingTime(13)+"&end="+bookingTime(14),
		http.StatusOK,
	)
	if containsRoomID(availableRooms, secondRoomID) {
		t.Fatalf("expected room %s to be unavailable after booking update", secondRoomID)
	}
}

func TestAPI_BookingRejectsOverlap(t *testing.T) {
	router := newTestRouter(t)

	guestID := createGuest(t, router)
	secondGuestID := createGuestWithPhone(t, router, "+79991234569")
	roomID := createRoom(t, router)

	createBooking(t, router, []string{guestID}, roomID, bookingWindow(10, 12))

	response := performRequest(t, router, http.MethodPost, "/bookings", map[string]any{
		"guestIds":  []string{secondGuestID},
		"roomId":    roomID,
		"startTime": bookingTime(11),
		"endTime":   bookingTime(13),
	}, http.StatusBadRequest)

	if response["error"] != "room is not available for selected time range" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestAPI_BookingRejectsUnknownGuest(t *testing.T) {
	router := newTestRouter(t)

	roomID := createRoom(t, router)

	response := performRequest(t, router, http.MethodPost, "/bookings", map[string]any{
		"guestIds":  []string{uuid.NewString()},
		"roomId":    roomID,
		"startTime": bookingTime(10),
		"endTime":   bookingTime(12),
	}, http.StatusBadRequest)

	if response["error"] != "one or more guests not found" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}
