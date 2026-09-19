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

type RegFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegFinishLogic {
	return &RegFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegFinishLogic) RegFinish(req *types.RegFinishReq, locale ...string) (resp *types.RegResp, err error) {
	resolved := geo.UnknownCountry
	if len(locale) > 0 && locale[0] != "" {
		resolved = locale[0]
	}
	r, err := l.svcCtx.Passkeys.FinishRegistration(l.ctx, &apollo.PasskeysFinishRegistrationReq{CredentialJson: []byte(req.Credential), SessionData: []byte(req.SessionID), Type: true, Locate: resolved, Language: req.Language})
	if err != nil {
		return nil, err
	}
	token, err := l.svcCtx.Sessions.Issue(l.ctx, r.UserId)
	if err != nil {
		return nil, err
	}
	return &types.RegResp{BaseResponse: response.OK(), Data: types.RegData{Token: token}}, nil
}
