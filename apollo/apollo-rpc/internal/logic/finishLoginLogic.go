package logic

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"
	"strings"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishLoginLogic {
	return &FinishLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishLogin 登陆第二步 - 完成
func (l *FinishLoginLogic) FinishLogin(in *apollo.FinishLoginReq) (*apollo.FinishLoginResp, error) {
	// todo: add your logic here and delete this line
	fmt.Println(string(in.SessionData))
	fmt.Println("RPC 拿到 Received credential json:", string(in.CredentialJson))

	// 1. 先Base64解码
	s := strings.Trim(string(in.SessionData), "\"")
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "base64 decode session data failed: "+err.Error())
	}

	// 2. 反序列化SessionData（使用完整包路径）
	var session webauthn.SessionData
	if e := json.Unmarshal(decoded, &session); e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data: "+err.Error())
	}

	fmt.Println("2 session:", session)
	//fmt.Println("sessionData:", s)
	//// 1. 尝试直接解析 SessionData
	//var session webauthn.SessionData
	//if err := json.Unmarshal([]byte(s), &session); err == nil {
	//	// 解析成功，直接使用
	//	fmt.Println("直接解析 SessionData 成功")
	//} else {
	//	// 2. 如果直接解析失败，尝试去除可能的额外引号
	//	var sessionStr string
	//	if err := json.Unmarshal(in.SessionData, &sessionStr); err != nil {
	//		return nil, status.Error(codes.InvalidArgument, "无法解析 session 数据: "+err.Error())
	//	}
	//
	//	// 3. 去除可能的额外引号
	//	sessionStr = strings.Trim(sessionStr, `"`)
	//
	//	// 4. 将字符串解析为 SessionData
	//	if err := json.Unmarshal([]byte(sessionStr), &session); err != nil {
	//		return nil, status.Error(codes.InvalidArgument, "无效的 session 数据: "+err.Error())
	//	}
	//}
	//
	//fmt.Printf("原始Session: %+v\n", session) // 调试用
	//usera := &types.User{ID: 1945866326408978432}
	//session.UserID = usera.WebAuthnID()

	// 2. 创建模拟HTTP请求
	httpRequest, err := createCredentialRequest(in.CredentialJson)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	//var assertion struct {
	//	ID       string `json:"id"`
	//	Type     string `json:"type"`
	//	RawID    string `json:"rawId"`
	//	Response struct {
	//		ClientDataJSON    string `json:"clientDataJSON"`
	//		AuthenticatorData string `json:"authenticatorData"`
	//		Signature         string `json:"signature"`
	//		UserHandle        string `json:"userHandle"`
	//	} `json:"response"`
	//	ClientExtensionResults map[string]interface{} `json:"clientExtensionResults"`
	//}
	//
	//if err = json.Unmarshal(in.CredentialJson, &assertion); err != nil {
	//	return nil, fmt.Errorf("解析断言失败: %w", err)
	//}
	//
	//// 3. Base64URL解码各个字段
	//decoder := base64.RawURLEncoding
	//
	//// 解码RawID
	//rawID, err := decoder.DecodeString(assertion.RawID)
	//if err != nil {
	//	return nil, fmt.Errorf("RawID解码失败: %w", err)
	//}
	//
	//// 解码ClientDataJSON
	//clientDataJSON, err := decoder.DecodeString(assertion.Response.ClientDataJSON)
	//if err != nil {
	//	return nil, fmt.Errorf("ClientDataJSON解码失败: %w", err)
	//}
	//
	//// 解码AuthenticatorData
	//authenticatorData, err := decoder.DecodeString(assertion.Response.AuthenticatorData)
	//if err != nil {
	//	return nil, fmt.Errorf("AuthenticatorData解码失败: %w", err)
	//}
	//
	//// 解码Signature
	//signature, err := decoder.DecodeString(assertion.Response.Signature)
	//if err != nil {
	//	return nil, fmt.Errorf("Signature解码失败: %w", err)
	//}
	//
	//// 解码UserHandle
	//userHandle, err := decoder.DecodeString(assertion.Response.UserHandle)
	////userHandle, err = decoder.DecodeString(string(userHandle))
	//if err != nil {
	//	return nil, fmt.Errorf("UserHandle解码失败: %w", err)
	//}
	//
	//type credentialb struct {
	//	ID       string `json:"id"`
	//	Type     string `json:"type"`
	//	RawID    []byte `json:"rawId"`
	//	Response struct {
	//		ClientDataJSON    []byte `json:"clientDataJSON"`
	//		AuthenticatorData []byte `json:"authenticatorData"`
	//		Signature         []byte `json:"signature"`
	//		UserHandle        []byte `json:"userHandle"`
	//	} `json:"response"`
	//	ClientExtensionResults map[string]interface{} `json:"clientExtensionResults"`
	//}
	//
	//aaa := &credentialb{
	//	ID:    assertion.ID,
	//	Type:  assertion.Type,
	//	RawID: rawID,
	//	Response: struct {
	//		ClientDataJSON    []byte `json:"clientDataJSON"`
	//		AuthenticatorData []byte `json:"authenticatorData"`
	//		Signature         []byte `json:"signature"`
	//		UserHandle        []byte `json:"userHandle"`
	//	}{
	//		ClientDataJSON:    clientDataJSON,
	//		AuthenticatorData: authenticatorData,
	//		Signature:         signature,
	//		UserHandle:        userHandle,
	//	},
	//	ClientExtensionResults: map[string]interface{}{},
	//}

	// 4. 验证断言
	//credential, err := l.svcCtx.WebAuthn.FinishLogin(user, session, httpRequest)

	//user := &types.User{ID: int64(binary.BigEndian.Uint64(aaa.Response.UserHandle))}
	//fmt.Printf("User.WebAuthnID(): %x\n", user.WebAuthnID())
	//session.UserID = user.WebAuthnID()
	fmt.Printf("Session.UserID: %x\n", session.UserID)

	credential, err := l.svcCtx.WebAuthn.FinishDiscoverableLogin(
		l.discoverableUserHandler,
		//user,
		session,
		httpRequest,
	)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "assertion verification failed: "+err.Error())
	}

	// 5. 返回验证结果
	return &passkeys.FinishLoginResp{
		CredentialId: credential.ID,
		UserId:       123, // 从SessionData还原
		//UserId:       int64(binary.BigEndian.Uint64(user.WebAuthnID())), // 从SessionData还原
	}, nil
}

