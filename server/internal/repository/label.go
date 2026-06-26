package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

type LabelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

func (r *LabelRepository) Create(label *model.Label) error {
	return r.db.Create(label).Error
}

func (r *LabelRepository) ListByUser(userID uuid.UUID) ([]model.Label, error) {
	var labels []model.Label
	err := r.db.Where("user_id = ?", userID).Find(&labels).Error
	return labels, err
}

func (r *LabelRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Label{}).Error
}
