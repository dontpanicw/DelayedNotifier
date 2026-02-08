package input

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
)

type createNotificationRequest struct {
	Text           string `json:"text"`
	ScheduledAt    string `json:"scheduled_at"`
	UserID         uint32 `json:"user_id"`
	TelegramChatID uint32 `json:"telegram_chat_id"`
}

func (s *Server) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		http.Error(w, "invalid scheduled_at, use RFC3339", http.StatusBadRequest)
		return
	}

	msg := domain.Message{
		Text:           req.Text,
		ScheduledAt:    scheduledAt,
		UserId:         req.UserID,
		TelegramChatId: req.TelegramChatID,
	}

	id, err := s.uc.CreateAndSendMessage(r.Context(), msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	messages, err := s.uc.ListMessages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(messages)
}

func (s *Server) handleGetNotificationStatus(w http.ResponseWriter, r *http.Request, id string) {
	status, err := s.uc.GetMessageStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (s *Server) handleDeleteNotification(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.uc.DeleteMessage(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
