package auth

import (
	"fmt"
	"strings"
)

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

func TokenFetchForRegion(region Region, message string) Error {
	parts := []string{fmt.Sprintf("区域 %s 获取 Token 失败: %s", region.Value, message)}
	parts = append(parts, fmt.Sprintf("建议: 确认当前 Cookie 属于 %s 区域且未过期。", region.Value))
	parts = append(parts, fmt.Sprintf("可运行 `byte-cli auth login -r %s` 重新登录，或用 `byte-cli auth config show --show-secret` 检查复制到开发机的 Cookie。", region.Value))
	return Error{Message: strings.Join(parts, "\n")}
}

func InvalidResponse(message string) Error {
	return Error{Message: "无效的响应: " + message}
}
