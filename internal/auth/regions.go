package auth

import (
	"fmt"
	"strings"
)

type Region struct {
	Value      string
	AuthURL    string
	CookieName string
	IsCodebase bool
}

var authURLs = map[string]string{
	"cn":       "https://cloud.bytedance.net/auth/api/v1/jwt",
	"i18n":     "https://cloud-i18n.bytedance.net/auth/api/v1/jwt",
	"us":       "https://cloud-ttp-us.bytedance.net/auth/api/v1/jwt",
	"eu":       "https://cloud-i18n.tiktok-eu.org/auth/api/v1/jwt",
	"codebase": "https://bits.bytedance.net/api/v1/codebase_token",
}

var allRegionValues = []string{"cn", "i18n", "us", "eu", "codebase"}

func ParseRegion(value string) (Region, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	url, ok := authURLs[v]
	if !ok {
		return Region{}, fmt.Errorf("未知区域: %s，可选: %s", value, strings.Join(allRegionValues, ", "))
	}
	cookieName := "CAS_SESSION"
	if v == "codebase" {
		cookieName = "CAS_SESSION_API"
	}
	return Region{Value: v, AuthURL: url, CookieName: cookieName, IsCodebase: v == "codebase"}, nil
}

func (r Region) LoginURL() string {
	if r.IsCodebase {
		return ""
	}
	return strings.TrimSuffix(r.AuthURL, "/jwt") + "/login"
}

func AllRegions() []Region {
	out := make([]Region, 0, len(allRegionValues))
	for _, value := range allRegionValues {
		region, _ := ParseRegion(value)
		out = append(out, region)
	}
	return out
}
