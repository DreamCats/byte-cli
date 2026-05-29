package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DreamCats/byte-cli/internal/config"
	"github.com/DreamCats/byte-cli/internal/httpclient"
)

var normalHeaders = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Language": "en-US,en;q=0.9",
	"Accept-Encoding": "gzip, deflate, br",
}

var codebaseHeaders = map[string]string{
	"Accept":          "application/json, text/plain, */*",
	"Accept-Language": "zh",
	"Domain":          "api-server",
}

type Manager struct {
	Region Region
}

func NewManager(region Region) Manager {
	return Manager{Region: region}
}

func (m Manager) GetToken(force bool) (string, error) {
	if !force {
		if cached := getCached(m.Region.Value); cached != nil && IsTokenValid(m.Region.Value) {
			return cached.Token, nil
		}
	}
	token, err := m.fetchToken()
	if err != nil {
		return "", err
	}
	if err := setCached(m.Region.Value, token); err != nil {
		return "", err
	}
	return token, nil
}

func (m Manager) fetchToken() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	cookie := config.GetCookie(cfg, m.Region.Value)
	if cookie == "" {
		return "", CookieNotFound(m.Region.Value)
	}
	if m.Region.IsCodebase {
		return m.fetchCodebaseToken(cookie)
	}
	return m.fetchNormalToken(cookie)
}

func (m Manager) fetchNormalToken(cookie string) (string, error) {
	headers := cloneHeaders(normalHeaders)
	headers["Cookie"] = fmt.Sprintf("%s=%s", m.Region.CookieName, cookie)
	resp, err := httpclient.Get(m.Region.AuthURL, headers)
	if err != nil {
		return "", TokenFetch(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", TokenFetch(fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	token := resp.Header.Get("x-jwt-token")
	if token == "" {
		return "", InvalidResponse("响应头中未找到 x-jwt-token")
	}
	return token, nil
}

func (m Manager) fetchCodebaseToken(cookie string) (string, error) {
	headers := cloneHeaders(codebaseHeaders)
	headers["Cookie"] = fmt.Sprintf("%s=%s", m.Region.CookieName, cookie)
	resp, err := httpclient.Get(m.Region.AuthURL, headers)
	if err != nil {
		return "", TokenFetch(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", TokenFetch(fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var payload struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Data    map[string]string `json:"data"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", InvalidResponse(err.Error())
	}
	if payload.Code != 0 {
		if payload.Message == "" {
			payload.Message = "未知错误"
		}
		return "", TokenFetch(payload.Message)
	}
	token := payload.Data["codebase_user_jwt"]
	if token == "" {
		return "", InvalidResponse("JSON 响应中未找到 data.codebase_user_jwt")
	}
	return token, nil
}

func cloneHeaders(headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		out[k] = v
	}
	return out
}

func Success(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
