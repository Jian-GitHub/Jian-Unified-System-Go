package util

import (
	"crypto/sha512"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"unicode"
)

func IsStrongPassword(password string) bool {
	// 检查最小长度
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// HashPasswordBcrypt 使用bcrypt (自动加盐)
func HashPasswordBcrypt(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPasswordBcrypt 验证密码和哈希是否匹配
func VerifyPasswordBcrypt(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashSHA512 哈希512 + 盐 -> 加密字符串
func HashSHA512(password, salt string) string {
	hasher := sha512.New()
	hasher.Write([]byte(salt))
	hasher.Write([]byte(password))
	hash := hasher.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}
