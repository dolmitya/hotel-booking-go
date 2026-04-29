package integration_test

import (
	"net/http"
	"testing"
)

func TestAPI_RoomCreateUpdateAndGet(t *testing.T) {
	router := newTestRouter(t)

	roomID := createRoom(t, router)

	updateBody := map[string]any{
		"floor":    5,
		"number":   "501",
		"capacity": 3,
	}
	performRequest(t, router, http.MethodPut, "/rooms/"+roomID, updateBody, http.StatusOK)

	response := performRequest(t, router, http.MethodGet, "/rooms/"+roomID, nil, http.StatusOK)
	if int(response["floor"].(float64)) != 5 {
		t.Fatalf("expected updated floor, got %+v", response)
	}
	if response["number"] != "501" {
		t.Fatalf("expected updated number, got %+v", response)
	}
}
