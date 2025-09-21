package jwt

// TokenConfig JWT 认证需要的密钥和过期时间配置
type TokenConfig struct {
	AccessSecret string
	AccessExpire int64
}
