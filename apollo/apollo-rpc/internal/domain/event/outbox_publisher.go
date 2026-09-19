package event

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrOutboxUnavailable = errors.New("event outbox is unavailable")

// FailureHandler receives synchronous serialization or persistence failures.
// Publisher keeps its compatibility signature, so production wiring must
// provide a handler that records the failure for operators.
type FailureHandler func(Event, error)

// OutboxPublisher serializes events and durably appends them before Publish
// returns. External broker delivery remains the relay's responsibility.
type OutboxPublisher struct {
	store     OutboxStore
	timeout   time.Duration
	onFailure FailureHandler
}

func NewOutboxPublisher(store OutboxStore, onFailure FailureHandler) *OutboxPublisher {
	return &OutboxPublisher{store: store, timeout: 5 * time.Second, onFailure: onFailure}
}

func (p *OutboxPublisher) Publish(domainEvent Event) {
	if p == nil || p.store == nil {
		p.report(domainEvent, ErrOutboxUnavailable)
		return
	}
	payload, err := json.Marshal(domainEvent)
	if err != nil {
		p.report(domainEvent, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if err = p.store.Append(ctx, domainEvent, payload); err != nil {
		p.report(domainEvent, err)
	}
}

func (p *OutboxPublisher) report(domainEvent Event, err error) {
	if p != nil && p.onFailure != nil {
		p.onFailure(domainEvent, err)
	}
}
