// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sso

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"google.golang.org/grpc/metadata"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleSheetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogleSheetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleSheetsLogic {
	return &GoogleSheetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GoogleSheetsLogic) GoogleSheets(req *types.GoogleSheetsReq) (resp *types.GoogleSheetsResp, err error) {
	ctx := metadata.AppendToOutgoingContext(l.ctx, "x-apollo-sso-service", l.svcCtx.Config.SSO.RPCSecret)
	v, e := l.svcCtx.SSO.GoogleSheets(ctx, &apollo.GoogleSheetsReq{ClientId: req.ClientId, ClientSecret: req.ClientSecret, Token: req.Token, Action: req.Action, State: req.State, Code: req.Code, Spreadsheet: req.Spreadsheet, Range: req.Range})
	if e != nil {
		return nil, e
	}
	out := &types.GoogleSheetsResp{Url: v.Url, Bound: v.Bound, Connected: v.Connected, Rows: [][]string{}}
	for _, r := range v.Rows {
		out.Rows = append(out.Rows, r.Cells)
	}
	return out, nil
}
