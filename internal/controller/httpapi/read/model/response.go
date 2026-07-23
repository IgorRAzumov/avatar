package model

import "time"

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ThumbnailInfo struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

type AvatarMetadataResponse struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	FileName   string          `json:"file_name"`
	MimeType   string          `json:"mime_type"`
	Size       int64           `json:"size"`
	Dimensions Dimensions      `json:"dimensions"`
	Thumbnails []ThumbnailInfo `json:"thumbnails"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type AvatarItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	FileName  string `json:"file_name"`
	Status    string `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}
