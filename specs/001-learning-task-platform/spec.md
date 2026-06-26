# Feature Specification: FlowTask - AI Learning Plan & Task Management Platform

**Feature Branch**: `001-learning-task-platform`

**Created**: 2026-06-17

**Status**: Draft

**Input**: User description: "开发一个 AI 驱动的学习计划和任务管理平台"

## User Scenarios & Testing

### User Story 1 - User Registration & Authentication (Priority: P1)

A new user visits FlowTask and creates an account using email and password. After registration, they are automatically logged in and can access the platform. The system maintains their session using JWT tokens, with automatic token refresh so the user never needs to re-login during active use.

**Why this priority**: Authentication is foundational — no other feature can be used without it.

**Independent Test**: Can be fully tested by registering a new account, logging in, accessing a protected page, and verifying session persistence across browser refreshes.

**Acceptance Scenarios**:

1. **Given** a user is on the registration page, **When** they enter valid email and password and submit, **Then** their account is created and they are logged in automatically
2. **Given** a user has an account, **When** they enter correct credentials on the login page, **Then** they are authenticated and redirected to the dashboard
3. **Given** a user is logged in, **When** their access token expires, **Then** the system automatically refreshes the token without interrupting their workflow
4. **Given** a user enters an email that already exists during registration, **When** they submit the form, **Then** they see a clear error message indicating the email is taken
5. **Given** a user enters an invalid email format or weak password, **When** they submit the registration form, **Then** they see inline validation errors before submission

---

### User Story 2 - AI Learning Goal Generation (Priority: P1)

A user describes a learning goal in natural language (e.g., "我想两个月学会 RAG"). The AI analyzes the goal and automatically generates a structured learning plan with a clear roadmap, ordered tasks, estimated durations, and recommended resources. The generated plan appears as an interactive task tree.

**Why this priority**: This is the core differentiator of FlowTask — the AI-powered learning plan generation is what makes this platform unique.

**Independent Test**: Can be tested by entering a learning goal, verifying a structured plan is generated with tasks, subtasks, durations, and resources, and confirming the plan is saved and accessible.

**Acceptance Scenarios**:

1. **Given** a user is on the learning goal page, **When** they enter "我想两个月学会 RAG", **Then** the AI generates a complete learning plan with tasks, subtasks, estimated durations, and recommended resources
2. **Given** an AI-generated learning plan is displayed, **When** the user views it, **Then** tasks are shown in a hierarchical tree structure with dependency-based ordering (prerequisites before dependent topics), and independent tasks can be worked on in any order
3. **Given** an AI-generated plan exists, **When** the user clicks on any task, **Then** they see full details including description, estimated time, resources, and status
4. **Given** the AI is generating a plan, **When** the user is waiting, **Then** they see streaming output with content appearing progressively
5. **Given** a user wants to adjust a generated plan, **When** they edit task details (title, description, duration), **Then** the changes are saved and reflected immediately

---

### User Story 3 - Task Management (Priority: P1)

A user manages their learning tasks through a task board. They can create tasks manually, edit existing tasks, change status (Todo/Doing/Done), set priorities, add labels, and set deadlines. Tasks can be searched, filtered, and sorted. Subtasks are supported for breaking down complex tasks.

**Why this priority**: Task management is the daily-use core of the platform — users interact with it constantly.

**Independent Test**: Can be tested by creating, editing, deleting, searching, filtering, and sorting tasks, as well as managing subtasks and status transitions.

**Acceptance Scenarios**:

1. **Given** a user is on the task page, **When** they click "New Task" and fill in title, description, priority, and deadline, **Then** the task is created and appears in the task list
2. **Given** a task exists, **When** the user drags it to a different status column (Todo/Doing/Done), **Then** the task status updates immediately
3. **Given** a task has incomplete prerequisite tasks, **When** the user moves it to Done, **Then** the system blocks completion and shows a clear message listing the unmet prerequisites
4. **Given** multiple tasks exist, **When** the user searches by keyword, **Then** only matching tasks are displayed
5. **Given** multiple tasks exist, **When** the user applies filters (priority, status, label, deadline date range), **Then** only matching tasks are shown
6. **Given** a task has subtasks, **When** the user views the task, **Then** they see all subtasks with individual completion status and overall progress
7. **Given** a user wants to delete a task, **When** they confirm deletion, **Then** the task and all its subtasks are removed

