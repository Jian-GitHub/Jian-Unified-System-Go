package passkeys

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
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

func (l *LoginStartLogic) LoginStart(req *types.LoginStartReq) (resp *types.LoginStartResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	//if req.UserID == "" {
	//	return nil, fmt.Errorf("user_id不能为空")
	//}

	// 2. 查询用户凭证（示例）
	//userID, err := strconv.ParseInt(req.UserID, 10, 64)
	//if err != nil {
	//	l.Logger.Errorf("userID 转换失败: err=%v", err)
	//	return nil, err
	//}
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.UserID)
	if err != nil {
		l.Logger.Errorf("查询用户凭证失败: user_id=%s, err=%v", req.UserID, err)
		return nil, fmt.Errorf("用户凭证获取失败")
	}

	// 3. 调用gRPC服务
	loginResp, err := l.svcCtx.PasskeysRpc.StartLogin(l.ctx, &apollo.StartLoginReq{
		UserId:          user.Id,
		CredentialsJson: []byte(req.Credentials),
	})
	if err != nil {
		l.Logger.Errorf("gRPC调用失败: err=%v", err)
		return nil, fmt.Errorf("登录初始化失败")
	}

	// 4. 存储会话数据
	sessionKey := "webauthn:login:" + hex.EncodeToString([]byte(strconv.FormatInt(user.Id, 10)))
	SessionDataJson, err := jsonx.MarshalToString(loginResp.SessionData)
	if err != nil {
		l.Logger.Errorf("SessionData 转 JSON 失败: err=%v", err)
		return nil, fmt.Errorf("SessionData 转 JSON 失败")
	}
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, sessionKey, SessionDataJson, 300); err != nil {
		l.Logger.Errorf("Redis存储失败: key=%s, err=%v", sessionKey, err)
		return nil, fmt.Errorf("系统错误")
	}

	// 5. 返回响应
	return &types.LoginStartResp{
		BaseResponse: types.BaseResponse{Code: 200, Message: "success"},
		Data: types.Data{
			OptionsJson: string(loginResp.OptionsJson),
			SessionID:   sessionKey,
		},
	}, nil
}
