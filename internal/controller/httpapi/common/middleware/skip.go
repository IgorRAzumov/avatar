package middleware

func ShouldSkipObservability(path string) bool {
	return path == "/health"
}
