package auth

import "fmt"

type Error struct {
	Message string
}

func (e Error) Error() string { return e.Message }

func IsAuthError(err error) bool {
	_, ok := err.(Error)
	return ok
}

func CookieNotFound(region string) Error {
	return Error{Message: fmt.Sprintf("区域 %s 未配置 Cookie，请先运行 byte-cli auth config set-cookie", region)}
}

func TokenFetch(message string) Error {
	return Error{Message: "获取 Token 失败: " + message}
}

func InvalidResponse(message string) Error {
	return Error{Message: "无效的响应: " + message}
}
