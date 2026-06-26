# API Contracts: FlowTask

**Branch**: `001-learning-task-platform` | **Date**: 2026-06-17

## 通用约定

### 响应格式

```json
// 成功响应
{
  "code": 0,
  "data": { ... },
  "message": "success"
}

// 错误响应
{
  "code": 40001,
  "data": null,
  "message": "Invalid email format"
}
```

### 错误码规范

| 范围     | 含义         |
| -------- | ------------ |
| 0        | 成功         |
| 400xx    | 客户端错误   |
| 401xx    | 认证相关     |
| 403xx    | 权限不足     |
| 409xx    | 资源状态冲突 |
| 404xx    | 资源不存在   |
| 500xx    | 服务端错误   |

### 认证方式

除 `/auth/*` 外所有接口需要在 Header 中携带:
```
Authorization: Bearer <access_token>
```

Token 过期时返回 `401`，前端自动用 Refresh Token 刷新。

**Refresh Token 传输方式**:
- Login/Register 响应中返回 `refresh_token` 字段，前端存储到内存 (Zustand store)
- Refresh 请求通过 JSON body 传递 `refresh_token`
- Logout 请求通过 JSON body 传递当前 `refresh_token`（用于服务端吊销）
- Access Token 仅通过 Authorization Header 传输，不存 Cookie

---

## Auth API

### POST /api/auth/register

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "display_name": "张三"
}
```

**Response (201)**:
```json
{
  "code": 0,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "display_name": "张三",
      "avatar_url": null,
      "created_at": "2026-06-17T10:00:00Z"
    },
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

**Validation**:
- email: 有效邮箱格式
- password: 8-128 位，至少含大小写字母和数字
- display_name: 2-100 字符

---

### POST /api/auth/login

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "user": { ... },
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

---

### POST /api/auth/refresh

**Request**:
```json
{
  "refresh_token": "eyJ..."
}
```

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

---

### POST /api/auth/logout

**Request**: (empty body, uses current refresh token)

**Response (200)**:
```json
{
  "code": 0,
  "data": null,
  "message": "Logged out successfully"
}
```

---

## User API

### GET /api/user/profile

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "张三",
    "avatar_url": null,
    "preferences": {
      "theme": "dark",
      "language": "zh-CN",
      "learning_style": "project",
      "weekly_study_hours": 10,
      "preferred_session_minutes": 45
    },
    "created_at": "2026-06-17T10:00:00Z"
  }
}
```

### PUT /api/user/profile

**Request**:
```json
{
  "display_name": "新名字",
  "avatar_url": "https://...",
  "preferences": {
    "theme": "dark",
    "language": "zh-CN",
    "learning_style": "project",
    "weekly_study_hours": 10,
    "preferred_session_minutes": 45
  }
}
```

**Response (200)**: 同 GET profile

**Validation**:
- `preferences.theme`: `light` / `dark` / `system`
- `preferences.language`: `zh-CN` / `en`
- `preferences.learning_style`: `reading` / `video` / `project` / `mixed`
- `preferences.weekly_study_hours`: 1-168
- `preferences.preferred_session_minutes`: 15-240

---

## Learning Goals API

### POST /api/learning-goals

**Request**:
```json
{
  "description": "我想两个月学会 RAG",
  "target_duration": "2个月"
}
```

**Response (201)**: SSE 流式返回 AI 生成的学习计划

**Note**: 服务端在生成学习计划时自动读取用户已保存的学习偏好，并将其作为 AI 上下文的一部分。

**SSE Events**:
```
event: task
data: {"id":"uuid","title":"学习 Embedding","description":"...","estimated_duration":"1周","recommended_resources":[...],"parent_task_id":null,"sort_order":1}

event: task
data: {"id":"uuid","title":"学习 Vector Database","description":"...","estimated_duration":"1周","recommended_resources":[...],"parent_task_id":null,"sort_order":2}

event: task
data: {"id":"uuid","title":"掌握文本向量化方法","description":"...","estimated_duration":"3天","recommended_resources":[...],"parent_task_id":"<embedding_task_id>","sort_order":1}

