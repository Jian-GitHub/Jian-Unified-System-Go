package passkeyslogic

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartLoginLogic {
	return &StartLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartLoginLogic) StartLogin() (*apollo.PasskeysStartLoginResp, error) {
	// todo: add your logic here and delete this line
	options, session, err := l.svcCtx.WebAuthn.BeginDiscoverableLogin(
	//webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		l.Logger.Errorf("BeginLogin error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 5. Deserialize response data
	optionsJson, err := json.Marshal(options)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal options")
	}

	sessionData, err := json.Marshal(session)
	fmt.Println("1 sessionData:", string(sessionData))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal session")
	}

	return &apollo.PasskeysStartLoginResp{
		OptionsJson: optionsJson,
		SessionData: sessionData,
	}, nil
}
