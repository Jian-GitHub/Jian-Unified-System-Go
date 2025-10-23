package security

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenPasskeysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenPasskeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenPasskeysLogic {
	return &GetTenPasskeysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenPasskeysLogic) GetTenPasskeys(req *types.GetTenPasskeysReq) (resp *types.GetTenPasskeysResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GetTenPasskeysResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	passkeysResp, err := l.svcCtx.ApolloPasskeys.FindTenPasskeys(l.ctx, &apollo.FindTenPasskeysReq{
		UserId: id,
		Page:   req.Page,
	})
	if err != nil {
		return nil, err
	}

	passkeys := make([]types.Passkey, 0)
	if len(passkeysResp.Passkeys) != 0 {
		for _, passkey := range passkeysResp.Passkeys {
			passkeys = append(passkeys, types.Passkey{
				Id:   passkey.Id,
				Name: passkey.Name,
				Date: types.RespnseDate{
					Year:  passkey.Year,
					Month: passkey.Month,
					Day:   passkey.Day,
				},
				IsEnabled: passkey.IsEnabled,
			})
		}
	}

	return &types.GetTenPasskeysResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GetTenPasskeysData: struct {
			Passkeys []types.Passkey `json:"passkeys"`
		}{
			Passkeys: passkeys,
		},
	}, nil
}
