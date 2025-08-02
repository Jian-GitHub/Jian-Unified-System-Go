package redisUtil

import "github.com/zeromicro/go-zero/core/jsonx"

type Redis struct {
	Key  string
	Data *RedisData
}
type RedisData struct {
	Id   *string `json:"id"`
	Flag string  `json:"flag"`
}

const (
	ContinueFlag      = "continue"
	BindFlag          = "bind"
	continueKeyPrefix = "apollo:thirdParty:continue:"
	bindKeyPrefix     = "apollo:thirdParty:bind:"
)

func newRedis(key string, data *RedisData) *Redis {
	return &Redis{
		Key:  key,
		Data: data,
	}
}

func NewContinueRedis(redisKeyID string) *Redis {
	return newRedis(continueKeyPrefix+redisKeyID, newContinueRedisData())
}
func NewBindRedis(userID *string) *Redis {
	return newRedis(bindKeyPrefix+*userID, newBindRedisData(userID))
}
func newRedisData(id *string, flag string) *RedisData {
	return &RedisData{
		Id:   id,
		Flag: flag,
	}
}

func newContinueRedisData() *RedisData {
	return newRedisData(nil, ContinueFlag)
}

func newBindRedisData(id *string) *RedisData {
	return newRedisData(id, BindFlag)
}

func (r *RedisData) String() string {
	str, _ := jsonx.MarshalToString(r)
	return str
}
