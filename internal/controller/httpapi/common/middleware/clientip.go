package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func ClientIP(trustedProxyCIDRs []string) func(http.Handler) http.Handler {
	prefixes := make([]string, 0, len(trustedProxyCIDRs))
	for _, cidr := range trustedProxyCIDRs {
		if trimmed := strings.TrimSpace(cidr); trimmed != "" {
			prefixes = append(prefixes, trimmed)
		}
	}

	if len(prefixes) == 0 {
		return middleware.ClientIPFromRemoteAddr
	}
	return middleware.ClientIPFromXFF(prefixes...)
}

func clientIPKey(request *http.Request) (string, error) {
	address := middleware.GetClientIP(request.Context())
	if address == "" {
		host, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			host = request.RemoteAddr
		}
		address = host
	}
	return httprate.CanonicalizeIP(address), nil
}
