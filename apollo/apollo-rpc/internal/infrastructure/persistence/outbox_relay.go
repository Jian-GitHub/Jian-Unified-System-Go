package persistence

import (
	"context"
	"errors"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
)

// RelayBatch publishes one bounded page. Successful rows are acknowledged in
// one update; failed rows remain pending and record their latest error.
func RelayBatch(ctx context.Context, store event.OutboxStore, relay event.Relay, batch int) (int, error) {
	if store == nil || relay == nil || batch <= 0 || batch > 1000 {
		return 0, ErrInvalidOutbox
	}
	rows, err := store.FetchPending(ctx, batch)
	if err != nil {
		return 0, err
	}
	published := make([]int64, 0, len(rows))
	var failures error
	for _, row := range rows {
		if err = relay.Publish(ctx, row); err != nil {
			if markErr := store.MarkFailed(ctx, row.ID, err.Error()); markErr != nil {
				failures = errors.Join(failures, err, markErr)
			} else {
				failures = errors.Join(failures, err)
			}
			continue
		}
		published = append(published, row.ID)
	}
	if err = store.MarkPublished(ctx, published); err != nil {
		failures = errors.Join(failures, err)
	}
	return len(published), failures
}

// RelayLoop polls until ctx is canceled. It processes immediately on startup,
// then waits for the configured interval between bounded batches.
func RelayLoop(ctx context.Context, store event.OutboxStore, relay event.Relay, interval time.Duration, batch int, onError func(error)) {
	if interval <= 0 || batch <= 0 || batch > 1000 {
		if onError != nil {
			onError(ErrInvalidOutbox)
		}
		return
	}
	for {
		if _, err := RelayBatch(ctx, store, relay, batch); err != nil && !errors.Is(err, context.Canceled) && onError != nil {
			onError(err)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
