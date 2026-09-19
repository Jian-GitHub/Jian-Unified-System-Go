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

type BindStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindStartLogic {
	return &BindStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindStartLogic) BindStart(req *types.BindStartReq) (resp *types.CeremonyResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Passkeys.StartRegistration(l.ctx, &apollo.PasskeysStartRegistrationReq{UserId: id, DisplayName: req.Name, Bind: true, Locale: l.svcCtx.Config.DefaultLocale, Language: l.svcCtx.Config.DefaultLanguage})
	if err != nil {
		return nil, err
	}
	return &types.CeremonyResp{BaseResponse: response.OK(), Data: types.CeremonyData{OptionsJson: string(r.OptionsJson), SessionID: string(r.SessionData)}}, nil
}
