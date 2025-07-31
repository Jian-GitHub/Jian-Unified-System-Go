package passkeyslogic

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"net/http"
	"net/http/httptest"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.PasskeysFinishRegistrationReq) (*apollo.Empty, error) {
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

	// 3. 创建模拟HTTP请求
	req, err := createCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	user := &types.User{ID: in.UserId}

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
	return &apollo.Empty{}, nil
}

// 辅助函数：将JSON转换为http.Request
func createCredentialRequest(jsonData []byte) (*http.Request, error) {
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
		CredentialId: base64.RawURLEncoding.EncodeToString(credential.ID),
		UserId:       uid,
		DisplayName:  "Jian Unified System",
		PublicKey:    base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		SignCount:    int64(credential.Authenticator.SignCount),
		Transports: sql.NullString{
			String: string(transport),
			Valid:  true,
		},
	})
	return err
}
