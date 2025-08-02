package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ThirdPartyModel = (*customThirdPartyModel)(nil)

type (
	// ThirdPartyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customThirdPartyModel.
	ThirdPartyModel interface {
		thirdPartyModel
	}

	customThirdPartyModel struct {
		*defaultThirdPartyModel
	}
)

// NewThirdPartyModel returns a model for the database table.
func NewThirdPartyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ThirdPartyModel {
	return &customThirdPartyModel{
		defaultThirdPartyModel: newThirdPartyModel(conn, c, opts...),
	}
}
