# API Contracts: 修复学习目标生成学习计划功能

**Branch**: `002-fix-goal-plan-generation` | **Date**: 2026-06-23

## Learning Goals API (Updated)

### POST /api/learning-goals

**Request**:
```json
{
  "description": "我想两个月学会 RAG",
  "target_duration": "2个月"
}
```

**Response (200)**: 返回生成会话 ID，用于后续流式获取和确认保存

```json
{
  "code": 0,
  "data": {
    "session_id": "uuid",
    "learning_goal_id": "uuid",
    "status": "generating"
  },
  "message": "success"
}
```

**Changes**: 
- 原来直接返回 SSE 流，现在先返回 session_id
- 前端使用 session_id 建立 SSE 连接
- 支持中断恢复（使用 session_id 重新连接）

---

### GET /api/learning-goals/:id/generate-stream

**Query Parameters**:
- `session_id`: 生成会话 ID（必填）

**Response (200)**: SSE 流式返回 AI 生成的学习计划

**SSE Events**:
```
event: task
data: {"id":"uuid","title":"学习 Embedding","description":"...","estimated_duration":"1周","recommended_resources":[...],"parent_task_id":null,"sort_order":1}

event: task
data: {"id":"uuid","title":"学习 Vector Database","description":"...","estimated_duration":"1周","recommended_resources":[...],"parent_task_id":null,"sort_order":2}

event: task
data: {"id":"uuid","title":"掌握文本向量化方法","description":"...","estimated_duration":"3天","recommended_resources":[...],"parent_task_id":"<embedding_task_id>","sort_order":1}

event: progress
data: {"task_count": 3}

event: done
data: {"learning_goal_id":"uuid","task_count":12}
```

**Changes**:
- 新增 `progress` 事件，实时返回已生成任务数
- 使用 `session_id` 支持中断恢复
- 如果连接断开，前端可使用相同 `session_id` 重新连接，继续获取任务

**Error Events**:
```
event: error
data: {"code":"AI_SERVICE_UNAVAILABLE","message":"AI 服务暂时不可用，请稍后重试"}
```

---

### POST /api/learning-goals/:id/tasks/confirm

**Request**:
```json
{
  "session_id": "uuid",
  "tasks": [
    {
      "id": "uuid",
      "title": "学习 Embedding",
      "description": "深入理解 Embedding 的原理和应用",
      "estimated_duration": "1周",
      "recommended_resources": ["https://example.com/embedding"],
      "parent_task_id": null,
      "sort_order": 1
    }
  ]
}
```

**Response (200)**: 确认保存成功

```json
{
  "code": 0,
  "data": {
    "learning_goal_id": "uuid",
    "saved_task_count": 12,
    "message": "学习计划已保存"
  },
  "message": "success"
}
```

**Validation**:
- `session_id` 必须有效且未过期（24小时有效期）
- `tasks` 列表不能为空
- 每个 task 必须包含 `id`, `title`, `sort_order`
- 保存后自动清除 session

**Error Responses**:
```json
// Session 过期
{
  "code": 40001,
  "data": null,
  "message": "生成会话已过期，请重新生成"
}

// 无效任务数据
{
  "code": 40002,
  "data": null,
  "message": "任务数据格式无效"
}
```

---

### POST /api/learning-goals/:id/regenerate

**Request**: 无需请求体，使用现有 learning goal 的描述重新生成

**Response (200)**: 返回新的生成会话 ID

```json
{
  "code": 0,
  "data": {
    "session_id": "new-uuid",
    "learning_goal_id": "uuid",
    "status": "generating"
  },
  "message": "success"
}
```

**Changes**:
- 新增重新生成端点
- 清除旧的 session 和任务
- 返回新的 session_id

---

## 错误码规范（本功能新增）

| 错误码 | 含义             | 用户提示                           | 触发场景                                                                                      |
| ------ | ---------------- | ---------------------------------- | --------------------------------------------------------------------------------------------- |
| 40001  | Session 过期     | 生成会话已过期，请重新生成         | 用户尝试确认保存或继续生成时，session 已超过 24 小时有效期                                     |
| 40002  | 无效任务数据     | 任务数据格式无效                   | 确认保存时，tasks 列表为空或缺少必填字段（id, title, sort_order）                              |
| 40003  | AI 响应解析失败  | AI 返回了无效的响应，已自动重试    | AI 返回的响应无法解析为有效的 JSON（经过多层解析尝试后仍然失败）                               |
| 50001  | AI 服务不可用    | AI 服务暂时不可用，请稍后重试      | AI 服务返回 5xx 错误或连接超时（经过 3 次重试后仍然失败）                                      |
| 50002  | 生成超时         | 生成超时，请重试                   | AI 服务响应时间超过 60 秒                                                                     |

**测试覆盖要求** (Phase 8):
- T061: 验证所有错误码在正确场景下返回
- T062: 验证所有错误场景显示用户友好的前端提示信息

---

## 与现有 API 的关系

**保持不变**:
- `GET /api/learning-goals` - 获取学习目标列表
- `PUT /api/learning-goals/:id` - 更新学习目标状态
- `DELETE /api/learning-goals/:id` - 删除学习目标
- `PUT /api/learning-goals/:id/tasks/reorder` - 任务重排

**修改**:
- `POST /api/learning-goals` - 从直接返回 SSE 流改为返回 session_id
- 新增 `GET /api/learning-goals/:id/generate-stream` - SSE 流式端点
- 新增 `POST /api/learning-goals/:id/tasks/confirm` - 确认保存端点
- 新增 `POST /api/learning-goals/:id/regenerate` - 重新生成端点

**向后兼容性**:
- 如果前端未传递 `session_id`，后端可降级为旧版行为（直接返回 SSE 流）
- 建议设置 feature flag 控制新旧版本切换
