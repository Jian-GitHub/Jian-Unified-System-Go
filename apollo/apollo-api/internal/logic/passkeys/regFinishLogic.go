package passkeys

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
	"log"
	"strings"

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

func (l *RegFinishLogic) RegFinish(req *types.RegFinishReq) (resp *types.RegFinishResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if len(req.SessionID) == 0 || len(req.Credential) == 0 {
		return nil, errorx.Wrap(errors.New("session_id or credential is null"), "params error")
	}

	// 2. 从Redis获取SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, errorx.Wrap(errors.New("not found"), "session error")
	}

	// 从 sessionKey 中解析出 userID
	parts := strings.Split(req.SessionID, ":")
	if len(parts) != 3 {
		log.Fatalf("Invalid session key format")
	}
	hexStr := parts[2]
	userIDBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatalf("Failed to decode hex string: %v", err)
	}
	userID := util.BytesToInt64(userIDBytes)

	// 3. 反序列化 SessionData（它本质上是一个 JSON 字符串，编码的是 []byte）
	var sessionBytes []byte
	err = json.Unmarshal([]byte(sessionData), &sessionBytes)
	if err != nil {
		return nil, err
	}

	// 4. 调用gRPC服务完成注册
	_, err = l.svcCtx.ApolloPasskeys.FinishRegistration(l.ctx, &apollo.PasskeysFinishRegistrationReq{
		UserId:         userID,
		SessionData:    sessionBytes,
		CredentialJson: []byte(req.Credential),
	})
	if err != nil {
		l.Logger.Errorf("gRPC调用失败: err=%v", err)
		return nil, errorx.Wrap(errors.New("verify fail: %v"), "reg error")
	}

	// 5. 清理会话数据
	if _, err := l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID); err != nil {
		l.Logger.Errorf("删除会话失败: key=%s, err=%v", req.SessionID, err)
		// 此处不返回错误，因为主流程已成功
	}

	// 6. All done -> Generate JWT
	args := make(map[string]interface{})
	args["test"] = true
	args["id"] = userID
	token, err := util.GenToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, args)
	if err != nil {
		return &types.RegFinishResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "fail",
			},
		}, errorx.Wrap(errors.New("generate token fail"), "reg error")
	}

	return &types.RegFinishResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RegFinishRespData: struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	}, nil
}
