# Data Model: FlowTask

**Branch**: `001-learning-task-platform` | **Date**: 2026-06-17

## Entity Relationship Diagram

```text
User (1) ──── (*) Learning Goal
  │                    │
  │                    │ (1)
  │                    │
  │              (*) Task (*)
  │                    │
  │                    │ (self-ref parent)
  │                    │
  ├──── (*) Task       │
  ├──── (*) Label (*)──┘
  ├──── (*) Study Session
  └──── (*) AI Conversation
```

## Entities

### users

| Field        | Type         | Constraints            | Description     |
| ------------ | ------------ | ---------------------- | --------------- |
| id           | UUID         | PK, auto-gen           | 用户唯一标识    |
| email        | VARCHAR(255) | UNIQUE, NOT NULL       | 登录邮箱        |
| password_hash| VARCHAR(255) | NOT NULL               | 加密后的密码    |
| display_name | VARCHAR(100) | NOT NULL               | 显示名称        |
| avatar_url   | VARCHAR(500) | NULL                   | 头像链接        |
| preferences  | JSONB        | default '{}'           | 用户偏好设置（theme、language、learning_style、weekly_study_hours、preferred_session_minutes） |
| created_at   | TIMESTAMP    | NOT NULL, default now  | 创建时间        |
| updated_at   | TIMESTAMP    | NOT NULL, default now  | 更新时间        |

**Validation Rules**:
- email: 标准邮箱格式
- password: 最少 8 位，至少包含大小写字母和数字
- display_name: 2-100 字符
- preferences.theme: light/dark/system
- preferences.language: zh-CN/en
- preferences.learning_style: reading/video/project/mixed
- preferences.weekly_study_hours: 1-168
- preferences.preferred_session_minutes: 15-240

**Indexes**: email (unique)

---

### learning_goals

| Field          | Type         | Constraints          | Description       |
| -------------- | ------------ | -------------------- | ----------------- |
| id             | UUID         | PK, auto-gen         | 学习目标唯一标识  |
| user_id        | UUID         | FK → users.id, NOT NULL | 所属用户     |
| description    | TEXT         | NOT NULL             | 学习目标描述      |
| target_duration| VARCHAR(50)  | NULL                 | 目标时长（如"2个月"）|
| status         | VARCHAR(20)  | NOT NULL, default 'active' | 状态枚举   |
| ai_prompt      | TEXT         | NULL                 | 原始 AI 提示词    |
| created_at     | TIMESTAMP    | NOT NULL, default now| 创建时间          |
| updated_at     | TIMESTAMP    | NOT NULL, default now| 更新时间          |

**Status Transitions**:
```
active → paused    (用户暂停)
active → completed (用户标记完成)
active → archived  (用户归档)
paused → active    (恢复)
paused → archived  (归档)
completed → archived (归档)
```

**Validation Rules**:
- description: 1-2000 字符
- status: 仅允许 active/paused/completed/archived

**Indexes**: user_id, status, (user_id, status)

---

### tasks

| Field              | Type         | Constraints          | Description         |
| ------------------ | ------------ | -------------------- | ------------------- |
| id                 | UUID         | PK, auto-gen         | 任务唯一标识        |
| user_id            | UUID         | FK → users.id, NOT NULL | 所属用户        |
| learning_goal_id   | UUID         | FK → learning_goals.id, NULL | 所属学习目标（手动创建可为 NULL）|
| parent_task_id     | UUID         | FK → tasks.id, NULL  | 父任务（支持子任务）|
| title              | VARCHAR(200) | NOT NULL             | 任务标题            |
| description        | TEXT         | NULL                 | 任务描述            |
| status             | VARCHAR(20)  | NOT NULL, default 'todo' | 状态枚举       |
| priority           | VARCHAR(20)  | NOT NULL, default 'medium' | 优先级枚举   |
| deadline           | TIMESTAMP    | NULL                 | 截止时间            |
| estimated_duration | VARCHAR(50)  | NULL                 | 预计耗时            |
| recommended_resources| JSONB      | default '[]'         | 推荐资料列表        |
| sort_order         | INTEGER      | NOT NULL, default 0  | 排序序号            |
| created_at         | TIMESTAMP    | NOT NULL, default now| 创建时间            |
| updated_at         | TIMESTAMP    | NOT NULL, default now| 更新时间            |

