package securitylogic

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/jsonx"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/system"
	"jian-unified-system/jus-core/util"
	"strconv"
	"time"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSubsystemTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSubsystemTokenLogic {
	return &GenerateSubsystemTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 生成子系统令牌
func (l *GenerateSubsystemTokenLogic) GenerateSubsystemToken(in *apollo.GenerateSubsystemTokenReq) (*apollo.GenerateSubsystemTokenResp, error) {
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

	now := time.Now()
	return &apollo.GenerateSubsystemTokenResp{
		Token: &apollo.SubsystemToken{
			Id:    strconv.FormatInt(tokenId, 10),
			Value: token,
			Name:  in.Name,
			Year:  int64(now.Year()),
			Month: int64(now.Month()),
			Day:   int64(now.Day()),
		},
	}, nil
}
