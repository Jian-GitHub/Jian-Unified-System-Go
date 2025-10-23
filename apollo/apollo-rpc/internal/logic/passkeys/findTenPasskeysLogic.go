package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	ap "jian-unified-system/jus-core/data/mysql/apollo"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindTenPasskeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindTenPasskeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTenPasskeysLogic {
	return &FindTenPasskeysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FindTenPasskeys 查询 10 个Passkeys
func (l *FindTenPasskeysLogic) FindTenPasskeys(in *apollo.FindTenPasskeysReq) (*apollo.FindTenPasskeysResp, error) {
	var passkeys *[]ap.Passkey
	passkeys, err := l.svcCtx.PasskeyModel.FindBatch(l.ctx, in.UserId, in.Page)
	if err != nil {
		return nil, err
	}

	// 整理 Passkeys
	var apolloPasskeys []*apollo.Passkey
	if passkeys != nil {
		apolloPasskeys = make([]*apollo.Passkey, 0, len(*passkeys))
		for _, passkey := range *passkeys {
			apolloPasskeys = append(apolloPasskeys, &apollo.Passkey{
				Id:        passkey.CredentialId,
				Name:      passkey.DisplayName,
				Year:      int64(passkey.CreatedAt.Year()),
				Month:     int64(passkey.CreatedAt.Month()),
				Day:       int64(passkey.CreatedAt.Day()),
				IsEnabled: passkey.IsEnabled == 1,
			})
		}
	} else {
		apolloPasskeys = make([]*apollo.Passkey, 0)
	}

	return &apollo.FindTenPasskeysResp{
		Passkeys: apolloPasskeys,
	}, nil
}
