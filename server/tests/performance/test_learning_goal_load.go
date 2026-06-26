package performance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestConcurrentGenerationSessions_LoadTest tests handling multiple concurrent sessions
// Target: 10 concurrent users
func TestConcurrentGenerationSessions_LoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	concurrency := 10
	tasksPerSession := 5

	// Track active sessions
	activeSessions := sync.Map{}

	router := gin.New()
	router.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
		sessionID := c.Query("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
			return
		}

		// Track session
		activeSessions.Store(sessionID, true)
		defer activeSessions.Delete(sessionID)

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Simulate streaming tasks
		for i := 0; i < tasksPerSession; i++ {
			task := map[string]interface{}{
				"id":    fmt.Sprintf("%s-task-%d", sessionID, i),
				"title": fmt.Sprintf("Task %d", i),
			}
			data, _ := json.Marshal(task)
			fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
			c.Writer.Flush()
			time.Sleep(100 * time.Millisecond)
		}

		// Send done event
		doneData, _ := json.Marshal(map[string]interface{}{
			"task_count": tasksPerSession,
		})
		fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", doneData)
		c.Writer.Flush()
	})

	var wg sync.WaitGroup
	errors := make(chan error, concurrency)
	results := make(chan struct {
		sessionID string
		taskCount int
		duration  time.Duration
	}, concurrency)

	startTime := time.Now()

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sessionStart := time.Now()
			sessionID := fmt.Sprintf("session-%d", id)
			url := fmt.Sprintf("/api/learning-goals/goal-%d/generate-stream?session_id=%s", id, sessionID)

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("session %d: unexpected status %d", id, w.Code)
				return
			}

			// Count task events
			body := w.Body.String()
			taskCount := 0
			for _, line := range splitLines(body) {
				if len(line) > 6 && line[:6] == "event:" && line[6:] == "task" {
					taskCount++
				}
			}

			results <- struct {
				sessionID string
				taskCount int
				duration  time.Duration
			}{
				sessionID: sessionID,
				taskCount: taskCount,
				duration:  time.Since(sessionStart),
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(results)
	close(errors)

	totalDuration := time.Since(startTime)

	// Collect results
	successCount := 0
	errorCount := 0
	totalTasks := 0

	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	for result := range results {
		successCount++
		totalTasks += result.taskCount
		t.Logf("Session %s: %d tasks in %v", result.sessionID, result.taskCount, result.duration)
	}

	t.Logf("=== Load Test Results ===")
	t.Logf("Total sessions: %d", concurrency)
	t.Logf("Successful sessions: %d", successCount)
	t.Logf("Failed sessions: %d", errorCount)
	t.Logf("Total tasks generated: %d", totalTasks)
	t.Logf("Total duration: %v", totalDuration)
	t.Logf("Average time per session: %v", totalDuration/time.Duration(concurrency))

	// Assertions
	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	if successCount < concurrency {
		t.Errorf("Expected %d successful sessions, got %d", concurrency, successCount)
	}

	if totalTasks != concurrency*tasksPerSession {
		t.Errorf("Expected %d total tasks, got %d", concurrency*tasksPerSession, totalTasks)
	}

	// Target: All sessions should complete within reasonable time
	if totalDuration > 30*time.Second {
		t.Errorf("Total duration %v exceeds 30 second target", totalDuration)
	}
}

// splitLines splits a string into lines
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
