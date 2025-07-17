package logic

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"
	"net/http"
	"net/http/httptest"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"encoding/json"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FinishRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishRegistrationLogic {
	return &FinishRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishRegistration 注册第二步 - 完成
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.FinishRegistrationReq) (*apollo.FinishRegistrationResp, error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if len(in.CredentialJson) == 0 || len(in.SessionData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	// 2. 反序列化SessionData
	var session webauthn.SessionData
	if err := json.Unmarshal(in.SessionData, &session); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data: "+err.Error())
	}

	//var regData struct{
	//	User *model.User `json:"user"`
	//	SessionData *webauthn.SessionData `json:"sessionData"`
	//}
	//if err := json.Unmarshal(in.CredentialJson, &regData); err != nil {
	//	return nil, fmt.Errorf("数据解析失败")
	//}

	// 3. 创建模拟HTTP请求
	req, err := createCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	user := &types.User{ID: in.UserId}

	fmt.Println("进行验证")
	reqBody, _ := io.ReadAll(req.Body)
	fmt.Println("最终发送到 FinishRegistration 的 req.Body:")
	fmt.Println(string(reqBody))
	req.Body = io.NopCloser(bytes.NewReader(reqBody)) // 读取后要重设

	fmt.Println(session.UserID)
	//var body struct {
	//	protocol.ParsedCredentialCreationData
	//}
	//err = json.Unmarshal(reqBody, &body)
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Println("body")
	//fmt.Println(body)
	//fmt.Println(body.ID)

	// 4. 验证凭证
	credential, err := l.svcCtx.WebAuthn.FinishRegistration(
		user,
		session,
		req,
	)
	if err != nil {
		fmt.Println("验证凭证报错: %v", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}

	err = l.saveCredential(user.ID, credential)
	if err != nil {
		fmt.Println("saveCredential报错: %v", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}
	// 1. 保存Passkeys记录（完全匹配你的结构体）
	//passkey := model.Passkeys{
	//	Handle: sql.NullString{String: string(user.WebAuthnID()), Valid: true}, // 注意：这里按你的要求转为string
	//	UserId: 12,                                                             //l.svcCtx.snow,
	//}
	//
	//println(passkey.Id)

	return &passkeys.FinishRegistrationResp{
		CredentialId: credential.ID,
		PublicKey:    credential.PublicKey,
	}, nil
}

// 辅助函数：将JSON转换为http.Request
func createCredentialRequest(jsonData []byte) (*http.Request, error) {
	//fmt.Println("辅助函数：将JSON转换为http.Request")
	//fmt.Println(string(jsonData))
	//credential, err := json.Marshal(jsonData)
	//if err != nil {
	//	return nil, err
	//}
	req := httptest.NewRequest("POST", "/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// 保存凭证
func (l *FinishRegistrationLogic) saveCredential(uid int64, credential *webauthn.Credential) error {
	transport, err := json.Marshal(credential.Transport)
	if err != nil {
		return err
	}

	_, err = l.svcCtx.PasskeyModel.Insert(l.ctx, &model.Passkey{
		CredentialId: base64.URLEncoding.EncodeToString(credential.ID),
		UserId:       uid,
		DisplayName:  "Jian Unified System",
		PublicKey:    base64.URLEncoding.EncodeToString(credential.PublicKey),
		SignCount:    int64(credential.Authenticator.SignCount),
		Transports: sql.NullString{
			String: string(transport),
			Valid:  true,
		},
	})
	return err
}
