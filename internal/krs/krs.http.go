package krs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// writeJSON mengirim response JSON dengan status code tertentu.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError mengirim response error dalam format JSON.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// parseDateTime menerima beberapa format tanggal/jam yang umum dipakai di form.
func parseDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("tanggal wajib diisi")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("format tanggal tidak valid: %s", value)
}

func parseOptionalDateTime(value string) (time.Time, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, nil
	}

	t, err := parseDateTime(value)
	if err != nil {
		return time.Time{}, false, err
	}

	return t, true, nil
}
