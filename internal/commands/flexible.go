package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		*i = 0
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*i = flexibleInt(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*i = 0
		return nil
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid integer %q", s)
	}
	*i = flexibleInt(parsed)
	return nil
}

func (i flexibleInt) Int() int { return int(i) }

func (i flexibleInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}
