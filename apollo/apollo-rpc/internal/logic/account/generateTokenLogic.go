package accountlogic

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/jsonx"
	"jian-unified-system/jus-core/util"

	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/system"
)

type GenerateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateToken 生成子系统令牌
func (l *GenerateTokenLogic) GenerateToken(in *apollo.GenerateTokenReq) (*apollo.GenerateTokenResp, error) {
	args := make(map[string]interface{})
	args["id"] = in.UserId
	scopeNum := 0

	var scopes []int
	err := jsonx.Unmarshal(in.Scope, &scopes)
	if err != nil {
		return nil, err
	}

	for _, scope := range scopes {
		if _, ok := system.SubsystemScopes[scope]; ok {
			scopeNum |= scope
		}
	}
	args["scope"] = scopeNum

	tokenId := int64(uuid.New().ID())
	args["tokenId"] = tokenId

	token, err := util.GenToken(l.svcCtx.Config.SubSystem.AccessSecret, l.svcCtx.Config.SubSystem.AccessExpire, args)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.TokenModel.Insert(l.ctx, &ap.Token{
		Id:     tokenId,
		UserId: in.UserId,
		Name: sql.NullString{
			String: in.Name,
			Valid:  true,
		},
		Value:     token,
		IsEnabled: 1,
	})
	if err != nil {
		return nil, err
	}

	return &apollo.GenerateTokenResp{
		Token: token,
	}, nil
}
