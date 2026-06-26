package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

type LearningGoalRepository struct {
	db *gorm.DB
}

func NewLearningGoalRepository(db *gorm.DB) *LearningGoalRepository {
	return &LearningGoalRepository{db: db}
}

func (r *LearningGoalRepository) Create(goal *model.LearningGoal) error {
	return r.db.Create(goal).Error
}

func (r *LearningGoalRepository) FindByID(id uuid.UUID) (*model.LearningGoal, error) {
	var goal model.LearningGoal
	err := r.db.Where("id = ?", id).First(&goal).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *LearningGoalRepository) ListByUser(userID uuid.UUID, status string, page, pageSize int) ([]model.LearningGoal, int64, error) {
	var goals []model.LearningGoal
	var total int64

	query := r.db.Model(&model.LearningGoal{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&goals).Error

	return goals, total, err
}

func (r *LearningGoalRepository) Update(goal *model.LearningGoal) error {
	return r.db.Save(goal).Error
}
