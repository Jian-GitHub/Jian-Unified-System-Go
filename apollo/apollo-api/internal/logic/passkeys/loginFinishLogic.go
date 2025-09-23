package passkeys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
	"strconv"

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

func (l *LoginFinishLogic) LoginFinish(req *types.LoginFinishReq) (resp *types.LoginFinishResp, err error) {
	// todo: add your logic here and delete this line
	// 1. Check params
	if req.SessionID == "" || req.Assertion == "" {
		return nil, fmt.Errorf("params")
	}

	// 2. parse JSON
	var assertion struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		RawID    string `json:"rawId"`
		Response struct {
			ClientDataJSON    string `json:"clientDataJSON"`
			AuthenticatorData string `json:"authenticatorData"`
			Signature         string `json:"signature"`
			UserHandle        string `json:"userHandle"`
		} `json:"response"`
		ClientExtensionResults map[string]interface{} `json:"clientExtensionResults"`
	}
	if err := json.Unmarshal([]byte(req.Assertion), &assertion); err != nil {
		return nil, fmt.Errorf("parse json fail: %w", err)
	}

	// 2. SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, fmt.Errorf("会话已过期，请重新登录")
	}
	fmt.Println(sessionData)

	// 3. call gRPC
	response, err := l.svcCtx.ApolloPasskeys.FinishLogin(l.ctx, &apollo.PasskeysFinishLoginReq{
		SessionDataJson: sessionData,
		CredentialJson:  req.Assertion,
	})
	if err != nil {
		l.Logger.Errorf("登录验证失败: err=%v", err)
		return nil, fmt.Errorf("身份验证失败")
	}

	// 4. Generate JWT
	args := make(map[string]interface{})
	args["id"] = response.UserId
	token, err := util.GenToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, args)
	if err != nil {
		return &types.LoginFinishResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "fail",
			},
		}, errorx.Wrap(errors.New("generate token fail"), "reg error")
	}

	return &types.LoginFinishResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		LoginFinishRespData: struct {
			Token    string            `json:"token"`
			Id       string            `json:"id"`
			Name     types.UserName    `json:"name"`
			Avatar   string            `json:"avatar"`
			Locale   string            `json:"locale"`
			Language string            `json:"language"`
			Birthday types.RespnseDate `json:"birthday"`
		}{
			Token: token,
			Id:    strconv.FormatInt(response.UserId, 10),
			Name: types.UserName{
				GivenName:  response.GivenName,
				MiddleName: response.MiddleName,
				FamilyName: response.FamilyName,
			},
			Avatar:   response.Avatar,
			Locale:   response.Locale,
			Language: response.Language,
			Birthday: types.RespnseDate{
				Year:  response.BirthdayYear,
				Month: response.BirthdayMonth,
				Day:   response.BirthdayDay,
			},
		},
	}, nil
}
