// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package passkeys

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginFinishLogic {
	return &LoginFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginFinishLogic) LoginFinish(req *types.LoginFinishReq) (resp *types.LoginResp, err error) {
	r, err := l.svcCtx.Passkeys.FinishLogin(l.ctx, &apollo.PasskeysFinishLoginReq{CredentialJson: req.Assertion, SessionDataJson: req.SessionID})
	if err != nil {
		return nil, err
	}
	token, err := l.svcCtx.Sessions.Issue(l.ctx, r.UserId)
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{BaseResponse: response.OK(), Data: types.LoginData{Token: token, Id: strconv.FormatInt(r.UserId, 10), Name: types.UserName{GivenName: r.GivenName, MiddleName: r.MiddleName, FamilyName: r.FamilyName}, Avatar: r.Avatar, Locale: r.Locale, Language: r.Language, Birthday: types.Birthday{Year: r.BirthdayYear, Month: r.BirthdayMonth, Day: r.BirthdayDay}}}, nil
}
