package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/DreamCats/byte-cli/internal/config"
)

const (
	userAgent        = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	defaultTimeout   = 30 * time.Second
	contentTypeJSON  = "application/json"
	headerUserAgent  = "User-Agent"
	headerContentTyp = "Content-Type"
)

func Request(method, url string, body any, headers map[string]string) (*http.Response, error) {
	if proxy := configuredProxy(); proxy != "" {
		_ = os.Setenv("HTTPS_PROXY", proxy)
		_ = os.Setenv("HTTP_PROXY", proxy)
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerUserAgent, userAgent)
	if body != nil {
		req.Header.Set(headerContentTyp, contentTypeJSON)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := http.Client{Timeout: defaultTimeout}
	return client.Do(req)
}

func Get(url string, headers map[string]string) (*http.Response, error) {
	return Request(http.MethodGet, url, nil, headers)
}

func Post(url string, body any, headers map[string]string) (*http.Response, error) {
	return Request(http.MethodPost, url, body, headers)
}

func configuredProxy() string {
	cfg, err := config.Load()
	if err == nil {
		if cfg.Proxy.HTTPS != "" {
			return cfg.Proxy.HTTPS
		}
		if cfg.Proxy.HTTP != "" {
			return cfg.Proxy.HTTP
		}
	}
	if v := os.Getenv("HTTPS_PROXY"); v != "" {
		return v
	}
	return os.Getenv("HTTP_PROXY")
}
