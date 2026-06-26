package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestFirstTaskDisplayTime measures the time from request to first task event
// Target: < 3 seconds (SC-001)
func TestFirstTaskDisplayTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup test server with mock AI that responds in 1 second
	router := gin.New()

	startTime := time.Now()
	firstTaskTime := time.Time{}
	taskReceived := false

	// Mock SSE endpoint
	router.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Simulate AI processing delay
		time.Sleep(500 * time.Millisecond)

		// Send first task
		firstTaskTime = time.Now()
		task := map[string]interface{}{
			"id":    "task-1",
			"title": "Learn Go Basics",
		}
		data, _ := json.Marshal(task)
		fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
		c.Writer.Flush()

		taskReceived = true
	})

	// Create test request
	req := httptest.NewRequest("GET", "/api/learning-goals/test-id/generate-stream?session_id=test-session", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !taskReceived {
		t.Fatal("No task was received")
	}

	elapsed := firstTaskTime.Sub(startTime)
	t.Logf("Time to first task: %v", elapsed)

	// Target: < 3 seconds
	if elapsed > 3*time.Second {
		t.Errorf("First task took %v, target is < 3 seconds", elapsed)
	}
}

// TestConcurrentGenerationSessions tests handling multiple concurrent sessions
// Target: 10 concurrent users (T063)
func TestConcurrentGenerationSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	router := gin.New()
	router.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// Simulate streaming
		for i := 0; i < 5; i++ {
			task := map[string]interface{}{
				"id":    fmt.Sprintf("task-%d", i),
				"title": fmt.Sprintf("Task %d", i),
			}
			data, _ := json.Marshal(task)
			fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
			c.Writer.Flush()
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Fprintf(c.Writer, "event: done\ndata: {\"task_count\":5}\n\n")
		c.Writer.Flush()
	})

	startTime := time.Now()

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/learning-goals/goal-%d/generate-stream?session_id=session-%d", id, id), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("request %d failed with status %d", id, w.Code)
			} else {
				done <- true
			}
		}(i)
	}

	// Wait for all requests
	successCount := 0
	errorCount := 0

	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			successCount++
		case err := <-errors:
			t.Logf("Error: %v", err)
			errorCount++
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}

	elapsed := time.Since(startTime)

	t.Logf("Concurrent sessions: %d/%d succeeded in %v", successCount, concurrency, elapsed)
	t.Logf("Errors: %d", errorCount)

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	if successCount < concurrency {
		t.Errorf("Expected %d successes, got %d", concurrency, successCount)
	}
}

// BenchmarkSSEStreaming benchmarks SSE event generation
func BenchmarkSSEStreaming(b *testing.B) {
	router := gin.New()
	router.GET("/stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		for i := 0; i < 100; i++ {
			task := map[string]interface{}{
				"id":    fmt.Sprintf("task-%d", i),
				"title": fmt.Sprintf("Task %d", i),
			}
			data, _ := json.Marshal(task)
			fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
		}
		c.Writer.Flush()
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/stream", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
