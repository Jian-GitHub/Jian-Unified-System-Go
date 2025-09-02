package apollo

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ContactModel = (*customContactModel)(nil)

type (
	// ContactModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContactModel.
	ContactModel interface {
		contactModel
	}

	customContactModel struct {
		*defaultContactModel
	}
)

// NewContactModel returns a apollo for the database table.
func NewContactModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ContactModel {
	return &customContactModel{
		defaultContactModel: newContactModel(conn, c, opts...),
	}
}
