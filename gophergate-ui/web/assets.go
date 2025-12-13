package web

import "embed"

//go:embed templates/*.tmpl
//go:embed static/**
var FS embed.FS
