// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package passkeys

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginStartLogic {
	return &LoginStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginStartLogic) LoginStart(req *types.Empty) (resp *types.CeremonyResp, err error) {
	r, err := l.svcCtx.Passkeys.StartLogin(l.ctx, &apollo.Empty{})
	if err != nil {
		return nil, err
	}
	return &types.CeremonyResp{BaseResponse: response.OK(), Data: types.CeremonyData{OptionsJson: string(r.OptionsJson), SessionID: string(r.SessionData)}}, nil
}
