package middleware

func ShouldSkipObservability(path string) bool {
	return path == "/metrics" || path == "/health"
}
