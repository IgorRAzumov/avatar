package util

import (
	"fmt"
	"strings"
)

const (
	VersionPrefix = "/api/v1"
	AvatarsPrefix = "/api/v1/avatars"

	HealthPath   = "/health"
	LivenessPath = "/health/live"
	DocsPath     = "/docs"
	OpenAPIPath  = "/openapi.yaml"
)

func AvatarRelativePath(id string) string {
	return fmt.Sprintf("%s/%s", AvatarsPrefix, id)
}

func AvatarURL(baseURL, id string) string {
	return fmt.Sprintf("%s%s/%s", strings.TrimRight(baseURL, "/"), AvatarsPrefix, id)
}

func AvatarURLWithParams(baseURL, id, size, format string) string {
	url := AvatarURL(baseURL, id)
	var params []string
	if size != "" {
		params = append(params, "size="+size)
	}
	if format != "" {
		params = append(params, "format="+format)
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}
	return url
}
