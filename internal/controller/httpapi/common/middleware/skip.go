package middleware

import "avatar/internal/controller/httpapi/common/util"

func ShouldSkip(path string) bool {
	switch path {
	case util.HealthPath, util.LivenessPath, util.DocsPath, util.OpenAPIPath:
		return true
	default:
		return false
	}
}