event: done
data: {"learning_goal_id":"uuid","task_count":12}
```

---

### GET /api/learning-goals

**Query Parameters**:
- `status`: 筛选状态 (active/paused/completed/archived)
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": "uuid",
        "description": "我想两个月学会 RAG",
        "target_duration": "2个月",
        "status": "active",
        "task_count": 12,
        "completed_task_count": 3,
        "created_at": "2026-06-17T10:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

---

### PUT /api/learning-goals/:id

**Request**:
```json
{
  "status": "paused",
  "description": "更新后的描述"
}
```

**Response (200)**: 返回更新后的 learning goal

---

### PUT /api/learning-goals/:id/tasks/reorder

**Request**:
```json
{
  "items": [
    {"task_id": "uuid-1", "parent_task_id": null, "sort_order": 1},
    {"task_id": "uuid-2", "parent_task_id": "uuid-1", "sort_order": 1}
  ]
}
```

**Response (200)**: 返回重排后的任务树

**Validation**:
- 不允许形成循环父子关系
- 不允许破坏既有依赖关系
- 若校验失败，返回 `409 Conflict`

---

### POST /api/learning-goals/:id/tasks

**Request**:
```json
{
  "title": "补充学习任务",
  "description": "手动添加的任务",
  "parent_task_id": null,
  "sort_order": 3
}
```

**Response (201)**: 返回新建任务

---

### DELETE /api/learning-goals/:id/tasks/:taskId

**Response (200)**: 删除任务，并在需要时返回更新后的任务树摘要

---

## Tasks API

### POST /api/tasks

**Request**:
```json
{
  "title": "学习 LangChain",
  "description": "深入学习 LangChain 框架",
  "priority": "high",
  "deadline": "2026-07-17T23:59:59Z",
  "learning_goal_id": "uuid",
  "parent_task_id": null,
  "label_ids": ["uuid1", "uuid2"]
}
```

**Response (201)**: 返回创建的 task 完整对象

---

### GET /api/tasks

**Query Parameters**:
- `status`: 筛选状态 (todo/doing/done)，支持多选（逗号分隔）
- `priority`: 筛选优先级 (low/medium/high/urgent)，支持多选
- `label_ids`: 筛选标签（逗号分隔 UUID）
- `learning_goal_id`: 按学习目标筛选
- `search`: 关键词搜索（搜索 title 和 description）
- `deadline_from`: 截止日期起始边界 (ISO 8601)
- `deadline_to`: 截止日期结束边界 (ISO 8601)
- `sort_by`: 排序字段 (created_at/deadline/priority/status)
- `sort_order`: 排序方向 (asc/desc)，默认 desc
- `cursor`: 分页游标（无限滚动用）
- `limit`: 每页数量，默认 20，最大 100

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": "uuid",
        "title": "学习 LangChain",
        "description": "...",
        "status": "todo",
        "priority": "high",
        "deadline": "2026-07-17T23:59:59Z",
        "estimated_duration": "2周",
        "recommended_resources": [],
        "labels": [
          {"id": "uuid", "name": "AI", "color": "#6366f1"}
        ],
        "subtask_count": 3,
        "completed_subtask_count": 1,
        "learning_goal_id": "uuid",
        "parent_task_id": null,
        "dependencies": ["uuid_dep1"],
        "created_at": "2026-06-17T10:00:00Z",
        "updated_at": "2026-06-17T10:00:00Z"
      }
    ],
    "next_cursor": "eyJpZCI6...",
    "has_more": true
  }
}
```

---

### GET /api/tasks/:id

**Response (200)**: 返回 task 完整对象（含 subtasks 列表和 dependencies）

---

### PUT /api/tasks/:id

**Request**:
```json
{
  "title": "更新后的标题",
  "status": "doing",
  "priority": "urgent"
}
```

**Response (200)**: 返回更新后的 task

**Validation**:
- 当 `status` 更新为 `done` 时，若存在未完成前置任务，返回 `409 Conflict`
- 返回体应包含未完成前置任务列表与可读错误信息

---

### DELETE /api/tasks/:id

**Query Parameters**:
- `delete_subtasks`: 是否同时删除子任务，默认 true

**Response (200)**:
```json
{
  "code": 0,
  "data": null,
  "message": "Task deleted"
}
```

---

### POST /api/tasks/:id/dependencies

**Request**:
```json
{
  "depends_on_task_id": "uuid"
}
```

**Response (201)**: 返回创建的依赖关系

**Validation**:
- 拒绝自依赖
- 拒绝循环依赖
- 若创建后会形成循环，返回 `409 Conflict`

---

## Labels API

### POST /api/labels

