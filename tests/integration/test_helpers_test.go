package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"hotelService/internal/config"
	"hotelService/internal/handler"
	"hotelService/internal/httpserver"
	"hotelService/internal/metrics"
	"hotelService/internal/repository"
	"hotelService/internal/service"
	"hotelService/internal/testutil"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	db := testutil.StartPostgres(t)
	guestRepository := repository.NewGuestRepository(db)
	roomRepository := repository.NewRoomRepository(db)
	bookingRepository := repository.NewBookingRepository(db)

	guestHandler := handler.NewGuestHandler(service.NewGuestService(guestRepository, nil))
	roomHandler := handler.NewRoomHandler(service.NewRoomService(roomRepository, nil))
	bookingHandler := handler.NewBookingHandler(service.NewBookingService(
		bookingRepository,
		guestRepository,
		roomRepository,
		nil,
		metrics.NewBookingMetrics(prometheus.NewRegistry()),
	))

	return httpserver.NewRouter(config.Config{
		App: config.AppConfig{
			Name: "hotel-service-test",
			Env:  "test",
		},
	}, guestHandler, roomHandler, bookingHandler)
}

func createGuest(t *testing.T, router http.Handler) string {
	t.Helper()
	return createGuestWithPhone(t, router, "+79991234567")
}

func createGuestWithPhone(t *testing.T, router http.Handler, phone string) string {
	t.Helper()

	response := performRequest(t, router, http.MethodPost, "/guests", map[string]any{
		"lastName":    "Ivanov",
		"firstName":   "Ivan",
		"birthDate":   "1990-05-12",
		"phoneNumber": phone,
	}, http.StatusCreated)

	return extractID(t, response)
}

func createRoom(t *testing.T, router http.Handler) string {
	t.Helper()
	return createRoomWithNumber(t, router, "305-"+uuid.NewString()[:8])
}

func createRoomWithNumber(t *testing.T, router http.Handler, number string) string {
	t.Helper()

	response := performRequest(t, router, http.MethodPost, "/rooms", map[string]any{
		"floor":    3,
		"number":   number,
		"capacity": 2,
	}, http.StatusCreated)

	return extractID(t, response)
}

func createBooking(t *testing.T, router http.Handler, guestIDs []string, roomID string, window map[string]any) string {
	t.Helper()

	response := performRequest(t, router, http.MethodPost, "/bookings", map[string]any{
		"guestIds":  guestIDs,
		"roomId":    roomID,
		"startTime": window["startTime"],
		"endTime":   window["endTime"],
	}, http.StatusCreated)

	return extractID(t, response)
}

func bookingWindow(startHour, endHour int) map[string]any {
	return map[string]any{
		"startTime": bookingTime(startHour),
		"endTime":   bookingTime(endHour),
	}
}

func bookingTime(hour int) string {
	return time.Date(2026, time.April, 26, hour, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func performRequest(
	t *testing.T,
	router http.Handler,
	method, path string,
	body map[string]any,
	expectedStatus int,
) map[string]any {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != expectedStatus {
		t.Fatalf("expected %d, got %d: %s", expectedStatus, recorder.Code, recorder.Body.String())
	}

	var response map[string]any
	if len(recorder.Body.Bytes()) > 0 {
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("unmarshal response body: %v", err)
		}
	}

	return response
}

func performArrayRequest(
	t *testing.T,
	router http.Handler,
	method, path string,
	expectedStatus int,
) []map[string]any {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != expectedStatus {
		t.Fatalf("expected %d, got %d: %s", expectedStatus, recorder.Code, recorder.Body.String())
	}

	var response []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	return response
}

func extractID(t *testing.T, response map[string]any) string {
	t.Helper()

	id, ok := response["id"].(string)
	if !ok || id == "" {
		t.Fatalf("response does not contain id: %+v", response)
	}

	return id
}

func containsRoomID(rooms []map[string]any, roomID string) bool {
	for _, room := range rooms {
		if room["id"] == roomID {
			return true
		}
	}
	return false
}
