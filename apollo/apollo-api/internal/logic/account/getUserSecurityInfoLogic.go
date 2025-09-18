package account

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

type GetUserSecurityInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSecurityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityInfoLogic {
	return &GetUserSecurityInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSecurityInfoLogic) GetUserSecurityInfo( /*req *types.GetUserSecurityInfoReq*/ ) (resp *types.GetUserSecurityInfoResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GetUserSecurityInfoResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	securityInfo, err := l.svcCtx.ApolloAccount.UserSecurityInfo(l.ctx, &apollo.UserSecurityInfoReq{
		UserId: id,
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]types.UserContact, 0)
	if len(securityInfo.Contacts) != 0 {
		for _, contact := range securityInfo.Contacts {
			contacts = append(contacts, types.UserContact{
				Id:          contact.Id,
				Value:       contact.Value,
				Type:        contact.Type,
				PhoneRegion: contact.PhoneRegion,
			})
		}
	}

	var year, month, day int64
	if securityInfo.PasswordUpdatedDate != nil {
		year = securityInfo.PasswordUpdatedDate.Year
		month = securityInfo.PasswordUpdatedDate.Month
		day = securityInfo.PasswordUpdatedDate.Day
	}

	return &types.GetUserSecurityInfoResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GetUserSecurityInfoData: struct {
			Contacts            []types.UserContact `json:"contacts"`
			PasswordUpdatedDate struct {
				Year  int64 `json:"year"`
				Month int64 `json:"month"`
				Day   int64 `json:"day"`
			} `json:"passwordUpdatedDate"`
			AccountSecurityTokenNum int64 `json:"accountSecurityTokenNum"`
			PasskeysNum             int64 `json:"passkeysNum"`
			ThirdPartyAccounts      struct {
				Github bool `json:"github"`
				Google bool `json:"google"`
			} `json:"thirdPartyAccounts"`
		}{
			Contacts: contacts,
			PasswordUpdatedDate: struct {
				Year  int64 `json:"year"`
				Month int64 `json:"month"`
				Day   int64 `json:"day"`
			}{
				Year:  year,
				Month: month,
				Day:   day,
			},
			AccountSecurityTokenNum: securityInfo.AccountSecurityTokenNum,
			PasskeysNum:             securityInfo.PasskeysNum,
			ThirdPartyAccounts: struct {
				Github bool `json:"github"`
				Google bool `json:"google"`
			}{
				Github: securityInfo.ThirdPartyAccounts.Github,
				Google: securityInfo.ThirdPartyAccounts.Google,
			},
		},
	}, nil
}