**Status Values**: todo, doing, done

**Priority Values**: low, medium, high, urgent

**Validation Rules**:
- title: 1-200 字符
- status: 仅允许 todo/doing/done
- priority: 仅允许 low/medium/high/urgent
- 若存在未完成前置任务，任务可进入 doing，但不可进入 done
- 列表筛选需支持 `deadline_from` 和 `deadline_to` 作为截止日期范围边界

**Indexes**: user_id, learning_goal_id, parent_task_id, status, priority, deadline, (user_id, status), (user_id, learning_goal_id)

---

### task_dependencies

| Field              | Type  | Constraints          | Description       |
| ------------------ | ----- | -------------------- | ----------------- |
| task_id            | UUID  | FK → tasks.id, PK   | 依赖方任务        |
| depends_on_task_id | UUID  | FK → tasks.id, PK   | 被依赖的任务      |

**说明**: 表示任务之间的依赖关系。`task_id` 依赖 `depends_on_task_id`，
用于拓扑排序、完成态校验，以及阻止循环依赖。

**Indexes**: task_id, depends_on_task_id

---

### labels

| Field      | Type         | Constraints          | Description   |
| ---------- | ------------ | -------------------- | ------------- |
| id         | UUID         | PK, auto-gen         | 标签唯一标识  |
| user_id    | UUID         | FK → users.id, NOT NULL | 所属用户  |
| name       | VARCHAR(50)  | NOT NULL             | 标签名称      |
| color      | VARCHAR(7)   | NOT NULL, default '#6366f1' | 标签颜色（HEX）|

**Indexes**: (user_id, name) unique

---

### task_labels

| Field    | Type | Constraints        | Description |
| -------- | ---- | ------------------ | ----------- |
| task_id  | UUID | FK → tasks.id, PK | 任务        |
| label_id | UUID | FK → labels.id, PK| 标签        |

---

### study_sessions

| Field     | Type         | Constraints          | Description     |
| --------- | ------------ | -------------------- | --------------- |
| id        | UUID         | PK, auto-gen         | 学习记录唯一标识|
| user_id   | UUID         | FK → users.id, NOT NULL | 所属用户    |
| task_id   | UUID         | FK → tasks.id, NULL  | 关联任务        |
| duration  | INTEGER      | NOT NULL             | 时长（分钟）    |
| date      | DATE         | NOT NULL             | 学习日期        |
| notes     | TEXT         | NULL                 | 学习笔记        |
| created_at| TIMESTAMP    | NOT NULL, default now| 创建时间        |

**Indexes**: user_id, (user_id, date), task_id

---

### ai_conversations

| Field           | Type         | Constraints          | Description     |
| --------------- | ------------ | -------------------- | --------------- |
| id              | UUID         | PK, auto-gen         | 会话唯一标识    |
| user_id         | UUID         | FK → users.id, NOT NULL | 所属用户    |
| learning_goal_id| UUID         | FK → learning_goals.id, NULL | 关联学习目标 |
| title           | VARCHAR(200) | NULL                 | 会话标题        |
| created_at      | TIMESTAMP    | NOT NULL, default now| 创建时间        |
| updated_at      | TIMESTAMP    | NOT NULL, default now| 更新时间        |

**Indexes**: user_id, learning_goal_id

---

### ai_messages

| Field           | Type         | Constraints          | Description     |
| --------------- | ------------ | -------------------- | --------------- |
| id              | UUID         | PK, auto-gen         | 消息唯一标识    |
| conversation_id | UUID         | FK → ai_conversations.id, NOT NULL | 所属会话 |
| role            | VARCHAR(20)  | NOT NULL             | 角色枚举        |
| content         | TEXT         | NOT NULL             | 消息内容        |
| created_at      | TIMESTAMP    | NOT NULL, default now| 创建时间        |

**Role Values**: user, assistant, system

**Indexes**: (conversation_id, created_at)
