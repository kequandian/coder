package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter helps manage Server-Sent Events (SSE) responses
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSE writer
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	writer := &SSEWriter{
		w:       w,
		flusher: flusher,
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return writer, nil
}

// WriteEvent writes a single SSE event
func (s *SSEWriter) WriteEvent(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	if err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}

// WriteRaw writes raw data as an SSE event
func (s *SSEWriter) WriteRaw(data string) error {
	_, err := fmt.Fprintf(s.w, "data: %s\n\n", data)
	if err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}

// WriteDone writes the SSE [DONE] marker
func (s *SSEWriter) WriteDone() error {
	_, err := fmt.Fprint(s.w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}
