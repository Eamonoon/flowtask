package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName  string         `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL    *string        `gorm:"type:varchar(500)" json:"avatar_url"`
	Preferences  JSONB          `gorm:"type:jsonb;default:'{}'" json:"preferences"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type LearningGoal struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Description       string         `gorm:"type:text;not null" json:"description"`
	TargetDuration    *string        `gorm:"type:varchar(50)" json:"target_duration"`
	Status            string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	AIPrompt          *string        `gorm:"type:text" json:"ai_prompt"`
	TaskCount         int            `gorm:"not null;default:0" json:"task_count"`
	CompletedTaskCount int           `gorm:"not null;default:0" json:"completed_task_count"`
	CreatedAt         time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

type Task struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID               uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	LearningGoalID       *uuid.UUID     `gorm:"type:uuid;index" json:"learning_goal_id"`
	ParentTaskID         *uuid.UUID     `gorm:"type:uuid;index" json:"parent_task_id"`
	Title                string         `gorm:"type:varchar(200);not null" json:"title"`
	Description          *string        `gorm:"type:text" json:"description"`
	Status               string         `gorm:"type:varchar(20);not null;default:'todo'" json:"status"`
	Priority             string         `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	Deadline             *time.Time     `json:"deadline"`
	EstimatedDuration    *string        `gorm:"type:varchar(50)" json:"estimated_duration"`
	RecommendedResources *JSONRaw       `gorm:"type:jsonb;default:'[]'" json:"recommended_resources"`
	SortOrder            int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt            time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// SetResourcesJSON sets recommended resources from any JSON-serializable value
func (t *Task) SetResourcesJSON(v interface{}) error {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	raw := JSONRaw(b)
	t.RecommendedResources = &raw
	return nil
}

type TaskDependency struct {
	TaskID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"task_id"`
	DependsOnTaskID uuid.UUID `gorm:"type:uuid;primaryKey" json:"depends_on_task_id"`
}

type Label struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Name   string    `gorm:"type:varchar(50);not null" json:"name"`
	Color  string    `gorm:"type:varchar(7);not null;default:'#6366f1'" json:"color"`
}

type TaskLabel struct {
	TaskID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"task_id"`
	LabelID uuid.UUID `gorm:"type:uuid;primaryKey" json:"label_id"`
}

type StudySession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TaskID    *uuid.UUID `gorm:"type:uuid;index" json:"task_id"`
	Duration  int       `gorm:"not null" json:"duration"`
	Date      time.Time `gorm:"type:date;not null" json:"date"`
	Notes     *string   `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

type AIConversation struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	LearningGoalID *uuid.UUID     `gorm:"type:uuid;index" json:"learning_goal_id"`
	Title          *string        `gorm:"type:varchar(200)" json:"title"`
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type AIMessage struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index" json:"conversation_id"`
	Role           string    `gorm:"type:varchar(20);not null" json:"role"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
}
