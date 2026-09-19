package outbox

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"

	"github.com/zeromicro/go-zero/core/logx"
)

// LoggerRelay is the development relay. It logs delivery metadata only; event
// payloads may contain personal data and are intentionally excluded.
type LoggerRelay struct{}

func (LoggerRelay) Publish(_ context.Context, row event.OutboxRow) error {
	logx.Infof("outbox event delivered: id=%s type=%s schema_version=%d", row.EventID, row.EventType, row.SchemaVersion)
	return nil
}
