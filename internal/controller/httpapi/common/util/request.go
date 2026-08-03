package util

import (
	"net/http"
	"net/url"
)

func PathUserID(request *http.Request) string {
	userID := request.PathValue("user_id")
	if userID == "" {
		return ""
	}
	decoded, err := url.PathUnescape(userID)
	if err != nil {
		return userID
	}
	return decoded
}
