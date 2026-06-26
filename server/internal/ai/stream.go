package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GeneratedTask struct {
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	EstimatedDuration    string      `json:"estimated_duration"`
	RecommendedResources interface{} `json:"recommended_resources"`
	Subtasks             []GeneratedTask `json:"subtasks,omitempty"`
	Dependencies         []string        `json:"dependencies,omitempty"`
}

type GeneratedResource struct {
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

type GeneratedPlan struct {
	Tasks []GeneratedTask `json:"tasks"`
}

func ParsePlanFromAI(content string) (*GeneratedPlan, error) {
	var plan GeneratedPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("parse AI response: %w", err)
	}
	return &plan, nil
}

// StreamHelper handles Server-Sent Events (SSE) streaming
type StreamHelper struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

// NewStreamHelper creates a new SSE stream helper
func NewStreamHelper(w http.ResponseWriter) (*StreamHelper, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported: ResponseWriter does not implement http.Flusher")
	}

	return &StreamHelper{
		writer:  w,
		flusher: flusher,
	}, nil
}

// SetupSSEHeaders sets the required headers for SSE
func (s *StreamHelper) SetupSSEHeaders() {
	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")
	s.writer.Header().Set("Access-Control-Allow-Origin", "*")
	s.writer.Header().Set("X-Accel-Buffering", "no")
}

// SendEvent sends an SSE event
func (s *StreamHelper) SendEvent(eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = fmt.Fprintf(s.writer, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	s.flusher.Flush()
	return nil
}

// SendTaskEvent sends a task event
func (s *StreamHelper) SendTaskEvent(task interface{}) error {
	return s.SendEvent("task", task)
}

// SendProgressEvent sends a progress event with task count
func (s *StreamHelper) SendProgressEvent(taskCount int) error {
	return s.SendEvent("progress", map[string]interface{}{
		"task_count": taskCount,
	})
}

// SendDoneEvent sends a done event
func (s *StreamHelper) SendDoneEvent(learningGoalID string, taskCount int) error {
	return s.SendEvent("done", map[string]interface{}{
		"learning_goal_id": learningGoalID,
		"task_count":       taskCount,
	})
}

// SendErrorEvent sends an error event
func (s *StreamHelper) SendErrorEvent(code string, message string) error {
	return s.SendEvent("error", map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

// StreamReader reads SSE events from an io.Reader
type StreamReader struct {
	reader io.Reader
}

// NewStreamReader creates a new SSE stream reader
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{reader: r}
}

// ReadEvent reads a single SSE event
func (r *StreamReader) ReadEvent() (eventType string, data []byte, err error) {
	buf := make([]byte, 4096)
	n, err := r.reader.Read(buf)
	if err != nil {
		return "", nil, err
	}

	content := string(buf[:n])
	lines := splitLines(content)

	for _, line := range lines {
		if len(line) > 6 && line[:6] == "event:" {
			eventType = line[6:]
		} else if len(line) > 5 && line[:5] == "data:" {
			data = []byte(line[5:])
		}
	}

	return eventType, data, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
