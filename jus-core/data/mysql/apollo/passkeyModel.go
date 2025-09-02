package apollo

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PasskeyModel = (*customPasskeyModel)(nil)

type (
	// PasskeyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPasskeyModel.
	PasskeyModel interface {
		passkeyModel
	}

	customPasskeyModel struct {
		*defaultPasskeyModel
	}
)

// NewPasskeyModel returns a apollo for the database table.
func NewPasskeyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PasskeyModel {
	return &customPasskeyModel{
		defaultPasskeyModel: newPasskeyModel(conn, c, opts...),
	}
}
