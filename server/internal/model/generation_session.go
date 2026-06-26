package model

import (
	"time"

	"github.com/google/uuid"
)

// GenerationSession represents a session for generating learning plan tasks
type GenerationSession struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	LearningGoalID uuid.UUID  `json:"learning_goal_id" gorm:"type:uuid;not null;index"`
	Status         string     `json:"status" gorm:"type:varchar(20);not null;default:'generating';check:status IN ('generating', 'completed', 'expired')"`
	TaskCount      int        `json:"task_count" gorm:"not null;default:0"`
	CreatedAt      time.Time  `json:"created_at" gorm:"not null;default:NOW()"`
	ExpiresAt      time.Time  `json:"expires_at" gorm:"not null"`

	// Relationships
	LearningGoal LearningGoal `json:"learning_goal,omitempty" gorm:"foreignKey:LearningGoalID"`
	Tasks        []GeneratedTask `json:"tasks,omitempty" gorm:"foreignKey:SessionID"`
}

// TableName specifies the table name for GenerationSession
func (GenerationSession) TableName() string {
	return "generation_sessions"
}

// IsExpired checks if the session has expired
func (gs *GenerationSession) IsExpired() bool {
	return time.Now().After(gs.ExpiresAt)
}

// IsGenerating checks if the session is still generating
func (gs *GenerationSession) IsGenerating() bool {
	return gs.Status == "generating" && !gs.IsExpired()
}

// IsCompleted checks if the session is completed
func (gs *GenerationSession) IsCompleted() bool {
	return gs.Status == "completed"
}

// MarkCompleted marks the session as completed
func (gs *GenerationSession) MarkCompleted() {
	gs.Status = "completed"
}

// IncrementTaskCount increments the task count
func (gs *GenerationSession) IncrementTaskCount() {
	gs.TaskCount++
}
