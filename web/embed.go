// Package web holds the embedded SPA static assets served by the HTTP API.
package web

import "embed"

//go:embed static/*
var Static embed.FS
