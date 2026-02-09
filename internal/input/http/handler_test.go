package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
)

type usecasesMock struct {
	createCalled bool
	createdMsg   domain.Message

	listResult   []domain.Message
	statusByID   map[string]string
	deleteCalled bool
}

func (u *usecasesMock) CreateAndSendMessage(ctx context.Context, message domain.Message) (string, error) {
	u.createCalled = true
	u.createdMsg = message
	return "generated-id", nil
}

func (u *usecasesMock) GetMessageStatus(ctx context.Context, id string) (string, error) {
	if s, ok := u.statusByID[id]; ok {
		return s, nil
	}
	return "", http.ErrNoLocation
}

func (u *usecasesMock) ListMessages(ctx context.Context) ([]domain.Message, error) {
	return u.listResult, nil
}

func (u *usecasesMock) DeleteMessage(ctx context.Context, id string) error {
	u.deleteCalled = true
	return nil
}

var _ port.Usecases = (*usecasesMock)(nil)

func TestHandleCreateNotification_OK(t *testing.T) {
	uc := &usecasesMock{}
	srv := NewServer(uc)

	body := map[string]any{
		"text":             "hello",
		"scheduled_at":     time.Now().Format(time.RFC3339),
		"user_id":          1,
		"telegram_chat_id": 42,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewReader(data))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if !uc.createCalled {
		t.Fatalf("expected usecase CreateAndSendMessage to be called")
	}

	var resp map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["id"] == "" {
		t.Fatalf("expected id in response")
	}
}

func TestHandleCreateNotification_InvalidJSON(t *testing.T) {
	uc := &usecasesMock{}
	srv := NewServer(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewBufferString("{invalid-json"))
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if uc.createCalled {
		t.Fatalf("usecase must not be called on invalid JSON")
	}
}

func TestHandleListNotifications_OK(t *testing.T) {
	uc := &usecasesMock{
		listResult: []domain.Message{
			{Id: "1", Text: "t1"},
			{Id: "2", Text: "t2"},
		},
	}
	srv := NewServer(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []domain.Message
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(resp))
	}
}

func TestHandleGetNotificationStatus_OK(t *testing.T) {
	uc := &usecasesMock{
		statusByID: map[string]string{"abc": "Scheduled"},
	}
	srv := NewServer(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/abc/status", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "Scheduled" {
		t.Fatalf("expected status Scheduled, got %s", resp["status"])
	}
}

func TestHandleDeleteNotification_OK(t *testing.T) {
	uc := &usecasesMock{}
	srv := NewServer(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/xyz", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if !uc.deleteCalled {
		t.Fatalf("expected DeleteMessage to be called")
	}
}

