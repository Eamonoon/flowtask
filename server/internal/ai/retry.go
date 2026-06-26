package ai

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:  3,
		BaseBackoff: 1 * time.Second,
		MaxBackoff:  4 * time.Second,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() (interface{}, error)

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, config *RetryConfig, fn RetryableFunc, operation string) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		result, err := fn()
		if err == nil {
			if attempt > 0 {
				log.Printf("[RETRY] %s succeeded after %d retries", operation, attempt)
			}
			return result, nil
		}

		lastErr = err

		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
		}

		// Don't wait on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff with exponential increase
		backoff := config.BaseBackoff * time.Duration(1<<uint(attempt))
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}

		log.Printf("[RETRY] %s failed (attempt %d/%d), retrying in %v: %v",
			operation, attempt+1, config.MaxRetries+1, backoff, err)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
			continue
		}
	}

	return nil, fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err     error
	Message string
}

func (e *RetryableError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error, message string) *RetryableError {
	return &RetryableError{Err: err, Message: message}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if _, ok := err.(*RetryableError); ok {
		return true
	}
	// Add more retryable error checks here (network errors, timeouts, etc.)
	return false
}

// RetryMetrics tracks retry statistics
type RetryMetrics struct {
	TotalAttempts   int
	SuccessfulRetries int
	FailedRetries   int
	TotalBackoff    time.Duration
}

// RetryWithMetrics executes a function with retry and tracks metrics
func RetryWithMetrics(ctx context.Context, config *RetryConfig, fn RetryableFunc, operation string) (interface{}, *RetryMetrics, error) {
	metrics := &RetryMetrics{}
	startTime := time.Now()

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		metrics.TotalAttempts++

		result, err := fn()
		if err == nil {
			if attempt > 0 {
				metrics.SuccessfulRetries++
			}
			return result, metrics, nil
		}

		lastErr = err

		if ctx.Err() != nil {
			return nil, metrics, fmt.Errorf("operation cancelled: %w", ctx.Err())
		}

		if attempt == config.MaxRetries {
			metrics.FailedRetries++
			break
		}

		backoff := config.BaseBackoff * time.Duration(1<<uint(attempt))
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}

		metrics.TotalBackoff += backoff

		select {
		case <-ctx.Done():
			return nil, metrics, fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
			continue
		}
	}

	log.Printf("[RETRY_METRICS] %s: attempts=%d, successful_retries=%d, failed_retries=%d, total_backoff=%v, total_time=%v",
		operation, metrics.TotalAttempts, metrics.SuccessfulRetries, metrics.FailedRetries, metrics.TotalBackoff, time.Since(startTime))

	return nil, metrics, fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}