**Request**:
```json
{
  "name": "AI",
  "color": "#6366f1"
}
```

**Response (201)**: 返回创建的 label

---

### GET /api/labels

**Response (200)**: 返回用户所有 labels 列表

---

### PUT /api/labels/:id

**Request**:
```json
{
  "name": "AI 重点",
  "color": "#8b5cf6"
}
```

**Response (200)**: 返回更新后的 label

---

### DELETE /api/labels/:id

**Response (200)**: 删除标签

---

## Study Sessions API

### POST /api/study-sessions

**Request**:
```json
{
  "task_id": "uuid",
  "duration": 45,
  "date": "2026-06-17",
  "notes": "学习了 RAG 基础概念"
}
```

**Response (201)**: 返回创建的学习记录

---

### GET /api/study-sessions

**Query Parameters**:
- `start_date`: 起始日期
- `end_date`: 结束日期
- `task_id`: 按任务筛选

**Response (200)**: 返回学习记录列表

---

## Dashboard API

### GET /api/dashboard/stats

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "today_tasks": {
      "total": 5,
      "completed": 2,
      "items": [...]
    },
    "overall": {
      "total_tasks": 120,
      "completed_tasks": 45,
      "completion_rate": 0.375
    },
    "study_time": {
      "today_minutes": 90,
      "week_minutes": 540,
      "month_minutes": 2100
    },
    "upcoming_deadlines": [...],
    "recent_activity": [...]
  }
}
```

---

### GET /api/dashboard/charts/study-time

**Query Parameters**:
- `period`: 时间范围 (week/month)，默认 week

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "labels": ["06-11", "06-12", "06-13", ...],
    "values": [60, 90, 45, 120, 80, 0, 95]
  }
}
```

---

### GET /api/dashboard/charts/category-stats

**Response (200)**:
```json
{
  "code": 0,
  "data": [
    {"label": "AI", "count": 15, "completed": 8},
    {"label": "Backend", "count": 10, "completed": 5}
  ]
}
```

---

### GET /api/dashboard/charts/completion-rate

**Query Parameters**:
- `period`: 时间范围 (week/month)，默认 week

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "labels": ["06-11", "06-12", "06-13", ...],
    "values": [0.2, 0.35, 0.5, 0.45, 0.6, 0.55, 0.7]
  }
}
```

---

## AI Chat API

### POST /api/ai/chat

**Request**:
```json
{
  "conversation_id": "uuid",
  "learning_goal_id": "uuid",
  "message": "Docker 怎么学习？"
}
```

**Response (200)**: SSE 流式返回

**SSE Events**:
```
event: conversation
data: {"conversation_id":"uuid"}

event: delta
data: {"content":"Docker 是一个"}

event: delta
data: {"content":"容器化平台..."}

event: done
data: {"full_content":"Docker 是一个容器化平台..."}
```

**Note**: 聊天流程可根据前端模式生成普通回答或任务拆解结果；拆解模式
返回的内容应可被保存为任务树。涉及学习建议、任务拆解或下一步推荐时，
服务端自动读取用户已保存的学习偏好作为上下文。

---

### GET /api/ai/conversations

**Response (200)**: 返回用户的对话列表

---

### GET /api/ai/conversations/:id/messages

**Response (200)**: 返回对话的消息历史

---

### POST /api/ai/task-breakdown

**Request**:
```json
{
  "learning_goal_id": "uuid",
  "message": "帮我拆解 Kubernetes 学习计划",
  "save_to_goal": true
}
```

**Response (200)**: SSE 流式返回拆解后的任务树；当 `save_to_goal=true`
时，完成后返回保存结果和新增任务 ID 列表

---

### POST /api/ai/daily-summary

**Request**:
```json
{
  "date": "2026-06-17"
}
```

**Response (200)**: SSE 流式返回 AI 生成的每日总结

**Note**: 明日建议部分应结合用户学习偏好生成。

---

## Health API

### GET /api/health

**Response (200)**:
```json
{
  "code": 0,
  "data": {
    "status": "ok",
    "database": "ok",
    "redis": "ok",
    "ai_service": "ok"
  }
}
```

**AI 服务降级说明**: `ai_service` 字段值为 `ok` | `degraded` | `unavailable`。
当 AI 服务不可用时，整体 `status` 仍可为 `ok`（AI 非核心依赖），但
AI 相关功能（学习计划生成、聊天、总结）将返回友好降级提示。
