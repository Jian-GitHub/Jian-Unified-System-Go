package event

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type publisherStore struct {
	event   Event
	payload []byte
	err     error
}

func (s *publisherStore) Append(_ context.Context, domainEvent Event, payload []byte) error {
	s.event = domainEvent
	s.payload = append([]byte(nil), payload...)
	return s.err
}
func (*publisherStore) FetchPending(context.Context, int) ([]OutboxRow, error) { return nil, nil }
func (*publisherStore) MarkPublished(context.Context, []int64) error           { return nil }
func (*publisherStore) MarkFailed(context.Context, int64, string) error        { return nil }

func TestOutboxPublisherAppendsSerializedEvent(t *testing.T) {
	store := &publisherStore{}
	publisher := NewOutboxPublisher(store, nil)
	now := time.Now().UTC()
	domainEvent := AccountRegistered{Base: NewBase(now), AccountID: 7, Email: "alice@example.com"}
	publisher.Publish(domainEvent)
	if store.event.Type() != domainEvent.Type() || !json.Valid(store.payload) {
		t.Fatal("publisher did not append a serialized event")
	}
	var payload AccountRegistered
	if err := json.Unmarshal(store.payload, &payload); err != nil || payload.AccountID != 7 || payload.Email != "alice@example.com" || !payload.OccurredAt().Equal(now) {
		t.Fatal("serialized event lost data")
	}
}

func TestOutboxPublisherReportsAppendFailure(t *testing.T) {
	storeErr := errors.New("database unavailable")
	store := &publisherStore{err: storeErr}
	var reported error
	publisher := NewOutboxPublisher(store, func(_ Event, err error) { reported = err })
	publisher.Publish(GrantRevoked{Base: NewBase(time.Now().UTC()), AccountID: 1, GrantID: 2})
	if !errors.Is(reported, storeErr) {
		t.Fatal("publisher swallowed append failure")
	}
}
