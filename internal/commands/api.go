package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DreamCats/byte-cli/internal/jsonutil"
)

func readJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d%s", resp.StatusCode, responseErrorDetail(resp, data))
	}
	if err := jsonutil.DecodeNormalized(data, dst); err != nil {
		return err
	}
	return nil
}

func readRawJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d%s", resp.StatusCode, responseErrorDetail(resp, data))
	}
	return json.Unmarshal(data, dst)
}

func responseErrorDetail(resp *http.Response, data []byte) string {
	parts := []string{}
	if logID := resp.Header.Get("x-tt-logid"); logID != "" {
		parts = append(parts, "logid="+logID)
	}
	if body := strings.TrimSpace(string(data)); body != "" {
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		parts = append(parts, body)
	}
	if len(parts) == 0 {
		return ""
	}
	return ": " + strings.Join(parts, "; ")
}

func printPrettyJSON(out Output, payload any) (int, error) {
	if err := out.PrintJSON(payload); err != nil {
		return 1, err
	}
	return 0, nil
}
