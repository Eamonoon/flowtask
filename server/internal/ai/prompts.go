package ai

import "fmt"

const learningPlanSystemPrompt = `你是一个专业的学习规划助手。用户会描述他们想要学习的目标，你需要生成一个结构化的学习计划。

输出格式要求（严格 JSON）：
{
  "tasks": [
    {
      "title": "任务标题",
      "description": "任务描述",
      "estimated_duration": "预计时长（如 1周、3天）",
      "recommended_resources": [
        {"title": "资源名称", "url": "链接（可选）", "description": "简要描述"}
      ],
      "subtasks": [
        {
          "title": "子任务标题",
          "description": "子任务描述",
          "estimated_duration": "预计时长",
          "recommended_resources": []
        }
      ],
      "dependencies": []
    }
  ]
}

规则：
- tasks 按依赖关系排序（被依赖的在前）
- dependencies 使用任务标题表示依赖关系
- 每个任务的 title 必须唯一
- 子任务不需要 dependencies 字段
- 推荐资源优先使用中文`

func BuildLearningPlanPrompt(description string, targetDuration string) []ChatMessage {
	userPrompt := fmt.Sprintf("学习目标：%s\n目标时长：%s\n\n请生成完整的学习计划。", description, targetDuration)

	return []ChatMessage{
		{Role: "system", Content: learningPlanSystemPrompt},
		{Role: "user", Content: userPrompt},
	}
}

const chatSystemPrompt = `你是一个友好的 AI 学习助手。你帮助用户解答学习相关的问题，提供学习建议和资源推荐。

规则：
- 回答要简洁、实用
- 如果用户要求拆解学习计划，返回 JSON 格式的任务列表
- 使用用户相同的语言回答`

func BuildChatPrompt(history []ChatMessage, userMessage string) []ChatMessage {
	messages := []ChatMessage{
		{Role: "system", Content: chatSystemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: userMessage})
	return messages
}

// UserContext holds the user's learning data for contextual prompts
type UserContext struct {
	Goals         []GoalContext
	Tasks         []TaskContext
	StudyMinutes  int
}

type GoalContext struct {
	ID                string
	Description       string
	TargetDuration    string
	Status            string
	TaskCount         int
	CompletedTaskCount int
}

type TaskContext struct {
	ID             string
	Title          string
	Description    string
	Status         string
	Priority       string
	GoalDescription string
}

// BuildContextualChatPrompt builds a chat prompt enriched with user context
func BuildContextualChatPrompt(history []ChatMessage, userMessage string, ctx *UserContext) []ChatMessage {
	systemPrompt := chatSystemPrompt + "\n\n"

	if ctx != nil {
		systemPrompt += "## 用户的学习数据\n\n"

		if len(ctx.Goals) > 0 {
			systemPrompt += "### 学习目标\n"
			for _, g := range ctx.Goals {
				progress := fmt.Sprintf("%d/%d", g.CompletedTaskCount, g.TaskCount)
				systemPrompt += fmt.Sprintf("- 【%s】%s（进度：%s，状态：%s）\n",
					g.ID[:8], g.Description, progress, g.Status)
			}
			systemPrompt += "\n"
		}

		if len(ctx.Tasks) > 0 {
			systemPrompt += "### 任务列表\n"
			for _, t := range ctx.Tasks {
				statusLabel := map[string]string{
					"todo": "待办", "doing": "进行中", "done": "已完成",
				}[t.Status]
				if statusLabel == "" {
					statusLabel = t.Status
				}
				goalInfo := ""
				if t.GoalDescription != "" {
					goalInfo = fmt.Sprintf(" [目标: %s]", t.GoalDescription)
				}
				systemPrompt += fmt.Sprintf("- [%s] %s（优先级：%s，状态：%s）%s\n",
					t.ID[:8], t.Title, t.Priority, statusLabel, goalInfo)
			}
			systemPrompt += "\n"
		}

		if ctx.StudyMinutes > 0 {
			systemPrompt += fmt.Sprintf("### 近期学习时长\n最近 7 天累计学习 %d 分钟\n\n", ctx.StudyMinutes)
		}

		systemPrompt += "请基于以上用户数据回答问题。当用户提到学习计划或任务时，直接引用具体的目标和任务名称。\n"
	}

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: userMessage})
	return messages
}

const dailySummarySystemPrompt = `你是一个学习总结助手。根据用户当日的学习数据（完成的任务、学习时长），生成简洁的每日学习总结和明日建议。

输出格式：
- 今日完成情况概述
- 学习时长统计
- 明日建议（基于未完成任务和学习进度）`

func BuildDailySummaryPrompt(date string, completedTasks string, studyMinutes int, pendingTasks string) []ChatMessage {
	userPrompt := fmt.Sprintf("日期：%s\n已完成任务：\n%s\n学习时长：%d 分钟\n未完成任务：\n%s\n\n请生成今日学习总结。",
		date, completedTasks, studyMinutes, pendingTasks)

	return []ChatMessage{
		{Role: "system", Content: dailySummarySystemPrompt},
		{Role: "user", Content: userPrompt},
	}
}
