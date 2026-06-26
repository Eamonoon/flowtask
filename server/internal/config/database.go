package config

import (
	"fmt"
	"log"

	"flowtask-server/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	log.Println("Database connected successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.LearningGoal{},
		&model.Task{},
		&model.TaskDependency{},
		&model.Label{},
		&model.TaskLabel{},
		&model.StudySession{},
		&model.AIConversation{},
		&model.AIMessage{},
		&model.GenerationSession{},
		&model.GeneratedTask{},
	)
}
