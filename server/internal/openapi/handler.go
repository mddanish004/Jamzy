package openapi

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var specJSON []byte

//go:embed docs.html
var docsHTML []byte

func Spec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(specJSON)
}

func Docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(docsHTML)
}
