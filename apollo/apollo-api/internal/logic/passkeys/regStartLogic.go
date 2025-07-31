package passkeys

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegStartLogic {
	return &RegStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegStartLogic) RegStart(req *types.RegStartReq) (resp *types.RegStartResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 生成用户ID（雪花算法）
	uuid := l.svcCtx.Snowflake.Generate()
	userID := uuid.Int64()

	// 2. 调用gRPC服务
	regResp, err := l.svcCtx.ApolloPasskeys.StartRegistration(l.ctx, &apollo.PasskeysStartRegistrationReq{
		UserId:      userID,
		UserName:    hex.EncodeToString(uuid.Bytes()),
		DisplayName: "Jian Unified System",
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC调用失败: %v", err)
	}

	// 3. 存储SessionData到Redis（加密）
	sessionKey := fmt.Sprintf("webauthn:reg:%s", hex.EncodeToString(util.Int64ToBytes(userID)))
	dataJson, err := json.Marshal(regResp.SessionData)
	if err != nil {
		return nil, fmt.Errorf("SessionData解析失败: %v", err)
	}
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, sessionKey, string(dataJson), 300); err != nil {
		return nil, fmt.Errorf("Redis存储失败: %v", err)
	}

	return &types.RegStartResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RegStartRespData: struct {
			OptionsJson string `json:"options_json"`
			SessionID   string `json:"session_id"`
		}{
			OptionsJson: string(regResp.OptionsJson),
			SessionID:   sessionKey,
		},
	}, nil
}
