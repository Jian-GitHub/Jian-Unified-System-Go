package accountlogic

import (
	"context"
	"errors"
	"jian-unified-system/jus-core/data/mysql/model"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateToken 验证子系统令牌
func (l *ValidateTokenLogic) ValidateToken(in *apollo.ValidateTokenReq) (*apollo.ValidateTokenResp, error) {
	// resp		err
	// 1		0	-> system ok
	// 0		1	-> system err
	// 1		1	-> system ok	not exists
	token, err := l.svcCtx.TokenModel.FindOne(l.ctx, in.TokenId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &apollo.ValidateTokenResp{
				Validated: false,
			}, err
		}
		return nil, err
	}

	if token.IsEnabled == 0 || token.IsDeleted == 1 {
		return &apollo.ValidateTokenResp{
			Validated: false,
		}, nil
	}

	return &apollo.ValidateTokenResp{
		Validated: true,
	}, nil
}
