package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

// GenerationSessionService manages generation session lifecycle
type GenerationSessionService struct {
	sessionRepo *repository.GenerationSessionRepository
	taskRepo    *repository.GeneratedTaskRepository
}

// NewGenerationSessionService creates a new service instance
func NewGenerationSessionService(
	sessionRepo *repository.GenerationSessionRepository,
	taskRepo *repository.GeneratedTaskRepository,
) *GenerationSessionService {
	return &GenerationSessionService{
		sessionRepo: sessionRepo,
		taskRepo:    taskRepo,
	}
}

// CreateSession creates a new generation session
func (s *GenerationSessionService) CreateSession(ctx context.Context, learningGoalID uuid.UUID) (*model.GenerationSession, error) {
	session := &model.GenerationSession{
		ID:             uuid.New(),
		LearningGoalID: learningGoalID,
		Status:         "generating",
		TaskCount:      0,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("[SESSION] Created session %s for learning goal %s", session.ID, learningGoalID)
	return session, nil
}

// GetSession retrieves a session by ID
func (s *GenerationSessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*model.GenerationSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has expired
	if session.IsExpired() {
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}

// GetActiveSession retrieves active session for a learning goal
func (s *GenerationSessionService) GetActiveSession(ctx context.Context, learningGoalID uuid.UUID) (*model.GenerationSession, error) {
	sessions, err := s.sessionRepo.GetByLearningGoalID(ctx, learningGoalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Return the most recent active session
	for _, session := range sessions {
		if session.IsGenerating() {
			return &session, nil
		}
	}

	return nil, nil
}

// AddTask adds a task to the session
func (s *GenerationSessionService) AddTask(ctx context.Context, sessionID uuid.UUID, task *model.GeneratedTask) error {
	// Set session ID
	task.SessionID = sessionID
	task.ID = uuid.New()
	task.CreatedAt = time.Now()

	// Create task
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Increment task count
	if err := s.sessionRepo.IncrementTaskCount(ctx, sessionID); err != nil {
		log.Printf("[WARNING] Failed to increment task count for session %s: %v", sessionID, err)
	}

	return nil
}

// GetTasks retrieves all tasks for a session
func (s *GenerationSessionService) GetTasks(ctx context.Context, sessionID uuid.UUID) ([]model.GeneratedTask, error) {
	return s.taskRepo.GetBySessionID(ctx, sessionID)
}

// CompleteSession marks a session as completed
func (s *GenerationSessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.UpdateStatus(ctx, sessionID, "completed"); err != nil {
		return fmt.Errorf("failed to complete session: %w", err)
	}

	log.Printf("[SESSION] Completed session %s", sessionID)
	return nil
}

// DeleteSession deletes a session and its tasks
func (s *GenerationSessionService) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	// Delete tasks first (should cascade, but explicit is better)
	if err := s.taskRepo.DeleteBySessionID(ctx, sessionID); err != nil {
		log.Printf("[WARNING] Failed to delete tasks for session %s: %v", sessionID, err)
	}

	// Delete session
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	log.Printf("[SESSION] Deleted session %s", sessionID)
	return nil
}

// CleanupExpiredSessions cleans up expired sessions
func (s *GenerationSessionService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	count, err := s.sessionRepo.CleanupExpiredSessions(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	if count > 0 {
		log.Printf("[CLEANUP] Marked %d sessions as expired", count)
	}

	return count, nil
}

// GetSessionStats returns statistics about sessions
func (s *GenerationSessionService) GetSessionStats(ctx context.Context) (map[string]int64, error) {
	return s.sessionRepo.CountByStatus(ctx)
}

// MoveTasksToMainTable moves generated tasks to the main tasks table
func (s *GenerationSessionService) MoveTasksToMainTable(ctx context.Context, sessionID uuid.UUID, learningGoalID uuid.UUID, userID uuid.UUID) (int64, error) {
	count, err := s.taskRepo.MoveToTasks(ctx, sessionID, learningGoalID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to move tasks: %w", err)
	}

	// Mark session as completed
	if err := s.CompleteSession(ctx, sessionID); err != nil {
		log.Printf("[WARNING] Failed to complete session after moving tasks: %v", err)
	}

	log.Printf("[SESSION] Moved %d tasks from session %s to learning goal %s", count, sessionID, learningGoalID)
	return count, nil
}
