package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// JSON writes a JSON response with status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes a standardized JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

var nonSlugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a title into a URL-friendly slug.
func Slugify(s string) string {
	clean := strings.ToLower(strings.TrimSpace(s))
	clean = nonSlugRegex.ReplaceAllString(clean, "-")
	return strings.Trim(clean, "-")
}
