package redisUtil

import (
	"encoding/hex"
	"github.com/zeromicro/go-zero/core/jsonx"
)

type Redis struct {
	Key  string
	Data *RedisData
}
type RedisData struct {
	Id   string `json:"id"`
	Flag string `json:"flag"`
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
	return newRedis(continueKeyPrefix+hex.EncodeToString([]byte(redisKeyID)), newContinueRedisData(redisKeyID))
}
func NewBindRedis(userID string) *Redis {
	return newRedis(bindKeyPrefix+hex.EncodeToString([]byte(userID)), newBindRedisData(userID))
}
func newRedisData(id string, flag string) *RedisData {
	return &RedisData{
		Id:   id,
		Flag: flag,
	}
}

func newContinueRedisData(redisKeyID string) *RedisData {
	return newRedisData(redisKeyID, ContinueFlag)
}

func newBindRedisData(id string) *RedisData {
	return newRedisData(id, BindFlag)
}

func (r *RedisData) String() string {
	str, _ := jsonx.MarshalToString(r)
	return str
}