---

### User Story 4 - Dashboard Overview (Priority: P2)

A user opens the dashboard and immediately sees their learning progress: today's tasks, completion rate, recent activity, and upcoming deadlines. Charts visualize daily study time trends, completion rates over time, and category breakdowns.

**Why this priority**: The dashboard provides motivation and overview but is not required for core task management functionality.

**Independent Test**: Can be tested by completing various tasks over time, then verifying the dashboard displays accurate statistics and charts reflecting that activity.

**Acceptance Scenarios**:

1. **Given** a user has tasks in various states, **When** they open the dashboard, **Then** they see today's tasks, completed task count, and overall completion rate
2. **Given** a user has recorded study sessions, **When** they view the dashboard, **Then** they see a chart of daily study time over the past week/month
3. **Given** a user has tasks with upcoming deadlines, **When** they open the dashboard, **Then** they see a list of tasks due soon, sorted by deadline
4. **Given** a user has been active, **When** they view the dashboard, **Then** they see their recent activity log (task completions, plan creations, study sessions)

---

### User Story 5 - AI Chat Assistant (Priority: P2)

A user opens the AI assistant chat and asks learning-related questions (e.g., "Docker 怎么学习?", "RAG 应该先学什么?"). The AI responds with helpful, contextual answers. The assistant can also generate task breakdowns on request.

**Why this priority**: The chat assistant enhances the learning experience but the core AI functionality (plan generation) works without it.

**Independent Test**: Can be tested by opening the chat, asking various learning questions, and verifying the AI provides relevant, helpful responses with streaming output.

**Acceptance Scenarios**:

1. **Given** a user opens the AI assistant, **When** they type a question about a technology, **Then** the AI responds with a helpful answer with streaming output
2. **Given** a user asks "我想学习 Kubernetes", **When** the AI responds, **Then** it generates a learning roadmap with actionable tasks
3. **Given** a user has existing learning goals and task history, **When** they ask for recommendations, **Then** the AI provides context-aware suggestions based on their progress
4. **Given** the AI is generating a response, **When** the user is waiting, **Then** they see a streaming response with content appearing progressively

---

### User Story 6 - AI Daily Summary (Priority: P3)

At the end of each day (or on demand), the AI generates a summary of the user's learning activity: what was completed, total study time, and suggestions for the next day.

**Why this priority**: Nice-to-have feature that adds value but is not essential for the core workflow.

**Independent Test**: Can be tested by completing tasks during a day, requesting a summary, and verifying it accurately reflects the day's activity with actionable next-day suggestions.

**Acceptance Scenarios**:

1. **Given** a user has completed tasks today, **When** they request a daily summary, **Then** the AI generates a summary listing completed tasks, total study time, and tomorrow's suggested focus areas
2. **Given** a user has not completed any tasks today, **When** they request a summary, **Then** the AI gently encourages them and suggests starting with a specific task

---

### User Story 7 - User Profile Management (Priority: P3)

A user views and edits their profile: display name, avatar, and learning preferences. In v1, these preferences include both UI preferences and AI recommendation preferences, so later plan generation and assistant suggestions can adapt to how the user prefers to learn.

**Why this priority**: Profile management is a standard feature but not critical for the core learning workflow.

**Independent Test**: Can be tested by editing profile fields and verifying changes persist across sessions.

**Acceptance Scenarios**:

1. **Given** a user is on their profile page, **When** they update their display name and save, **Then** the change is reflected across the platform
2. **Given** a user changes their avatar, **When** they save, **Then** the new avatar appears on their profile and dashboard
3. **Given** a user updates their learning preferences, **When** they save, **Then** the updated preferences persist across sessions and influence later AI recommendations and learning plan generation

