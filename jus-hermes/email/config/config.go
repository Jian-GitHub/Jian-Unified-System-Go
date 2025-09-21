package config

type EmailConfig struct {
	Host        string
	Port        int
	DisplayName string
	Username    string
	Password    string
	From        string
}

func DefaultEmailConfig() *EmailConfig {
	return &EmailConfig{
		Host:        "smtp.example.com",
		Port:        465,
		DisplayName: "EMAIL_DISPLAY_NAME",
		Username:    "example@example.com",
		Password:    "PASSWORD",
		From:        "example@example.com",
	}
}
