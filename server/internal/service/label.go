package service

import (
	"github.com/google/uuid"

	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

type LabelService struct {
	labelRepo *repository.LabelRepository
}

func NewLabelService(labelRepo *repository.LabelRepository) *LabelService {
	return &LabelService{labelRepo: labelRepo}
}

func (s *LabelService) Create(userID uuid.UUID, name, color string) (*model.Label, error) {
	label := &model.Label{
		UserID: userID,
		Name:   name,
		Color:  color,
	}
	if err := s.labelRepo.Create(label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *LabelService) List(userID uuid.UUID) ([]model.Label, error) {
	return s.labelRepo.ListByUser(userID)
}

func (s *LabelService) Delete(id uuid.UUID) error {
	return s.labelRepo.Delete(id)
}
