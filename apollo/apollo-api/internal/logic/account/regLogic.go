// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/geo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegLogic {
	return &RegLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegLogic) Reg(req *types.RegReq, locale ...string) (resp *types.RegResp, err error) {
	resolved := geo.UnknownCountry
	if len(locale) > 0 && locale[0] != "" {
		resolved = locale[0]
	}
	token, err := l.svcCtx.Sessions.RegisterWithLocale(l.ctx, req.Email, req.Password, req.ConfirmPassword, req.Language, resolved)
	if err != nil {
		return nil, err
	}
	return &types.RegResp{BaseResponse: types.BaseResponse{Code: 200, Message: "success"}, Data: types.RegData{Token: token}}, nil
}
