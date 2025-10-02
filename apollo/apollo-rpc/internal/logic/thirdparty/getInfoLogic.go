package thirdpartylogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetInfo 获取第三方账号绑定信息
func (l *GetInfoLogic) GetInfo(in *apollo.ThirdPartyGetInfoReq) (*apollo.ThirdPartyGetInfoResp, error) {
	accounts, err := l.svcCtx.ThirdPartyModel.FindBatch(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	resp := make([]*apollo.ThirdPartyAccountInfo, 0)
	if len(*accounts) > 0 {
		for _, v := range *accounts {
			resp = append(resp, &apollo.ThirdPartyAccountInfo{
				Id:       v.Id,
				Provider: v.Provider,
				Content:  v.Name,
			})
		}
	}
	return &apollo.ThirdPartyGetInfoResp{
		Accounts: resp,
	}, nil
}
