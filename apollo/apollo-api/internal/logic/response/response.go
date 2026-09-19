package response

import (
	"context"
	"encoding/json"
	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strconv"
)

func OK() types.BaseResponse { return types.BaseResponse{Code: 200, Message: "success"} }
func Subject(ctx context.Context) (int64, error) {
	v, ok := ctx.Value("id").(json.Number)
	if !ok {
		return 0, application.ErrCredentials
	}
	id, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil || id <= 0 {
		return 0, application.ErrCredentials
	}
	return id, nil
}
func ID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, application.ErrInvalid
	}
	return id, nil
}
func Grant(g *apollo.SubsystemToken) types.SubsystemToken {
	return types.SubsystemToken{Id: g.Id, Value: g.Value, Name: g.Name, Date: types.Birthday{Year: g.Year, Month: g.Month, Day: g.Day}}
}