---

### Edge Cases

- What happens when the AI service is unavailable or times out during plan generation? The system automatically retries with exponential backoff (1s/2s/4s intervals, max 3 attempts). If all retries fail, it displays an error message and allows the user to retry manually without losing their input.
- What happens when a user tries to create a task with a deadline in the past? The system should allow it but display a warning.
- What happens when a user deletes a parent task that has incomplete subtasks? All subtasks should be deleted with a confirmation prompt.
- What happens when a user has over 100 tasks? Infinite scrolling ensures the UI remains responsive by loading tasks in batches as the user scrolls.
- What happens when the AI generates a learning plan but the user wants to significantly restructure it? The user should be able to freely edit, reorder, add, or remove tasks from the generated plan.
- What happens when the JWT refresh token expires (user inactive for extended period)? The user should be redirected to the login page with a message.
- What happens when a user tries to mark a task as Done before its prerequisite tasks are completed? The system should allow moving it to Doing, but block the transition to Done and show which prerequisite tasks remain incomplete.

## Requirements

### Functional Requirements

- **FR-001**: System MUST allow users to register with email and password
- **FR-002**: System MUST authenticate users using JWT with automatic token refresh
- **FR-003**: System MUST allow users to view and edit their profile (name, avatar, preferences including theme, language, learning_style, weekly_study_hours, and preferred_session_minutes)
- **FR-004**: System MUST allow users to input a learning goal in natural language and receive an AI-generated structured learning plan. Users can have multiple active learning plans simultaneously.
- **FR-005**: System MUST generate learning plans that include: task hierarchy, estimated duration per task, recommended resources, dependency-based ordering (topological sort), and support for flexible independent task completion
- **FR-006**: System MUST support streaming output for all AI-generated content
- **FR-007**: System MUST allow users to create, edit, and delete tasks with fields: title, description, status (Todo/Doing/Done), priority, labels, deadline, and subtasks
- **FR-008**: System MUST support task search by keyword across title and description
- **FR-009**: System MUST support task filtering by status, priority, labels, and deadline date range
- **FR-010**: System MUST support task sorting by creation date, deadline, priority, and status
- **FR-011**: System MUST display a dashboard with: today's tasks (across all active plans), completed count, completion rate, recent activity, and upcoming deadlines
- **FR-012**: System MUST display charts for: daily study time trends, completion rate over time, and task category distribution
- **FR-013**: System MUST provide an AI chat assistant that answers learning-related questions with streaming responses and maintains multi-turn conversation context within a session
- **FR-014**: System MUST allow the AI assistant to generate task breakdowns from natural language input within the chat
- **FR-015**: System MUST generate AI-powered daily summaries on demand, including completed tasks, study time, and next-day suggestions
- **FR-016**: System MUST provide context-aware AI recommendations based on user's learning history, progress, and saved learning preferences
- **FR-017**: System MUST support responsive design for mobile and desktop
- **FR-018**: System MUST support dark mode
- **FR-019**: System MUST handle AI service errors gracefully with automatic exponential backoff retry (1s/2s/4s intervals, max 3 attempts) and user-friendly error messages when all retries are exhausted
- **FR-020**: System MUST maintain learning plans as editable after generation — users can add, remove, reorder, and modify tasks
- **FR-021**: System MUST use infinite scrolling for task lists, loading tasks in batches as the user scrolls
- **FR-022**: System MUST allow tasks with unmet prerequisites to move into Doing, but MUST block transition to Done until all prerequisite tasks are completed, and MUST display a clear explanation of the unmet prerequisites

### Key Entities

