package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PasskeysModel = (*customPasskeysModel)(nil)

type (
	// PasskeysModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPasskeysModel.
	PasskeysModel interface {
		passkeysModel
	}

	customPasskeysModel struct {
		*defaultPasskeysModel
	}
)

// NewPasskeysModel returns a model for the database table.
func NewPasskeysModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PasskeysModel {
	return &customPasskeysModel{
		defaultPasskeysModel: newPasskeysModel(conn, c, opts...),
	}
}
