package account

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

type GenerateTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateTokenLogic) GenerateToken(req *types.GenerateTokenReq) (resp *types.GenerateTokenResp, err error) {
	if len(req.Scope) < 1 {
		return &types.GenerateTokenResp{
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
		return &types.GenerateTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	scopeBytes, err := json.Marshal(req.Scope)
	if err != nil {
		fmt.Println("报错")
		return &types.GenerateTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "Scope Marshal err",
			},
		}, errorx.Wrap(errors.New("scope"), "caller err")
	}

	rpcResp, err := l.svcCtx.ApolloAccount.GenerateToken(l.ctx, &apollo.GenerateTokenReq{
		UserId: id,
		Name:   name,
		Scope:  scopeBytes,
	})
	if err != nil {
		return &types.GenerateTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -4,
				Message: "jus err",
			},
		}, errorx.Wrap(err, "system err")
	}

	return &types.GenerateTokenResp{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		GenerateTokenData: struct {
			Token string `json:"token"`
		}{
			rpcResp.Token,
		},
	}, nil
}
