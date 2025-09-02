package config

type EmailConfig struct {
	Host        string
	Port        int
	DisplayName string
	Username    string
	Password    string
	From        string // 默认发件人
}

func DefaultEmailConfig() *EmailConfig {
	return &EmailConfig{
		Host:        "mail.JianUnifiedSystem.com",
		Port:        465,
		DisplayName: "Hermes",
		Username:    "noreply@jianunifiedsystem.com",
		Password:    "KsW2MDpfrh",
		From:        "noreply@jianunifiedsystem.com",
	}
}
