package passkeys

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
	"jian-unified-system/apollo/apollo-rpc/passkeys"

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

func (l *LoginFinishLogic) LoginFinish(req *types.LoginFinishReq) (resp *types.LoginFinishResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if req.SessionID == "" || req.Assertion == "" {
		return nil, fmt.Errorf("参数不完整")
	}

	fmt.Println("API拿到:", req.Assertion)
	// 2. 解析前端传来的JSON断言
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
		return nil, fmt.Errorf("解析断言失败: %w", err)
	}

	// 3. Base64URL解码各个字段
	decoder := base64.RawURLEncoding

	// 解码RawID
	rawID, err := decoder.DecodeString(assertion.RawID)
	if err != nil {
		return nil, fmt.Errorf("RawID解码失败: %w", err)
	}

	// 解码ClientDataJSON
	clientDataJSON, err := decoder.DecodeString(assertion.Response.ClientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("ClientDataJSON解码失败: %w", err)
	}

	// 解码AuthenticatorData
	authenticatorData, err := decoder.DecodeString(assertion.Response.AuthenticatorData)
	if err != nil {
		return nil, fmt.Errorf("AuthenticatorData解码失败: %w", err)
	}

	// 解码Signature
	signature, err := decoder.DecodeString(assertion.Response.Signature)
	if err != nil {
		return nil, fmt.Errorf("Signature解码失败: %w", err)
	}

	// 解码UserHandle
	userHandle, err := decoder.DecodeString(assertion.Response.UserHandle)
	//userHandle, err = decoder.DecodeString(string(userHandle))
	if err != nil {
		return nil, fmt.Errorf("UserHandle解码失败: %w", err)
	}

	//4. 构建WebAuthn凭证对象
	type credential struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		RawID    []byte `json:"rawId"`
		Response struct {
			ClientDataJSON    []byte `json:"clientDataJSON"`
			AuthenticatorData []byte `json:"authenticatorData"`
			Signature         []byte `json:"signature"`
			UserHandle        []byte `json:"userHandle"`
		} `json:"response"`
		ClientExtensionResults map[string]interface{} `json:"clientExtensionResults"`
	}

	aaa := &credential{
		ID:    assertion.ID,
		Type:  assertion.Type,
		RawID: rawID,
		Response: struct {
			ClientDataJSON    []byte `json:"clientDataJSON"`
			AuthenticatorData []byte `json:"authenticatorData"`
			Signature         []byte `json:"signature"`
			UserHandle        []byte `json:"userHandle"`
		}{
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authenticatorData,
			Signature:         signature,
			UserHandle:        userHandle,
		},
		ClientExtensionResults: map[string]interface{}{},
	}

	fmt.Println("两次解码解析出来")
	fmt.Println(aaa)
	fmt.Println("string(aaa.Response.UserHandle)")
	fmt.Println(string(aaa.Response.UserHandle))
	//decodeString, err := base64.RawURLEncoding.DecodeString(string(aaa.Response.UserHandle))
	//if err != nil {
	//	fmt.Println(err)
	//	return nil, err
	//}
	fmt.Println(int64(binary.BigEndian.Uint64(aaa.Response.UserHandle)))
	//aaa.Response.UserHandle = decodeString

	// 2. 获取SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, fmt.Errorf("会话已过期，请重新登录")
	}

	fmt.Println("req.Assertion", req.Assertion)

	param, err := json.Marshal(aaa)
	fmt.Println("param", string(param))
	if err != nil {
		l.Logger.Errorf("转parameters,失败, err=%v", err)
	}
	// 3. 调用gRPC验证
	_, err = l.svcCtx.ApolloRpc.FinishLogin(l.ctx, &passkeys.FinishLoginReq{
		SessionData:    []byte(sessionData),
		CredentialJson: []byte(req.Assertion),
		//CredentialJson: param,
	})
	if err != nil {
		l.Logger.Errorf("登录验证失败: err=%v", err)
		return nil, fmt.Errorf("身份验证失败")
	}

	// 4. 生成业务Token
	//token, err := l.svcCtx.Auth.GenerateToken(finishResp.UserId)
	//if err != nil {
	//	l.Logger.Errorf("Token生成失败: user_id=%x, err=%v", finishResp.UserId, err)
	//	return nil, fmt.Errorf("系统错误")
	//}
	token := "token here"
	marshalToString, err := jsonx.MarshalToString(token)
	if err != nil {
		return nil, err
	}
	// 5. 清理会话
	_, _ = l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID)

	return &types.LoginFinishResp{
		BaseResponse: types.BaseResponse{Code: 200, Message: "success"},
		Token:        marshalToString,
	}, nil
}
