package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DreamCats/byte-cli/internal/jsonutil"
)

func readJSON(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
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
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.Unmarshal(data, dst)
}

func printPrettyJSON(out Output, payload any) (int, error) {
	if err := out.PrintJSON(payload); err != nil {
		return 1, err
	}
	return 0, nil
}
