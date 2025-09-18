package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/stringx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strings"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSubsystemTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSubsystemTokenLogic {
	return &GenerateSubsystemTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateSubsystemTokenLogic) GenerateSubsystemToken(req *types.GenerateSubsystemTokenReq) (resp *types.GenerateSubsystemTokenResp, err error) {
	if len(req.Scope) < 1 {
		return &types.GenerateSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Parameter name or Scope is empty",
			},
		}, errorx.Wrap(errors.New("params"), "caller err")
	}

	var name string
	//name, err := stringx.Substr(stringx.Remove([]string{uuid.New().String()}, "-")[0], 0, 5)
	if stringx.NotEmpty(req.Name) {
		name = req.Name
	} else {
		name, err = stringx.Substr(strings.ToUpper(uuid.NewString()), 0, 5)
		if err != nil {
			name = "New Token"
		}
	}

	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GenerateSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	scopeBytes, err := json.Marshal(req.Scope)
	if err != nil {
		fmt.Println("报错")
		return &types.GenerateSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "Scope Marshal err",
			},
		}, errorx.Wrap(errors.New("scope"), "caller err")
	}

	rpcResp, err := l.svcCtx.ApolloSecurity.GenerateSubsystemToken(l.ctx, &apollo.GenerateSubsystemTokenReq{
		UserId: id,
		Name:   name,
		Scope:  scopeBytes,
	})
	if err != nil {
		return &types.GenerateSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -4,
				Message: "jus err",
			},
		}, errorx.Wrap(err, "system err")
	}

	return &types.GenerateSubsystemTokenResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GenerateSubsystemTokenData: struct {
			Token types.SubsystemToken `json:"token"`
		}{
			Token: types.SubsystemToken{
				Id:    rpcResp.Token.Id,
				Value: rpcResp.Token.Value,
				Name:  rpcResp.Token.Name,
				Date: types.RespnseDate{
					Year:  rpcResp.Token.Year,
					Month: rpcResp.Token.Month,
					Day:   rpcResp.Token.Day,
				},
			},
		},
	}, nil
}
