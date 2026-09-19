// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package passkeys

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/geo"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegStartLogic {
	return &RegStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegStartLogic) RegStart(req *types.RegStartReq, locale ...string) (resp *types.CeremonyResp, err error) {
	resolved := geo.UnknownCountry
	if len(locale) > 0 && locale[0] != "" {
		resolved = locale[0]
	}
	r, err := l.svcCtx.Passkeys.StartRegistration(l.ctx, &apollo.PasskeysStartRegistrationReq{UserId: l.svcCtx.Sessions.NextID(), UserName: req.UserName, DisplayName: req.DisplayName, Locale: resolved, Language: l.svcCtx.Config.DefaultLanguage})
	if err != nil {
		return nil, err
	}
	return &types.CeremonyResp{BaseResponse: response.OK(), Data: types.CeremonyData{OptionsJson: string(r.OptionsJson), SessionID: string(r.SessionData)}}, nil
}
