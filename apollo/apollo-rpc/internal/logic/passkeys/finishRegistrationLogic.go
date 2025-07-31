package passkeyslogic

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	passkeyUtil "jian-unified-system/apollo/apollo-rpc/util"

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

// FinishRegistration Passkeys Registration - Step 2
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.PasskeysFinishRegistrationReq) (*apollo.Empty, error) {
	// 1. Check params
	if len(in.CredentialJson) == 0 || len(in.SessionData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}
	// 2. Deserialize SessionData
	var session webauthn.SessionData
	if err := json.Unmarshal(in.SessionData, &session); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data: "+err.Error())
	}

	// 3. create HTTP request
	req, err := passkeyUtil.CreateCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	webauthnUser := &types.WebauthnUser{ID: in.UserId}

	// 4. Check credential
	credential, err := l.svcCtx.WebAuthn.FinishRegistration(
		webauthnUser,
		session,
		req,
	)
	if err != nil {
		fmt.Println("verify fail: %v", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}

	_, err = l.svcCtx.UserModel.Insert(
		l.ctx,
		&model.User{
			Id:       in.UserId,
			Locate:   in.Locate,
			Language: in.Language,
		},
	)
	if err != nil {
		return nil, err
	}

	err = l.saveCredential(webauthnUser.ID, credential)
	if err != nil {
		_ = l.svcCtx.UserModel.Delete(l.ctx, in.UserId)
		fmt.Println("save Credential fail: %v", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}
	return &apollo.Empty{}, nil
}

func (l *FinishRegistrationLogic) saveCredential(uid int64, credential *webauthn.Credential) error {
	transport, err := json.Marshal(credential.Transport)
	if err != nil {
		return err
	}

	_, err = l.svcCtx.PasskeyModel.Insert(l.ctx, &model.Passkey{
		CredentialId: base64.RawURLEncoding.EncodeToString(credential.ID),
		UserId:       uid,
		DisplayName:  "Apollo System",
		PublicKey:    base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		SignCount:    int64(credential.Authenticator.SignCount),
		Transports: sql.NullString{
			String: string(transport),
			Valid:  true,
		},
	})
	return err
}
