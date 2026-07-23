package model

import (
	"time"
)

type Avatar struct {
	ID               string
	UserID           string
	FileName         string
	MimeType         string
	SizeBytes        int64
	UploadStatus     string
	ProcessingStatus string
	Width            int
	Height           int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type AvatarImage struct {
	Data     []byte
	MimeType string
}

type AvatarUploadEvent struct {
	AvatarID string `json:"avatar_id"`
	UserID   string `json:"user_id"`
}

type AvatarDeleteEvent struct {
	AvatarID string `json:"avatar_id"`
}

type HealthReport struct {
	Status     string
	Components map[string]string
}

const (
	HealthStatusOK       = "ok"
	HealthStatusDegraded = "degraded"
	ComponentStatusOK    = "ok"
	ComponentStatusError = "error"
)
