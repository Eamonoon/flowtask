package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

type StudySessionRepository struct {
	db *gorm.DB
}

func NewStudySessionRepository(db *gorm.DB) *StudySessionRepository {
	return &StudySessionRepository{db: db}
}

func (r *StudySessionRepository) Create(session *model.StudySession) error {
	return r.db.Create(session).Error
}

func (r *StudySessionRepository) ListByUser(userID uuid.UUID, startDate, endDate *time.Time, taskID *uuid.UUID) ([]model.StudySession, error) {
	var sessions []model.StudySession
	query := r.db.Where("user_id = ?", userID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}
	if taskID != nil {
		query = query.Where("task_id = ?", taskID)
	}
	err := query.Order("date DESC").Find(&sessions).Error
	return sessions, err
}

func (r *StudySessionRepository) GetTotalMinutesByDateRange(userID uuid.UUID, start, end time.Time) (int, error) {
	var total int
	err := r.db.Model(&model.StudySession{}).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, start, end).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&total).Error
	return total, err
}
