package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
)

type relayStore struct {
	pending   []event.OutboxRow
	published []int64
	failed    map[int64]string
}

func (*relayStore) Append(context.Context, event.Event, []byte) error { return nil }
func (s *relayStore) FetchPending(_ context.Context, batch int) ([]event.OutboxRow, error) {
	if batch > len(s.pending) {
		batch = len(s.pending)
	}
	return append([]event.OutboxRow(nil), s.pending[:batch]...), nil
}
func (s *relayStore) MarkPublished(_ context.Context, ids []int64) error {
	s.published = append(s.published, ids...)
	return nil
}
func (s *relayStore) MarkFailed(_ context.Context, id int64, failure string) error {
	if s.failed == nil {
		s.failed = map[int64]string{}
	}
	s.failed[id] = failure
	return nil
}

type selectiveRelay struct{ fail int64 }

func (r selectiveRelay) Publish(_ context.Context, row event.OutboxRow) error {
	if row.ID == r.fail {
		return errors.New("downstream unavailable")
	}
	return nil
}

func TestRelayBatchAcknowledgesOnlySuccessfulRows(t *testing.T) {
	store := &relayStore{pending: []event.OutboxRow{{ID: 1}, {ID: 2}}}
	published, err := RelayBatch(context.Background(), store, selectiveRelay{fail: 2}, 10)
	if published != 1 || err == nil {
		t.Fatalf("relay result = (%d, %v), want one success and one failure", published, err)
	}
	if len(store.published) != 1 || store.published[0] != 1 || store.failed[2] == "" {
		t.Fatal("relay acknowledged failed row or lost failure state")
	}
}

func TestRelayLoopStopsWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		RelayLoop(ctx, &relayStore{}, selectiveRelay{}, time.Millisecond, 1, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("relay loop ignored cancellation")
	}
}

func TestOutboxIdentifierValidation(t *testing.T) {
	for _, valid := range []string{"event_outbox", "Outbox2"} {
		if !sqlIdentifier(valid) {
			t.Fatalf("valid identifier rejected: %q", valid)
		}
	}
	for _, invalid := range []string{"", "event-outbox", "event_outbox;DROP", string(make([]byte, 65))} {
		if sqlIdentifier(invalid) {
			t.Fatalf("invalid identifier accepted: %q", invalid)
		}
	}
}
