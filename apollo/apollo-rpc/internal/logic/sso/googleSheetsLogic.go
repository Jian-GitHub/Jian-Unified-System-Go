package ssologic

import (
	"context"
	"errors"
	gs "jian-unified-system/apollo/apollo-rpc/internal/application/googlesheets"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleSheetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogleSheetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleSheetsLogic {
	return &GoogleSheetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Authenticated server-to-server Sheets broker; Google credentials never leave Apollo.
func (l *GoogleSheetsLogic) GoogleSheets(in *apollo.GoogleSheetsReq) (*apollo.GoogleSheetsResp, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	session, e := l.svcCtx.SSO.Introspect(l.ctx, in.ClientId, in.ClientSecret, in.Token)
	if e != nil {
		return nil, rpcError(e)
	}
	if session.Scope&4 != 4 {
		return nil, status.Error(codes.PermissionDenied, "subsystem access denied")
	}
	owner, e := strconv.ParseInt(session.Subject, 10, 64)
	if e != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid subject")
	}
	v, e := l.svcCtx.GoogleSheets.Execute(l.ctx, owner, in.Token, in.Action, in.State, in.Code, in.Spreadsheet, in.Range)
	if e != nil {
		switch {
		case errors.Is(e, gs.ErrAuthorization):
			return nil, status.Error(codes.FailedPrecondition, "Google Sheets authorization required")
		case errors.Is(e, gs.ErrInvalid):
			return nil, status.Error(codes.InvalidArgument, "invalid spreadsheet or range")
		case errors.Is(e, gs.ErrAccess):
			return nil, status.Error(codes.NotFound, "Google spreadsheet inaccessible")
		default:
			return nil, status.Error(codes.Unavailable, "Google Sheets unavailable")
		}
	}
	out := &apollo.GoogleSheetsResp{Url: v.URL, Bound: v.Bound, Connected: v.Connected}
	for _, r := range v.Rows {
		out.Rows = append(out.Rows, &apollo.GoogleSheetRow{Cells: r})
	}
	return out, nil
}
