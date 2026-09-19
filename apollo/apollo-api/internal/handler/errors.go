package handler

import (
	"context"
	"errors"
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConfigureResponses installs go-zero's error extension once at startup.
func ConfigureResponses() {
	httpx.SetErrorHandlerCtx(func(_ context.Context, err error) (int, any) {
		code, message := http.StatusInternalServerError, "service unavailable"
		switch {
		case errors.Is(err, application.ErrInvalid):
			code, message = http.StatusBadRequest, "invalid input"
		case errors.Is(err, application.ErrCredentials):
			code, message = http.StatusUnauthorized, "invalid credentials"
		case errors.Is(err, application.ErrChallenge):
			code, message = http.StatusUnauthorized, "challenge rejected"
		case errors.Is(err, application.ErrConflict):
			code, message = http.StatusConflict, "account already exists"
		}
		if code == http.StatusInternalServerError {
			switch status.Code(err) {
			case codes.InvalidArgument:
				code, message = 400, "invalid input"
			case codes.Unauthenticated:
				code, message = 401, "invalid credentials or session"
			case codes.PermissionDenied:
				code, message = 403, "access denied"
			case codes.NotFound:
				code, message = 404, "resource not found"
			case codes.AlreadyExists:
				code, message = 409, "resource already exists"
			case codes.FailedPrecondition:
				code, message = 409, "at least one login credential is required"
			case codes.Unavailable:
				code, message = 503, "service unavailable"
			case codes.DeadlineExceeded:
				code, message = 504, "service timeout"
			}
		}
		return code, types.BaseResponse{Code: code, Message: message}
	})
}
