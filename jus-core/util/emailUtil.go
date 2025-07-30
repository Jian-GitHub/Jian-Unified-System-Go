package util

import "regexp"

// 判断是否为电子邮件地址
func IsEmail(s string) bool {
	// 简单但实用的邮箱正则
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(s)
}
