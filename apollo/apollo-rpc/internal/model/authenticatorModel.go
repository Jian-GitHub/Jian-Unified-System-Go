package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthenticatorModel = (*customAuthenticatorModel)(nil)

type (
	// AuthenticatorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthenticatorModel.
	AuthenticatorModel interface {
		authenticatorModel
	}

	customAuthenticatorModel struct {
		*defaultAuthenticatorModel
	}
)

// NewAuthenticatorModel returns a model for the database table.
func NewAuthenticatorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthenticatorModel {
	return &customAuthenticatorModel{
		defaultAuthenticatorModel: newAuthenticatorModel(conn, c, opts...),
	}
}
