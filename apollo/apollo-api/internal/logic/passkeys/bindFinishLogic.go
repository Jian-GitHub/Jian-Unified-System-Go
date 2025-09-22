package passkeys

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"
)

type BindFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindFinishLogic {
	return &BindFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindFinishLogic) BindFinish(req *types.BindFinishReq) (resp *types.BindFinishResp, err error) {
	// 1. Check params
	if len(req.SessionID) == 0 || len(req.Credential) == 0 {
		return nil, errorx.Wrap(errors.New("session_id or credential is null"), "params error")
	}

	// 2. Redis -> SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, errorx.Wrap(errors.New("not found"), "session error")
	}

	userID, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.BindFinishResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	// 3. Deserialize SessionData
	var sessionBytes []byte
	err = json.Unmarshal([]byte(sessionData), &sessionBytes)
	if err != nil {
		return nil, err
	}

	// 4. gRPC -> finish
	passkeysInfo, err := l.svcCtx.ApolloPasskeys.FinishRegistration(l.ctx, &apollo.PasskeysFinishRegistrationReq{
		UserId:         userID,
		SessionData:    sessionBytes,
		CredentialJson: []byte(req.Credential),
		Type:           false,
	})
	if err != nil {
		l.Logger.Errorf("gRPC fail: err=%v", err)
		return nil, errorx.Wrap(errors.New("verify fail: %v"), "reg error")
	}

	// 5. 清理会话数据
	if _, err := l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID); err != nil {
		l.Logger.Errorf("remove session fail: key=%s, err=%v", req.SessionID, err)
	}

	return &types.BindFinishResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		BindFinishRespData: struct {
			Id   string            `json:"id"`
			Name string            `json:"name"`
			Date types.RespnseDate `json:"date"`
		}{
			Id:   passkeysInfo.Id,
			Name: passkeysInfo.Name,
			Date: types.RespnseDate{
				Year:  passkeysInfo.Year,
				Month: passkeysInfo.Month,
				Day:   passkeysInfo.Day,
			},
		},
	}, nil
}