- **User**: Represents a registered user. Key attributes: email, password (hashed), display name, avatar, preferences (theme, language, learning_style, weekly_study_hours, preferred_session_minutes), creation time.
- **Learning Goal**: Represents a user's learning objective. Key attributes: description, target duration, generated plan, status (Active/Paused/Completed/Archived), creation time. A user can have multiple learning goals simultaneously. Belongs to a User.
- **Task**: Represents a unit of work. Key attributes: title, description, status, priority, labels, deadline, estimated duration, recommended resources, creation time, update time. Can belong to a Learning Goal and/or a User. Supports hierarchical subtasks (parent-child) and dependency relationships (prerequisite tasks). Tasks may start before prerequisites are finished, but cannot be marked Done until all prerequisite tasks are completed.
- **Label**: Represents a category/tag for tasks. Key attributes: name, color. Many-to-many relationship with Tasks.
- **Study Session**: Represents a recorded learning activity. Key attributes: duration, associated task(s), date, notes. Belongs to a User.
- **AI Conversation**: Represents a chat session with the AI assistant. Key attributes: messages (role, content, timestamp), associated learning goal (optional), session context (maintains conversation history for multi-turn interactions). Belongs to a User.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can register and log in within 1 minute
- **SC-002**: AI generates a complete learning plan from natural language input within 30 seconds (streaming begins within 3 seconds)
- **SC-003**: Users can create a new task in under 15 seconds
- **SC-004**: Task search returns results in under 1 second for up to 1000 tasks
- **SC-005**: Dashboard loads with all statistics and charts within 2 seconds
- **SC-006**: AI chat responds to a question with streaming output beginning within 3 seconds
- **SC-007**: 90% of users can navigate the platform without needing a tutorial on first use
- **SC-008**: System supports 100 concurrent users without performance degradation
- **SC-009**: 95% of AI requests succeed on first attempt (with exponential backoff retry handling the remaining 5%)
- **SC-010**: Users report feeling more organized and productive in their learning after 1 week of use

## Clarifications

### Session 2026-06-17

- Q: 学习计划中的任务默认排序方式？ → A: 拓扑+灵活 — 按依赖关系排序，但独立任务可任意顺序完成
- Q: 任务列表的分页策略？ → A: 无限滚动 — 滚动到底部自动加载更多
- Q: AI 请求失败的重试策略？ → A: 指数退避重试 — 按 1s/2s/4s 间隔重试，最多 3 次
- Q: 用户是否可以同时拥有多个学习计划？ → A: 允许多个并行 — 用户可同时创建和管理多个学习计划
- Q: AI 聊天是否需要保持上下文记忆？ → A: 支持多轮上下文 — 同一会话内 AI 记住之前的对话内容
- Q: 任务依赖是否限制状态流转？ → A: 可开始不可完成 — 前置依赖未完成时，任务可以开始，但不能标记为完成
- Q: v1 是否包含通知设置？ → A: 不包含 — v1 仅保留昵称、头像和学习偏好，不提供通知设置

## Assumptions

- **Target users**: Individual self-learners (developers, students, professionals) who want to structure their learning journey. No team/collaboration features in v1.
- **Language support**: The platform UI and AI interactions support Chinese (Simplified) as the primary language, with English as secondary. The AI can respond in the language the user uses.
- **Learning preferences**: v1 stores UI preferences (`theme`, `language`) and recommendation preferences (`learning_style`, `weekly_study_hours`, `preferred_session_minutes`). AI-generated plans, chat suggestions, and daily summaries use the recommendation preferences when available.
- **AI service**: An OpenAI-compatible API endpoint is provided by the user or deployed separately. The platform connects to it via a configurable endpoint and API key. The specific model is not fixed.
- **Time tracking**: Users manually record study sessions rather than using an automatic timer. This is simpler to implement and gives users control over what counts as "study time."
- **Resource format**: Recommended resources are text-based descriptions (titles, URLs, brief descriptions). The system does not integrate with external learning platforms or course providers.
- **Data scope**: All data is per-user. No sharing, public profiles, or collaborative features in v1.
- **Plan flexibility**: AI-generated plans are starting points. Users have full control to modify any aspect after generation.
- **Notifications**: No push notifications or email reminders in v1. Upcoming deadlines are shown on the dashboard only.
- **Password recovery**: Deferred to post-v1. Email-based password reset flow is not included in v1 scope. Users who forget their password will need to contact support or re-register. This is excluded because v1 does not include email service integration.
