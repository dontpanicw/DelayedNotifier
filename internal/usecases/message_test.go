package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
)

type repoMock struct {
	createCalled bool
	createdMsg   domain.Message

	statusByID map[string]string
}

func (r *repoMock) CreateMessage(ctx context.Context, message domain.Message) error {
	r.createCalled = true
	r.createdMsg = message
	return nil
}

func (r *repoMock) GetMessageStatus(ctx context.Context, id string) (string, error) {
	if s, ok := r.statusByID[id]; ok {
		return s, nil
	}
	return "", errors.New("not found")
}

func (r *repoMock) ListMessages(ctx context.Context) ([]domain.Message, error) {
	return nil, nil
}

func (r *repoMock) UpdateMessageStatus(ctx context.Context, id, status string) error {
	if r.statusByID == nil {
		r.statusByID = make(map[string]string)
	}
	r.statusByID[id] = status
	return nil
}

func (r *repoMock) DeleteMessage(ctx context.Context, id string) error {
	return nil
}

type queueMock struct {
	sent []domain.Message
	fail bool
}

func (q *queueMock) SendMessage(ctx context.Context, message domain.Message) error {
	if q.fail {
		return errors.New("publish failed")
	}
	q.sent = append(q.sent, message)
	return nil
}

type cacheMock struct {
	values map[string]string
}

func (c *cacheMock) GetStatus(ctx context.Context, id string) (string, error) {
	if c.values == nil {
		return "", nil
	}
	return c.values[id], nil
}

func (c *cacheMock) SetStatus(ctx context.Context, id, status string, ttl time.Duration) error {
	if c.values == nil {
		c.values = make(map[string]string)
	}
	c.values[id] = status
	return nil
}

func TestCreateAndSendMessage_Success(t *testing.T) {
	r := &repoMock{}
	q := &queueMock{}
	c := &cacheMock{}

	uc := NewMessageUsecases(r, q, c)

	msg := domain.Message{
		Text:        "hello",
		UserId:      1,
		ScheduledAt: time.Now().Add(time.Minute),
	}

	id, err := uc.CreateAndSendMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Fatalf("expected non-empty id")
	}
	if !r.createCalled {
		t.Fatalf("expected repository CreateMessage to be called")
	}
	if len(q.sent) != 1 {
		t.Fatalf("expected 1 message sent to queue, got %d", len(q.sent))
	}
	if q.sent[0].Id != r.createdMsg.Id {
		t.Fatalf("expected message id in queue and repo to match")
	}
	if r.createdMsg.Status != domain.JobStatusScheduled {
		t.Fatalf("expected status %s, got %s", domain.JobStatusScheduled, r.createdMsg.Status)
	}
}

func TestCreateAndSendMessage_InvalidUser(t *testing.T) {
	r := &repoMock{}
	q := &queueMock{}
	c := &cacheMock{}

	uc := NewMessageUsecases(r, q, c)

	_, err := uc.CreateAndSendMessage(context.Background(), domain.Message{
		Text:        "hello",
		UserId:      0,
		ScheduledAt: time.Now(),
	})
	if err == nil {
		t.Fatalf("expected error for invalid userId")
	}
	if r.createCalled {
		t.Fatalf("repository must not be called on invalid input")
	}
	if len(q.sent) != 0 {
		t.Fatalf("queue must not be called on invalid input")
	}
}

func TestCreateAndSendMessage_QueueError(t *testing.T) {
	r := &repoMock{}
	q := &queueMock{fail: true}
	c := &cacheMock{}

	uc := NewMessageUsecases(r, q, c)

	_, err := uc.CreateAndSendMessage(context.Background(), domain.Message{
		Text:        "hello",
		UserId:      1,
		ScheduledAt: time.Now(),
	})
	if err == nil {
		t.Fatalf("expected error when queue publish fails")
	}
	if !r.createCalled {
		t.Fatalf("expected repository CreateMessage to be called before queue")
	}
}

func TestGetMessageStatus_UsesCache(t *testing.T) {
	r := &repoMock{
		statusByID: map[string]string{
			"1": "DBStatus",
		},
	}
	c := &cacheMock{
		values: map[string]string{
			"1": "CachedStatus",
		},
	}
	q := &queueMock{}

	uc := NewMessageUsecases(r, q, c)

	status, err := uc.GetMessageStatus(context.Background(), "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "CachedStatus" {
		t.Fatalf("expected status from cache, got %s", status)
	}
}

func TestGetMessageStatus_FillsCacheOnMiss(t *testing.T) {
	r := &repoMock{
		statusByID: map[string]string{
			"1": "DBStatus",
		},
	}
	c := &cacheMock{} // пустой кэш
	q := &queueMock{}

	uc := NewMessageUsecases(r, q, c)

	status, err := uc.GetMessageStatus(context.Background(), "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "DBStatus" {
		t.Fatalf("expected status from DB, got %s", status)
	}
	if cached := c.values["1"]; cached != "DBStatus" {
		t.Fatalf("expected cache to be filled with DB status, got %s", cached)
	}
}
