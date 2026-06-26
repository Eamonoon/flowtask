package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

type AIConversationRepository struct {
	db *gorm.DB
}

func NewAIConversationRepository(db *gorm.DB) *AIConversationRepository {
	return &AIConversationRepository{db: db}
}

func (r *AIConversationRepository) Create(conv *model.AIConversation) error {
	return r.db.Create(conv).Error
}

func (r *AIConversationRepository) FindByID(id uuid.UUID) (*model.AIConversation, error) {
	var conv model.AIConversation
	err := r.db.Where("id = ?", id).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *AIConversationRepository) ListByUser(userID uuid.UUID) ([]model.AIConversation, error) {
	var convs []model.AIConversation
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&convs).Error
	return convs, err
}

func (r *AIConversationRepository) CreateMessage(msg *model.AIMessage) error {
	return r.db.Create(msg).Error
}

func (r *AIConversationRepository) GetMessages(conversationID uuid.UUID) ([]model.AIMessage, error) {
	var messages []model.AIMessage
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *AIConversationRepository) UpdateTitle(id uuid.UUID, title string) error {
	return r.db.Model(&model.AIConversation{}).
		Where("id = ?", id).
		Update("title", title).Error
}

func (r *AIConversationRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", id).Delete(&model.AIMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.AIConversation{}).Error
	})
}
