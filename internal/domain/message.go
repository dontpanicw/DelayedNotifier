package domain

import "time"

const (
	JobStatusScheduled        = "Scheduled"
	JobStatusSent             = "Sent"
	JobStatusFailed           = "Failed"
	JobStatusTerminallyFailed = "Terminally_Failed"
)

type Message struct {
	Id             string    `json:"id"`
	Text           string    `json:"text"`
	Status         string    `json:"status"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	UserId         uint32    `json:"user_id"`
	TelegramChatId uint32    `json:"telegram_chat_id"`
}
