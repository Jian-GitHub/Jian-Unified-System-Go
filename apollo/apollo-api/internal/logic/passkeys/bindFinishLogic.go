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

type BindFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindFinishLogic {
	return &BindFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindFinishLogic) BindFinish(req *types.BindFinishReq) (resp *types.BindFinishResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Passkeys.FinishRegistration(l.ctx, &apollo.PasskeysFinishRegistrationReq{UserId: id, CredentialJson: []byte(req.Credential), SessionData: []byte(req.SessionID), Type: false, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &types.BindFinishResp{BaseResponse: response.OK(), Data: types.PasskeyData{Id: r.PasskeysId, Name: r.PasskeysName, Date: types.Birthday{Year: r.Year, Month: r.Month, Day: r.Day}}}, nil
}
