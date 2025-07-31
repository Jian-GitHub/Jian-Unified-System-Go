package passkeys

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strconv"

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

func (l *LoginStartLogic) LoginStart() (resp *types.LoginStartResp, err error) {
	// todo: add your logic here and delete this line
	// 1. gRPC
	loginResp, err := l.svcCtx.ApolloPasskeys.StartLogin(l.ctx, &apollo.Empty{})
	if err != nil {
		l.Logger.Errorf("gRPC调用失败: err=%v", err)
		return nil, fmt.Errorf("登录初始化失败")
	}

	// 4. save session
	sessionID := l.svcCtx.Snowflake.Generate().Int64()
	sessionKey := "webauthn:login:" + hex.EncodeToString([]byte(strconv.FormatInt(sessionID, 10)))
	sessionDataJson, err := json.Marshal(loginResp.SessionData)
	if err != nil {
		l.Logger.Errorf("SessionData 转 JSON 失败: err=%v", err)
		return nil, fmt.Errorf("SessionData 转 JSON 失败")
	}
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, sessionKey, string(sessionDataJson), 300); err != nil {
		l.Logger.Errorf("Redis存储失败: key=%s, err=%v", sessionKey, err)
		return nil, fmt.Errorf("系统错误")
	}

	return &types.LoginStartResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		LoginStartRespData: struct {
			OptionsJson string `json:"options_json"`
			SessionID   string `json:"session_id"`
		}{
			OptionsJson: string(loginResp.OptionsJson),
			SessionID:   sessionKey,
		},
	}, nil
}
