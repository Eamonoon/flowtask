package ai_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowtask-server/internal/ai"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	config := &ai.RetryConfig{
		MaxRetries:  3,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
	}

	attempts := 0
	fn := func() (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	}

	result, err := ai.RetryWithBackoff(context.Background(), config, fn, "test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%v'", result)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_ExhaustedRetries(t *testing.T) {
	config := &ai.RetryConfig{
		MaxRetries:  2,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
	}

	attempts := 0
	fn := func() (interface{}, error) {
		attempts++
		return nil, errors.New("persistent error")
	}

	_, err := ai.RetryWithBackoff(context.Background(), config, fn, "test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	config := &ai.RetryConfig{
		MaxRetries:  5,
		BaseBackoff: 100 * time.Millisecond,
		MaxBackoff:  1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	fn := func() (interface{}, error) {
		attempts++
		if attempts == 2 {
			cancel()
		}
		return nil, errors.New("error")
	}

	_, err := ai.RetryWithBackoff(ctx, config, fn, "test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if attempts > 3 {
		t.Errorf("Expected cancellation to stop retries, but got %d attempts", attempts)
	}
}

func TestRetryWithMetrics(t *testing.T) {
	config := &ai.RetryConfig{
		MaxRetries:  2,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
	}

	attempts := 0
	fn := func() (interface{}, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	}

	result, metrics, err := ai.RetryWithMetrics(context.Background(), config, fn, "test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%v'", result)
	}
	if metrics.TotalAttempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", metrics.TotalAttempts)
	}
	if metrics.SuccessfulRetries != 1 {
		t.Errorf("Expected 1 successful retry, got %d", metrics.SuccessfulRetries)
	}
}

func TestIsRetryable(t *testing.T) {
	retryableErr := ai.NewRetryableError(errors.New("network error"), "connection failed")
	normalErr := errors.New("normal error")

	if !ai.IsRetryable(retryableErr) {
		t.Error("Expected RetryableError to be retryable")
	}
	if ai.IsRetryable(normalErr) {
		t.Error("Expected normal error to not be retryable")
	}
}
