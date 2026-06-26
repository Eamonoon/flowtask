package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"flowtask-server/internal/ai"
	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

const dashboardCachePrefix = "dashboard:"
const dashboardCacheTTL = 60 * time.Second

type DashboardService struct {
	taskRepo    *repository.TaskRepository
	sessionRepo *repository.StudySessionRepository
	rdb         *redis.Client
}

func NewDashboardService(
	taskRepo *repository.TaskRepository,
	sessionRepo *repository.StudySessionRepository,
	rdb *redis.Client,
) *DashboardService {
	return &DashboardService{
		taskRepo:    taskRepo,
		sessionRepo: sessionRepo,
		rdb:         rdb,
	}
}

func (s *DashboardService) InvalidateCache(userID uuid.UUID) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	key := fmt.Sprintf("%s%s", dashboardCachePrefix, userID.String())
	s.rdb.Del(ctx, key)
}

func (s *DashboardService) GetStats(userID uuid.UUID) (map[string]interface{}, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("%s%s", dashboardCachePrefix, userID.String())

	// Try to read from cache
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var result map[string]interface{}
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result, nil
			}
		}
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := todayStart.AddDate(0, -1, 0)

	todayTasks, _, _ := s.taskRepo.ListByUser(userID, repository.TaskListOptions{Limit: 100})
	todayCount := 0
	completedToday := 0
	for _, t := range todayTasks {
		if t.CreatedAt.After(todayStart) || t.CreatedAt.Equal(todayStart) {
			todayCount++
			if t.Status == "done" {
				completedToday++
			}
		}
	}

	allTasks, _, _ := s.taskRepo.ListByUser(userID, repository.TaskListOptions{Limit: 1000})
	totalTasks := len(allTasks)
	completedTasks := 0
	for _, t := range allTasks {
		if t.Status == "done" {
			completedTasks++
		}
	}

	todayMinutes, _ := s.sessionRepo.GetTotalMinutesByDateRange(userID, todayStart, now)
	weekMinutes, _ := s.sessionRepo.GetTotalMinutesByDateRange(userID, weekStart, now)
	monthMinutes, _ := s.sessionRepo.GetTotalMinutesByDateRange(userID, monthStart, now)

	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks)
	}

	// Collect upcoming deadlines (tasks with deadline in the future, not done)
	var upcomingDeadlines []map[string]interface{}
	for _, t := range allTasks {
		if t.Deadline != nil && t.Status != "done" {
			upcomingDeadlines = append(upcomingDeadlines, map[string]interface{}{
				"id":                t.ID.String(),
				"title":             t.Title,
				"description":       t.Description,
				"status":            t.Status,
				"priority":          t.Priority,
				"deadline":          t.Deadline,
				"estimated_duration": t.EstimatedDuration,
				"learning_goal_id":  t.LearningGoalID,
				"sort_order":        t.SortOrder,
				"created_at":        t.CreatedAt,
				"updated_at":        t.UpdatedAt,
			})
		}
	}
	// Sort by deadline ascending, limit to 10
	if len(upcomingDeadlines) > 10 {
		upcomingDeadlines = upcomingDeadlines[:10]
	}

	// Collect recent activity (last 10 items)
	var recentActivity []map[string]interface{}
	// Recent completed tasks
	recentCompleted := 0
	for _, t := range allTasks {
		if t.Status == "done" && recentCompleted < 5 {
			recentActivity = append(recentActivity, map[string]interface{}{
				"type":        "task_completed",
				"description": "完成任务: " + t.Title,
				"timestamp":   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
			recentCompleted++
		}
	}
	// Recent study sessions
	recentSessions, _ := s.sessionRepo.ListByUser(userID, &weekStart, &now, nil)
	for i, sess := range recentSessions {
		if i >= 5 {
			break
		}
		recentActivity = append(recentActivity, map[string]interface{}{
			"type":        "study_session",
			"description": "学习了 " + strconv.Itoa(sess.Duration) + " 分钟",
			"timestamp":   sess.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	result := map[string]interface{}{
		"today_tasks": map[string]interface{}{
			"total":     todayCount,
			"completed": completedToday,
		},
		"overall": map[string]interface{}{
			"total_tasks":     totalTasks,
			"completed_tasks": completedTasks,
			"completion_rate": completionRate,
		},
		"study_time": map[string]interface{}{
			"today_minutes": todayMinutes,
			"week_minutes":  weekMinutes,
			"month_minutes": monthMinutes,
		},
		"upcoming_deadlines": upcomingDeadlines,
		"recent_activity":    recentActivity,
	}

	// Cache the result
	if s.rdb != nil {
		if data, err := json.Marshal(result); err == nil {
			s.rdb.Set(ctx, cacheKey, string(data), dashboardCacheTTL)
		}
	}

	return result, nil
}

func (s *DashboardService) GetStudyTimeChart(userID uuid.UUID, period string) (map[string]interface{}, error) {
	now := time.Now()
	var days int
	switch period {
	case "month":
		days = 30
	default:
		days = 7
	}

	labels := make([]string, days)
	values := make([]int, days)
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -(days - 1 - i))
		labels[i] = date.Format("01-02")
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)
		minutes, _ := s.sessionRepo.GetTotalMinutesByDateRange(userID, dayStart, dayEnd)
		values[i] = minutes
	}

	return map[string]interface{}{
		"labels": labels,
		"values": values,
	}, nil
}

func (s *DashboardService) GetCategoryStats(userID uuid.UUID) ([]map[string]interface{}, error) {
	allTasks, _, _ := s.taskRepo.ListByUser(userID, repository.TaskListOptions{Limit: 1000})

	categoryMap := make(map[string]map[string]int)
	for _, t := range allTasks {
		goalID := "ungrouped"
		if t.LearningGoalID != nil {
			goalID = t.LearningGoalID.String()
		}
		if _, ok := categoryMap[goalID]; !ok {
			categoryMap[goalID] = map[string]int{"count": 0, "completed": 0}
		}
		categoryMap[goalID]["count"]++
		if t.Status == "done" {
			categoryMap[goalID]["completed"]++
		}
	}

	var result []map[string]interface{}
	for label, stats := range categoryMap {
		result = append(result, map[string]interface{}{
			"label":     label,
			"count":     stats["count"],
			"completed": stats["completed"],
		})
	}
	return result, nil
}

func (s *DashboardService) GetCompletionRateChart(userID uuid.UUID, period string) (map[string]interface{}, error) {
	now := time.Now()
	var days int
	switch period {
	case "month":
		days = 30
	default:
		days = 7
	}

	labels := make([]string, days)
	values := make([]float64, days)
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -(days - 1 - i))
		labels[i] = date.Format("01-02")
		_ = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		values[i] = 0.0
	}

	return map[string]interface{}{
		"labels": labels,
		"values": values,
	}, nil
}

