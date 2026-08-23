package middleware

import (
	"net/http"
	"time"

	commonmodel "avatar/internal/controller/httpapi/common/model"
	"avatar/internal/controller/httpapi/common/util"

	"github.com/go-chi/httprate"
)

type RateLimitSettings struct {
	Enabled bool
	RPS     int
}

func RateLimit(settings RateLimitSettings) func(http.Handler) http.Handler {
	if !settings.Enabled || settings.RPS <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	limiter := httprate.LimitBy(
		settings.RPS,
		time.Second,
		clientIPKey,
		httprate.WithLimitHandler(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Retry-After", "1")
			util.WriteError(writer, http.StatusTooManyRequests, commonmodel.ErrorResponse{
				Error: commonmodel.MsgTooManyRequests,
			})
		}),
	)

	return func(next http.Handler) http.Handler {
		limited := limiter(next)
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if ShouldSkip(request.URL.Path) {
				next.ServeHTTP(writer, request)
				return
			}
			limited.ServeHTTP(writer, request)
		})
	}
}
