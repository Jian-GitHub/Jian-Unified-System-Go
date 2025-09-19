package securitylogic

import (
	"context"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"strconv"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindTenSubsystemTokensLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindTenSubsystemTokensLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTenSubsystemTokensLogic {
	return &FindTenSubsystemTokensLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FindTenSubsystemTokens 查询 10 个子系统令牌
func (l *FindTenSubsystemTokensLogic) FindTenSubsystemTokens(in *apollo.FindTenSubsystemTokensReq) (*apollo.FindTenSubsystemTokensResp, error) {
	var tokens *[]ap.Token
	tokens, err := l.svcCtx.TokenModel.FindBatch(l.ctx, in.UserId, in.Page)
	if err != nil {
		return nil, err
	}

	// 整理 tokens
	var subsystemTokens []*apollo.SubsystemToken
	if tokens != nil {
		subsystemTokens = make([]*apollo.SubsystemToken, 0, len(*tokens))
		for _, token := range *tokens {
			subsystemTokens = append(subsystemTokens, &apollo.SubsystemToken{
				Id:    strconv.FormatInt(token.Id, 10),
				Value: token.Value,
				Name:  token.Name.String,
				Year:  int64(token.CreateTime.Year()),
				Month: int64(token.CreateTime.Month()),
				Day:   int64(token.CreateTime.Day()),
			})
		}
	} else {
		subsystemTokens = make([]*apollo.SubsystemToken, 0)
	}

	return &apollo.FindTenSubsystemTokensResp{
		Tokens: subsystemTokens,
	}, nil
}