type AIChatService struct {
	convRepo    *repository.AIConversationRepository
	taskRepo    *repository.TaskRepository
	goalRepo    *repository.LearningGoalRepository
	sessionRepo *repository.StudySessionRepository
	aiClient    *ai.Client
}

func NewAIChatService(
	convRepo *repository.AIConversationRepository,
	taskRepo *repository.TaskRepository,
	goalRepo *repository.LearningGoalRepository,
	sessionRepo *repository.StudySessionRepository,
	aiClient *ai.Client,
) *AIChatService {
	return &AIChatService{
		convRepo:    convRepo,
		taskRepo:    taskRepo,
		goalRepo:    goalRepo,
		sessionRepo: sessionRepo,
		aiClient:    aiClient,
	}
}

type ChatInput struct {
	ConversationID *uuid.UUID `json:"conversation_id"`
	LearningGoalID *uuid.UUID `json:"learning_goal_id"`
	Message        string     `json:"message" binding:"required"`
}

func (s *AIChatService) Chat(userID uuid.UUID, input ChatInput, onDelta func(content string) error, onDone func(fullContent string, convID uuid.UUID) error) error {
	var conv *model.AIConversation
	var err error
	isNewConv := false

	if input.ConversationID != nil {
		conv, err = s.convRepo.FindByID(*input.ConversationID)
		if err != nil {
			return fmt.Errorf("conversation not found: %w", err)
		}
	} else {
		title := truncateTitle(input.Message, 30)
		conv = &model.AIConversation{
			UserID:         userID,
			LearningGoalID: input.LearningGoalID,
			Title:          &title,
		}
		if err := s.convRepo.Create(conv); err != nil {
			return fmt.Errorf("create conversation: %w", err)
		}
		isNewConv = true
	}

	// For existing conversations with no title, set it from the first message
	if !isNewConv && conv.Title == nil {
		title := truncateTitle(input.Message, 30)
		s.convRepo.UpdateTitle(conv.ID, title)
	}

	s.convRepo.CreateMessage(&model.AIMessage{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        input.Message,
	})

	messages, _ := s.convRepo.GetMessages(conv.ID)
	var history []ai.ChatMessage
	for _, msg := range messages {
		history = append(history, ai.ChatMessage{Role: msg.Role, Content: msg.Content})
	}

	// Build user context from database
	userCtx := s.buildUserContext(userID)

	prompt := ai.BuildContextualChatPrompt(history[:len(history)-1], input.Message, userCtx)

	var fullContent string
	err = s.aiClient.ChatStream(prompt, func(delta string) error {
		fullContent += delta
		return onDelta(delta)
	})
	if err != nil {
		return err
	}

	s.convRepo.CreateMessage(&model.AIMessage{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        fullContent,
	})

	return onDone(fullContent, conv.ID)
}

