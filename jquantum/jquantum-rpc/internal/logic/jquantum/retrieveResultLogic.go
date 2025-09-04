package jquantumlogic

import (
	"context"
	"errors"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"jian-unified-system/jquantum/jquantum-rpc/internal/types/jobResultStatus"
	"os"
	"path/filepath"
	"strconv"

	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrieveResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrieveResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrieveResultLogic {
	return &RetrieveResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RetrieveResult 获取计算任务结果
func (l *RetrieveResultLogic) RetrieveResult(in *jquantum.RetrieveResultReq) (*jquantum.RetrieveResultResp, error) {
	// todo: add your logic here and delete this line
	result, err := l.svcCtx.JobModel.FindOne(l.ctx, in.JobId, in.UserId)
	if err != nil {
		return nil, err
	}

	switch result.Status {
	case jobResultStatus.QUEUED:
		return nil, errors.New("job is queued")
	case jobResultStatus.PROCESSING:
		return nil, errors.New("job is still processing")
	case jobResultStatus.COMPILATION_ERROR:
		return nil, errors.New("job encountered compilation error")
	case jobResultStatus.RUNNING_ERROR:
		return nil, errors.New("job encountered running error")
	case jobResultStatus.FINISHED:
	default:
		panic("unhandled default case")
	}

	path := filepath.Join(l.svcCtx.Config.JQuantum.BaseDir, strconv.FormatInt(in.UserId, 10), in.JobId, "result.json")
	if !fileutil.Exist(path) {
		return nil, errors.New("file not exist")
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &jquantum.RetrieveResultResp{
		Result: bytes,
	}, nil
}
