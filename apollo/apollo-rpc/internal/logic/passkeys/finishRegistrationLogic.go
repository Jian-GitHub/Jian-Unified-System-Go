package passkeyslogic

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/errorx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	passkeyUtil "jian-unified-system/apollo/apollo-rpc/util"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"strconv"
	"time"

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
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.PasskeysFinishRegistrationReq) (*apollo.PasskeysFinishRegistrationResp, error) {
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

	displayName := in.Name
	if len(displayName) == 0 {
		displayName = strconv.FormatInt(in.UserId, 10)
	}

	webauthnUser := &types.WebauthnUser{ID: in.UserId}

	// 4. Check credential
	credential, err := l.svcCtx.WebAuthn.FinishRegistration(
		webauthnUser,
		session,
		req,
	)
	if err != nil {
		fmt.Printf("verify fail: %v \n", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}

	if in.Type {
		_, err = l.svcCtx.UserModel.Insert(
			l.ctx,
			&ap.User{
				Id:       in.UserId,
				Locate:   in.Locate,
				Language: in.Language,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	transport, err := json.Marshal(credential.Transport)
	if err != nil {
		return nil, errorx.Wrap(err, "marshal credential transport failed")
	}

	_, err = l.svcCtx.PasskeyModel.Insert(l.ctx, &ap.Passkey{
		CredentialId: base64.RawURLEncoding.EncodeToString(credential.ID),
		UserId:       webauthnUser.ID,
		DisplayName:  displayName,
		PublicKey:    base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		SignCount:    int64(credential.Authenticator.SignCount),
		Transports: sql.NullString{
			String: string(transport),
			Valid:  true,
		},
	})
	if err != nil {
		if in.Type {
			_ = l.svcCtx.UserModel.Delete(l.ctx, in.UserId)
		}
		fmt.Printf("save Credential fail: %v \n", err)
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}
	date := time.Now()
	return &apollo.PasskeysFinishRegistrationResp{
		PasskeysId:   base64.RawURLEncoding.EncodeToString(credential.ID),
		Locale:       in.Locate,
		Language:     in.Language,
		PasskeysName: displayName,
		Year:         int64(date.Year()),
		Month:        int64(date.Month()),
		Day:          int64(date.Day()),
	}, nil
}

func (l *FinishRegistrationLogic) saveCredential(uid int64, credential *webauthn.Credential) error {
	transport, err := json.Marshal(credential.Transport)
	if err != nil {
		return err
	}

	_, err = l.svcCtx.PasskeyModel.Insert(l.ctx, &ap.Passkey{
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
