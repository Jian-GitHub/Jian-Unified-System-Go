package passkeys

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindStartLogic {
	return &BindStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindStartLogic) BindStart(req *types.BindStartReq) (resp *types.BindStartResp, err error) {
	// 1. user UUID
	userID, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.BindStartResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	var name string
	if len(req.Name) == 0 {
		name = strconv.FormatInt(userID, 10)
	} else {
		name = req.Name
	}

	// 2. Call gRPC service
	regResp, err := l.svcCtx.ApolloPasskeys.StartRegistration(l.ctx, &apollo.PasskeysStartRegistrationReq{
		UserId:      userID,
		UserName:    name,
		DisplayName: name,
	})
	if err != nil {
		return &types.BindStartResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "gRPC err",
			},
		}, fmt.Errorf("gRPC调用失败: %v", err)
	}

	// 3. save SessionData -> Redis (base64)
	sessionKey := fmt.Sprintf("webauthn:bind:%s", hex.EncodeToString(util.Int64ToBytes(userID)))
	dataJson, err := json.Marshal(regResp.SessionData)
	if err != nil {
		return &types.BindStartResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: fmt.Sprintf("SessionData解析失败: %v", err),
			},
		}, fmt.Errorf("SessionData解析失败: %v", err)
	}
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, sessionKey, string(dataJson), 300); err != nil {
		return &types.BindStartResp{
			BaseResponse: types.BaseResponse{
				Code:    -4,
				Message: fmt.Sprintf("redis存储失败: %v", err),
			},
		}, fmt.Errorf("redis存储失败: %v", err)
	}

	return &types.BindStartResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		BindStartRespData: struct {
			OptionsJson string `json:"options_json"`
			SessionID   string `json:"session_id"`
		}{
			OptionsJson: string(regResp.OptionsJson),
			SessionID:   sessionKey,
		},
	}, nil
}
