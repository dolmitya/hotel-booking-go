package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI_Health(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAPI_GuestCreateAndGet(t *testing.T) {
	router := newTestRouter(t)

	guestID := createGuest(t, router)
	response := performRequest(t, router, http.MethodGet, "/guests/"+guestID, nil, http.StatusOK)

	if response["id"] != guestID {
		t.Fatalf("expected guest id %s, got %+v", guestID, response)
	}
}

func TestAPI_GuestRejectsInvalidPayload(t *testing.T) {
	router := newTestRouter(t)

	response := performRequest(t, router, http.MethodPost, "/guests", map[string]any{
		"lastName":    "Ivanov",
		"firstName":   "Ivan",
		"birthDate":   "1990/05/12",
		"phoneNumber": "+79991234567",
	}, http.StatusBadRequest)

	if response["error"] == "" {
		t.Fatalf("expected error response, got %+v", response)
	}
}
