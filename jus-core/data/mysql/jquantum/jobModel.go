package jquantum

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ JobModel = (*customJobModel)(nil)

type (
	// JobModel is an interface to be customized, add more methods here,
	// and implement the added methods in customJobModel.
	JobModel interface {
		jobModel
	}

	customJobModel struct {
		*defaultJobModel
	}
)

// NewJobModel returns a apollo for the database table.
func NewJobModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) JobModel {
	return &customJobModel{
		defaultJobModel: newJobModel(conn, c, opts...),
	}
}
