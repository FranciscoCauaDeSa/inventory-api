package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request must contain only one JSON object")
	}

	return nil
}

func decodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError

	if errors.As(err, &maxBytesErr) {
		writeError(w, http.StatusRequestHeaderFieldsTooLarge, "request body too large")
		return
	}

	writeError(w, http.StatusBadRequest, "invalid JSON request body")
}
