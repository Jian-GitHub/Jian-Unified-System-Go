package accountlogic

import "jian-unified-system/apollo/apollo-rpc/internal/logic/response"

func transportError(err error) error { return response.Error(err) }
