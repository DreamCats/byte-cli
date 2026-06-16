package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DreamCats/byte-cli/internal/config"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	browserLoginWait = 5 * time.Minute
	browserPollEvery = time.Second
)

type browserSession struct {
	cmd        *exec.Cmd
	port       int
	profileDir string
}

type browserCommand struct {
	Name string
	Args []string
}

type devtoolsTarget struct {
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type cdpCookie struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Domain  string  `json:"domain"`
	Expires float64 `json:"expires"`
}

type cdpCookieResponse struct {
	ID     int `json:"id"`
	Result struct {
		Cookies []cdpCookie `json:"cookies"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (m Manager) LoginInteractive(out io.Writer) (string, error) {
	if token, err := m.fetchToken(); err == nil {
		if err := setCached(m.Region.Value, token); err != nil {
			return "", err
		}
		return token, nil
	}
	return m.loginViaBrowser(out)
}

func (m Manager) loginViaBrowser(out io.Writer) (string, error) {
	loginURL := m.Region.LoginURL()
	if loginURL == "" {
		return "", TokenFetch("当前区域不支持浏览器登录")
	}
	ctx, cancel := context.WithTimeout(context.Background(), browserLoginWait)
	defer cancel()
	session, err := startBrowserSession(ctx, m.Region, loginURL, out)
	if err != nil {
		return "", TokenFetch(err.Error())
	}
	defer session.Close()
	cookie, err := session.WaitCookie(ctx, m.Region)
	if err != nil {
		return "", TokenFetch(err.Error())
	}
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	config.SetCookie(&cfg, m.Region.Value, cookie.Value)
	if err := config.Save(cfg); err != nil {
		return "", err
	}
	return m.GetToken(true)
}

func startBrowserSession(ctx context.Context, region Region, loginURL string, out io.Writer) (*browserSession, error) {
	browser, err := browserPath()
	if err != nil {
		return nil, err
	}
	port, err := freePort()
	if err != nil {
		return nil, err
	}
	profileBase := filepath.Join(config.ConfigDir(), "browser-login")
	if err := os.MkdirAll(profileBase, 0o700); err != nil {
		return nil, err
	}
	profileDir, err := os.MkdirTemp(profileBase, region.Value+"-")
	if err != nil {
		return nil, err
	}
	args := []string{
		"--new-window",
		"--no-first-run",
		"--no-default-browser-check",
		"--remote-allow-origins=*",
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--user-data-dir=" + profileDir,
		loginURL,
	}
	launch := browserLaunchCommand(browser, args)
	cmd := exec.CommandContext(ctx, launch.Name, launch.Args...)
	if err := cmd.Start(); err != nil {
		_ = os.RemoveAll(profileDir)
		return nil, err
	}
	session := &browserSession{cmd: cmd, port: port, profileDir: profileDir}
	fmt.Fprintf(out, "已打开浏览器登录窗口，请使用账号密码或通行密钥完成 SSO 登录\n")
	fmt.Fprintf(out, "登录成功后会自动继续，不需要手动复制 Cookie\n")
	return session, nil
}

func (s *browserSession) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_ = s.closeBrowser(ctx)
	cancel()
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
	}
	if s.profileDir != "" {
		_ = os.RemoveAll(s.profileDir)
	}
}

func (s *browserSession) closeBrowser(ctx context.Context) error {
	wsURL, err := s.browserWebSocketURL(ctx)
	if err != nil {
		return err
	}
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	_ = wsjson.Write(ctx, conn, map[string]any{"id": 1, "method": "Browser.close"})
	return nil
}

func (s *browserSession) WaitCookie(ctx context.Context, region Region) (cdpCookie, error) {
	wsURL, err := s.pageWebSocketURL(ctx)
	if err != nil {
		return cdpCookie{}, err
	}
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return cdpCookie{}, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	host := authHost(region)
	id := 1
	for {
		cookie, err := readCDPCookie(ctx, conn, id, host, region.CookieName)
		id++
		if err != nil {
			return cdpCookie{}, err
		}
		if cookie.Value != "" {
			return cookie, nil
		}
		if err := sleepContext(ctx, browserPollEvery); err != nil {
			return cdpCookie{}, fmt.Errorf("等待浏览器登录超时")
		}
	}
}

func (s *browserSession) pageWebSocketURL(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/json/list", s.port), nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err == nil {
			targets, readErr := readDevtoolsTargets(resp)
			if readErr != nil {
				return "", readErr
			}
			for _, target := range targets {
				if target.Type == "page" && target.WebSocketDebuggerURL != "" {
					return target.WebSocketDebuggerURL, nil
				}
			}
		}
		if err := sleepContext(ctx, 300*time.Millisecond); err != nil {
			return "", fmt.Errorf("无法连接浏览器调试端口")
		}
	}
}

func (s *browserSession) browserWebSocketURL(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/json/version", s.port), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("浏览器调试端口返回 HTTP %d", resp.StatusCode)
	}
	var payload struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("浏览器调试端口缺少 websocket 地址")
	}
	return payload.WebSocketDebuggerURL, nil
}

func readDevtoolsTargets(resp *http.Response) ([]devtoolsTarget, error) {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("浏览器调试端口返回 HTTP %d", resp.StatusCode)
	}
	var targets []devtoolsTarget
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return nil, err
	}
	return targets, nil
}

func readCDPCookie(ctx context.Context, conn *websocket.Conn, id int, host, name string) (cdpCookie, error) {
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := wsjson.Write(callCtx, conn, map[string]any{"id": id, "method": "Network.getAllCookies"}); err != nil {
		return cdpCookie{}, err
	}
	for {
		var raw map[string]json.RawMessage
		if err := wsjson.Read(callCtx, conn, &raw); err != nil {
			return cdpCookie{}, err
		}
		var gotID int
		if err := json.Unmarshal(raw["id"], &gotID); err != nil || gotID != id {
			continue
		}
		var payload cdpCookieResponse
		data, _ := json.Marshal(raw)
		if err := json.Unmarshal(data, &payload); err != nil {
			return cdpCookie{}, err
		}
		if payload.Error != nil {
			return cdpCookie{}, fmt.Errorf("读取浏览器 Cookie 失败: %s", payload.Error.Message)
		}
		for _, cookie := range payload.Result.Cookies {
			if cdpCookieMatches(cookie, host, name, time.Now()) {
				return cookie, nil
			}
		}
		return cdpCookie{}, nil
	}
}

func cdpCookieMatches(cookie cdpCookie, host, name string, now time.Time) bool {
	if cookie.Name != name || cookie.Value == "" {
		return false
	}
	if cookie.Expires > 0 && int64(cookie.Expires) <= now.Unix() {
		return false
	}
	domain := strings.TrimPrefix(strings.ToLower(cookie.Domain), ".")
	host = strings.ToLower(host)
	return host == domain || strings.HasSuffix(host, "."+domain)
}

func browserPath() (string, error) {
	if value := os.Getenv("BYTE_CLI_BROWSER"); value != "" {
		return value, nil
	}
	for _, candidate := range browserCandidates() {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "microsoft-edge"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("未找到可用浏览器；可通过 BYTE_CLI_BROWSER 指定 Chrome/Chromium/Edge 可执行文件")
}

func browserLaunchCommand(browser string, args []string) browserCommand {
	app := browserAppPath(browser)
	if runtime.GOOS == "darwin" && app != "" {
		openArgs := []string{"-na", app, "--args"}
		openArgs = append(openArgs, args...)
		return browserCommand{Name: "open", Args: openArgs}
	}
	return browserCommand{Name: browser, Args: args}
}

func browserAppPath(browser string) string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	const marker = ".app/Contents/MacOS/"
	idx := strings.Index(browser, marker)
	if idx < 0 {
		return ""
	}
	return browser[:idx+len(".app")]
}

func browserCandidates() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
	}
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func authHost(region Region) string {
	u, err := url.Parse(region.AuthURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
