package securitylogic

import (
	"context"
	"errors"
	"jian-unified-system/jus-core/data/mysql/model"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateSubsystemTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateSubsystemTokenLogic {
	return &ValidateSubsystemTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证子系统令牌
func (l *ValidateSubsystemTokenLogic) ValidateSubsystemToken(in *apollo.ValidateSubsystemTokenReq) (*apollo.ValidateSubsystemTokenResp, error) {
	// resp		err
	// 1		0	-> system ok
	// 0		1	-> system err
	// 1		1	-> system ok	not exists
	token, err := l.svcCtx.TokenModel.FindOne(l.ctx, in.TokenId, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &apollo.ValidateSubsystemTokenResp{
				Validated: false,
			}, err
		}
		return nil, err
	}

	if token.IsEnabled == 0 || token.IsDeleted == 1 {
		return &apollo.ValidateSubsystemTokenResp{
			Validated: false,
		}, nil
	}

	return &apollo.ValidateSubsystemTokenResp{
		Validated: true,
	}, nil
}
