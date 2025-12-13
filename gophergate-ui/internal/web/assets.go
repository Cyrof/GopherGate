package web

import "embed"

//go:embed web/templates/*.tmpl
//go:embed web/static/**
var webFS embed.FS