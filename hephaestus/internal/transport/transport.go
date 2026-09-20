package transport

import (
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reflect"
	"strings"
)

func Convert[T any](v any) (*T, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var out T
	e = json.Unmarshal(b, &out)
	if e == nil {
		emptyCollections(reflect.ValueOf(&out).Elem())
	}
	return &out, e
}

// Protobuf omits empty repeated fields on the wire. Restore JSON arrays at
// the conversion boundary while preserving nullable scalars such as wages.
func emptyCollections(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			emptyCollections(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				emptyCollections(v.Field(i))
			}
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return
		}
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		for i := 0; i < v.Len(); i++ {
			emptyCollections(v.Index(i))
		}
	}
}
func RPCError(e error) error {
	if e == nil {
		return nil
	}
	var business interface{ BusinessCode() string }
	if errors.As(e, &business) {
		c := codes.InvalidArgument
		switch business.BusinessCode() {
		case "NOT_FOUND":
			c = codes.NotFound
		case "UNAUTHENTICATED":
			c = codes.Unauthenticated
		case "FORBIDDEN":
			c = codes.PermissionDenied
		case "VERSION_CONFLICT", "IDEMPOTENCY_CONFLICT", "CONFLICT":
			c = codes.Aborted
		case "PRECONDITION_REQUIRED":
			c = codes.FailedPrecondition
		case "TOO_LARGE":
			c = codes.ResourceExhausted
		}
		return status.Error(c, business.BusinessCode()+": "+e.Error())
	}
	if errors.Is(e, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "service timeout")
	}
	if errors.Is(e, context.Canceled) {
		return status.Error(codes.Canceled, "request cancelled")
	}
	return status.Error(codes.Unavailable, "income service unavailable")
}
func Result[T any](v any, e error) (*T, error) {
	if e != nil {
		return nil, RPCError(e)
	}
	out, e := Convert[T](v)
	return out, RPCError(e)
}
func HTTPError(w http.ResponseWriter, e error) {
	c := http.StatusInternalServerError
	code := "INTERNAL"
	message := "income service unavailable"
	s, ok := status.FromError(e)
	if ok {
		message = s.Message()
		switch s.Code() {
		case codes.InvalidArgument:
			c = 422
			code = "INVALID_ARGUMENT"
		case codes.NotFound:
			c = 404
			code = "NOT_FOUND"
		case codes.Unauthenticated:
			c = 401
			code = "UNAUTHENTICATED"
		case codes.PermissionDenied:
			c = 403
			code = "FORBIDDEN"
		case codes.Aborted:
			c = 409
			code = "CONFLICT"
		case codes.FailedPrecondition:
			c = 428
			code = "PRECONDITION_REQUIRED"
		case codes.ResourceExhausted:
			c = 413
			code = "TOO_LARGE"
		case codes.Unavailable, codes.DeadlineExceeded:
			c = 503
			code = "UNAVAILABLE"
		}
	}
	if prefix, rest, exists := strings.Cut(message, ": "); exists {
		switch prefix {
		case "INVALID_ARGUMENT", "NOT_FOUND", "UNAUTHENTICATED", "FORBIDDEN", "VERSION_CONFLICT", "IDEMPOTENCY_CONFLICT", "CONFLICT", "PRECONDITION_REQUIRED", "TOO_LARGE":
			code = prefix
			message = rest
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": message, "request_id": w.Header().Get("X-Request-ID"), "field_errors": []any{}})
}