func (s *AIChatService) buildUserContext(userID uuid.UUID) *ai.UserContext {
	ctx := &ai.UserContext{}

	// Fetch learning goals
	goals, _, _ := s.goalRepo.ListByUser(userID, "", 1, 20)
	for _, g := range goals {
		ctx.Goals = append(ctx.Goals, ai.GoalContext{
			ID:                 g.ID.String(),
			Description:        g.Description,
			TargetDuration:     strOrDefault(g.TargetDuration),
			Status:             g.Status,
			TaskCount:          g.TaskCount,
			CompletedTaskCount: g.CompletedTaskCount,
		})
	}

	// Fetch all tasks
	allTasks, _, _ := s.taskRepo.ListByUser(userID, repository.TaskListOptions{Limit: 100})
	// Build goal description map for task context
	goalDescMap := make(map[string]string)
	for _, g := range goals {
		goalDescMap[g.ID.String()] = g.Description
	}
	for _, t := range allTasks {
		goalDesc := ""
		if t.LearningGoalID != nil {
			goalDesc = goalDescMap[t.LearningGoalID.String()]
		}
		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}
		ctx.Tasks = append(ctx.Tasks, ai.TaskContext{
			ID:              t.ID.String(),
			Title:           t.Title,
			Description:     desc,
			Status:          t.Status,
			Priority:        t.Priority,
			GoalDescription: goalDesc,
		})
	}

	// Fetch recent study time
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	minutes, _ := s.sessionRepo.GetTotalMinutesByDateRange(userID, weekAgo, now)
	ctx.StudyMinutes = minutes

	return ctx
}

func strOrDefault(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func truncateTitle(msg string, maxLen int) string {
	// Take first line, trim whitespace
	title := msg
	if idx := indexOfNewline(msg); idx > 0 {
		title = msg[:idx]
	}
	title = trimWhitespace(title)
	runes := []rune(title)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	if title == "" {
		return "新对话"
	}
	return title
}

func indexOfNewline(s string) int {
	for i, c := range s {
		if c == '\n' || c == '\r' {
			return i
		}
	}
	return -1
}

func trimWhitespace(s string) string {
	runes := []rune(s)
	start, end := 0, len(runes)-1
	for start <= end && runes[start] == ' ' {
		start++
	}
	for end >= start && runes[end] == ' ' {
		end--
	}
	if start > end {
		return ""
	}
	return string(runes[start : end+1])
}

func (s *AIChatService) ListConversations(userID uuid.UUID) ([]model.AIConversation, error) {
	return s.convRepo.ListByUser(userID)
}

func (s *AIChatService) GetMessages(conversationID uuid.UUID) ([]model.AIMessage, error) {
	return s.convRepo.GetMessages(conversationID)
}

func (s *AIChatService) DeleteConversation(userID, convID uuid.UUID) error {
	conv, err := s.convRepo.FindByID(convID)
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}
	if conv.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.convRepo.Delete(convID)
}

type StudySessionService struct {
	sessionRepo *repository.StudySessionRepository
}

func NewStudySessionService(sessionRepo *repository.StudySessionRepository) *StudySessionService {
	return &StudySessionService{sessionRepo: sessionRepo}
}

type CreateSessionInput struct {
	TaskID   *uuid.UUID `json:"task_id"`
	Duration int        `json:"duration" binding:"required,min=1"`
	Date     string     `json:"date" binding:"required"`
	Notes    string     `json:"notes"`
}

func (s *StudySessionService) Create(userID uuid.UUID, input CreateSessionInput) (*model.StudySession, error) {
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	session := &model.StudySession{
		UserID:   userID,
		TaskID:   input.TaskID,
		Duration: input.Duration,
		Date:     date,
	}
	if input.Notes != "" {
		session.Notes = &input.Notes
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *StudySessionService) List(userID uuid.UUID, startDate, endDate *time.Time, taskID *uuid.UUID) ([]model.StudySession, error) {
	return s.sessionRepo.ListByUser(userID, startDate, endDate, taskID)
}

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

type UpdateProfileInput struct {
	DisplayName *string         `json:"display_name"`
	AvatarURL   *string         `json:"avatar_url"`
	Preferences *model.JSONB    `json:"preferences"`
}

func (s *UserService) UpdateProfile(userID uuid.UUID, input UpdateProfileInput) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.Preferences != nil {
		for k, v := range *input.Preferences {
			user.Preferences[k] = v
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}
