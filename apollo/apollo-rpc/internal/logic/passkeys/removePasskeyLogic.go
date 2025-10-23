package passkeyslogic

import (
	"context"
	ap "jian-unified-system/jus-core/data/mysql/apollo"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePasskeyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemovePasskeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePasskeyLogic {
	return &RemovePasskeyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RemovePasskey 移除 Passkey
func (l *RemovePasskeyLogic) RemovePasskey(in *apollo.RemovePasskeyReq) (*apollo.RemovePasskeyResp, error) {
	// todo: add your logic here and delete this line
	err := l.svcCtx.PasskeyModel.DeleteOrRestorePasskey(
		l.ctx,
		&ap.Passkey{
			CredentialId: in.PasskeyId,
			UserId:       in.UserId,
			IsDeleted:    1,
		})
	if err != nil {
		return &apollo.RemovePasskeyResp{
			Success: false,
		}, err
	}

	return &apollo.RemovePasskeyResp{
		Success: true,
	}, nil
}
