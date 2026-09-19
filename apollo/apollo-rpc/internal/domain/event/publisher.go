// Package event defines the domain events that bounded contexts publish
// when state of an aggregate changes. Application code collects and
// forwards them to a Publisher port; no transport or persistence detail
// leaks here.
package event

import (
	"context"
	"errors"
	"time"
)

// ErrUnsupportedUpgrade is returned by NoopUpgrader and any Upgrader
// that does not know how to migrate from one schema version to another.
var ErrUnsupportedUpgrade = errors.New("unsupported event schema upgrade")

// Publisher is the application-layer port for forwarding domain events
// to whatever downstream subscribes (outbox table, log, broker). The
// default no-op implementation simply drops events so the bounded
// context keeps working without a publisher wired up.
type Publisher interface {
	Publish(Event)
}

// NoopPublisher discards every event. Useful as a default in tests and
// for the legacy wiring that has not opted into the publisher yet.
type NoopPublisher struct{}

// Publish satisfies Publisher by dropping the event.
func (NoopPublisher) Publish(Event) {}

// OutboxStore is the persistence port for the durable outbox table. Appends
// are synchronous, but they are not atomic with aggregate writes unless a
// deployment supplies a shared unit-of-work adapter. The application layer
// depends on this interface; the
// MySQL-backed implementation lives under infrastructure/persistence.
type OutboxStore interface {
	Append(ctx context.Context, e Event, payload []byte) error
	FetchPending(ctx context.Context, batch int) ([]OutboxRow, error)
	MarkPublished(ctx context.Context, ids []int64) error
	MarkFailed(ctx context.Context, id int64, err string) error
}

// OutboxRow is the projection of a pending outbox row read by a relay.
// It mirrors the storage shape; relay implementations should treat
// each row as opaque until they decide where to forward it.
type OutboxRow struct {
	ID            int64
	EventID       string
	EventType     string
	SchemaVersion int
	Payload       []byte
	OccurredAt    time.Time
}

// Relay forwards an outbox row to whatever downstream subscribes
// (NATS, Kafka, structured log, another service). Production
// deployments inject a real implementation; tests use NoopRelay or a
// recording fake.
type Relay interface {
	Publish(ctx context.Context, row OutboxRow) error
}

// NoopRelay succeeds without doing anything. Useful when the outbox
// table is kept purely for audit / replay and downstream delivery is
// not required (e.g. local development).
type NoopRelay struct{}

// Publish satisfies Relay.
func (NoopRelay) Publish(context.Context, OutboxRow) error { return nil }

// Upgrader migrates a serialized event payload from one schema version
// to another. The default NoopUpgrader refuses any upgrade so callers
// either use the original payload or fail loudly.
type Upgrader interface {
	Upgrade(eventType string, payload []byte, fromVersion, toVersion int) ([]byte, error)
}

// NoopUpgrader rejects every upgrade. Production deployments inject an
// implementation that maps known schemas forward.
type NoopUpgrader struct{}

// Upgrade satisfies Upgrader by refusing every request.
func (NoopUpgrader) Upgrade(string, []byte, int, int) ([]byte, error) {
	return nil, ErrUnsupportedUpgrade
}
