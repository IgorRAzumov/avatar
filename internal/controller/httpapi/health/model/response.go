package model

type HealthResponse struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
}