// 实现 DiscoverableUserHandler
func (l *FinishLoginLogic) discoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	// rawID - Passkeys ID
	// userHandle - User ID
	fmt.Println("rawID 进来是: ", string(rawID))
	fmt.Println("userHandle 进来是: ", string(userHandle))

	//userHandleData, err := base64.RawURLEncoding.DecodeString(string(userHandle))
	//if err != nil {
	//	return nil, fmt.Errorf("userHandleData转不成")
	//}
	// 1. 将二进制 userHandle 转为整数 user_id
	if len(userHandle) < 8 {
		return nil, fmt.Errorf("invalid userHandle length")
	}
	userID := int64(binary.BigEndian.Uint64(userHandle))
	fmt.Println("UserID:", userID)

	// 3. 验证凭证属于该用户
	credentialID := base64.RawURLEncoding.EncodeToString(rawID)
	fmt.Println("credentialID:", credentialID)
	passkey, err := l.svcCtx.PasskeyModel.FindOne(
		l.ctx,
		credentialID,
	)
	if err != nil {
		return nil, fmt.Errorf("credential not associated with user")
	}

	fmt.Println("Passkeys id:", passkey.CredentialId)
	fmt.Println("Passkeys user id:", passkey.UserId)

	// 解析 publicKey 和 credentialId
	credID, err := base64.RawURLEncoding.DecodeString(passkey.CredentialId)
	if err != nil {
		return nil, fmt.Errorf("invalid credentialId: %w", err)
	}

	pubKey, err := base64.RawURLEncoding.DecodeString(passkey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid publicKey: %w", err)
	}

	var trans []protocol.AuthenticatorTransport
	err = json.Unmarshal([]byte(passkey.Transports.String), &trans)
	if err != nil {
		return nil, err
	}

	// 构造 Credential
	credential := webauthn.Credential{
		ID:        credID,
		PublicKey: pubKey,
		Transport: trans,
		Authenticator: webauthn.Authenticator{
			SignCount: uint32(passkey.SignCount),
		},
		Flags: webauthn.CredentialFlags{
			BackupEligible: true,
			BackupState:    false,
		},
	}
	fmt.Println("credential:", credential)

	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(passkey.UserId))

	//t := base64.RawURLEncoding.EncodeToString(buf)

	// 构造 LoginUser
	user := &types.User{
		//ID: int64(binary.BigEndian.Uint64(userHandle)), // 用 user_id 作为 ID
		ID:          passkey.UserId, // 用 user_id 作为 ID
		Name:        fmt.Sprintf("user_${Date.now()}@test.com"),
		DisplayName: passkey.DisplayName,
		Credentials: []webauthn.Credential{credential},
	}

	fmt.Println("UserHandle", userHandle)
	fmt.Println("UserID", user.WebAuthnID())
	return user, nil
}
