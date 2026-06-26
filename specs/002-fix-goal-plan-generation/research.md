# Research: 修复学习目标生成学习计划功能

**Branch**: `002-fix-goal-plan-generation` | **Date**: 2026-06-23

## 1. AI 响应解析容错策略

**Decision**: 实现多层解析策略，支持纯 JSON、markdown 代码块、带解释文本的 JSON

**Rationale**: 
- AI 模型（如 GPT-4）经常在 JSON 前后添加解释性文字或使用 markdown 代码块
- 错误 `invalid character '`' looking for beginning of value` 表明后端期望纯 JSON 但收到了 markdown 格式
- 需要健壮的解析逻辑以提高成功率（目标 SC-003: <1% 错误率）

**Alternatives Considered**:
1. **严格 JSON 解析** - 拒绝非纯 JSON 响应
   - 优点：简单
   - 缺点：解析错误率高，用户体验差
2. **正则表达式提取** - 使用正则匹配 JSON 块
   - 优点：灵活性高
   - 缺点：正则复杂，边界情况多
3. **多层解析策略** - 先尝试纯 JSON，失败后提取 markdown 代码块，最后尝试查找 JSON 对象
   - 优点：容错性强，覆盖各种 AI 输出格式
   - 缺点：实现稍复杂

**Recommended**: 方案 3 - 多层解析策略

**Implementation Approach**:
```go
func parseAIResponse(raw string) ([]Task, error) {
    // 1. 尝试直接 JSON 解析
    var tasks []Task
    if err := json.Unmarshal([]byte(raw), &tasks); err == nil {
        return tasks, nil
    }

    // 2. 提取 markdown 代码块中的 JSON
    jsonBlock := extractMarkdownCodeBlock(raw)
    if jsonBlock != "" {
        if err := json.Unmarshal([]byte(jsonBlock), &tasks); err == nil {
            return tasks, nil
        }
    }

    // 3. 查找 JSON 对象/数组
    jsonContent := findJSONArray(raw)
    if jsonContent != "" {
        if err := json.Unmarshal([]byte(jsonContent), &tasks); err == nil {
            return tasks, nil
        }
    }

    return nil, fmt.Errorf("unable to parse AI response: %w", err)
}

func extractMarkdownCodeBlock(content string) string {
    // 匹配 ```json ... ``` 或 ``` ... ```
    re := regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")
    matches := re.FindStringSubmatch(content)
    if len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }
    return ""
}

func findJSONArray(content string) string {
    // 查找第一个 [ 和最后一个 ] 之间的内容
    start := strings.Index(content, "[")
    end := strings.LastIndex(content, "]")
    if start != -1 && end > start {
        return content[start : end+1]
    }
    return ""
}
```

**References**:
- OpenAI API best practices for structured output
- Go JSON parsing with fallback strategies

---

## 2. SSE (Server-Sent Events) 流式传输最佳实践

**Decision**: 使用标准 SSE 协议，每个任务作为独立事件发送

**Rationale**:
- SSE 是浏览器原生支持的流式协议，无需额外库
- 适合单向服务端推送（AI 生成任务）
- 比 WebSocket 更简单，资源消耗更低

**Alternatives Considered**:
1. **WebSocket** - 双向通信
   - 优点：实时性更好
   - 缺点：实现复杂，对于单向推送过度设计
2. **长轮询** - 客户端定期请求
   - 优点：兼容性好
   - 缺点：延迟高，资源浪费
3. **SSE** - 服务端推送事件
   - 优点：简单、高效、浏览器原生支持
   - 缺点：单向通信（但本场景只需单向）

**Recommended**: 方案 3 - SSE

**Implementation Approach**:
```go
// 后端 SSE 端点
func (h *LearningGoalHandler) GeneratePlan(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 调用 AI 服务生成任务
    taskChan := make(chan Task)
    go h.service.GeneratePlanStream(goalID, taskChan)

    for task := range taskChan {
        // 发送 SSE 事件
        c.SSEvent("task", task)
        c.Writer.Flush()
    }

    // 发送完成事件
    c.SSEvent("done", map[string]interface{}{
        "learning_goal_id": goalID,
        "task_count":       taskCount,
    })
}
```

**Frontend SSE Client**:
```typescript
const eventSource = new EventSource(`/api/learning-goals/${goalId}/generate-plan`, {
    headers: { Authorization: `Bearer ${token}` }
});

eventSource.addEventListener('task', (e) => {
    const task = JSON.parse(e.data);
    addTask(task);
});

eventSource.addEventListener('done', (e) => {
    const data = JSON.parse(e.data);
    setPhase('done');
    eventSource.close();
});

eventSource.onerror = (e) => {
    setPhase('error');
    eventSource.close();
};
```

**References**:
- MDN Server-Sent Events API
- Go net/http SSE implementation patterns

---

## 3. 前端流式状态管理策略

**Decision**: 使用 Zustand store + custom hook 管理生成状态

**Rationale**:
- 状态复杂度高（phase、tasks、error、进度）
- 需要在多个组件间共享状态
- Zustand 已在项目中使用，保持一致性

**Alternatives Considered**:
1. **useState + props drilling** - 组件内状态
   - 优点：简单
   - 缺点：状态难以共享，代码臃肿
2. **React Context** - 全局状态
   - 优点：无需额外库
   - 缺点：性能问题，重渲染多
3. **Zustand store** - 状态管理库
   - 优点：性能好，易用，已有基础设施
   - 缺点：需要学习

**Recommended**: 方案 3 - Zustand store

**Implementation Approach**:
```typescript
// stores/goal-store.ts
interface GoalState {
  // 生成状态
  generationPhase: 'idle' | 'connecting' | 'streaming' | 'preview' | 'done' | 'error';
  generatedTasks: Task[];
  errorMessage: string;
  taskCount: number;

  // Actions
  setPhase: (phase: GoalState['generationPhase']) => void;
  addTask: (task: Task) => void;
  setError: (error: string) => void;
  reset: () => void;
  confirmSave: () => Promise<void>;
}

export const useGoalStore = create<GoalState>((set, get) => ({
  generationPhase: 'idle',
  generatedTasks: [],
  errorMessage: '',
  taskCount: 0,

  setPhase: (phase) => set({ generationPhase: phase }),
  addTask: (task) => set((state) => ({
    generatedTasks: [...state.generatedTasks, task],
    taskCount: state.taskCount + 1,
  })),
  setError: (error) => set({ errorMessage: error, generationPhase: 'error' }),
  reset: () => set({
    generationPhase: 'idle',
    generatedTasks: [],
    errorMessage: '',
    taskCount: 0,
  }),
  confirmSave: async () => {
    const { generatedTasks } = get();
    await api.post(`/learning-goals/${goalId}/tasks`, { tasks: generatedTasks });
    set({ generationPhase: 'done' });
  },
}));
```

**Custom Hook**:
```typescript
// hooks/use-goal-stream.ts
export function useGoalStream(goalId: string) {
  const { setPhase, addTask, setError, reset } = useGoalStore();

  const startStream = useCallback(async () => {
    setPhase('connecting');
    try {
      const eventSource = new EventSource(
        `${API_URL}/learning-goals/${goalId}/generate-plan`,
        { headers: { Authorization: `Bearer ${token}` } }
      );

      eventSource.addEventListener('task', (e) => {
        const task = JSON.parse(e.data);
        addTask(task);
        setPhase('streaming');
      });

      eventSource.addEventListener('done', () => {
        setPhase('preview');
        eventSource.close();
      });

      eventSource.onerror = () => {
        setError('连接中断，请重试');
        eventSource.close();
      };
    } catch (err) {
      setError(err.message);
    }
  }, [goalId]);

  return { startStream, reset };
}
```

**References**:
- Zustand documentation
- React hooks best practices

---

## 4. 网络中断恢复策略

**Decision**: 使用 localStorage 临时保存生成进度，提供继续/重新生成选项

**Rationale**:
- 用户可能在生成过程中离开页面
- 需要保留已生成内容，避免用户重复等待
- 符合澄清决策："显示已生成部分 + 提供'继续生成'按钮"

**Implementation Approach**:
```typescript
// 保存到 localStorage
const saveProgress = (goalId: string, tasks: Task[]) => {
  localStorage.setItem(`goal-progress-${goalId}`, JSON.stringify({
    tasks,
    timestamp: Date.now(),
  }));
};

// 恢复进度
const loadProgress = (goalId: string): Task[] | null => {
  const saved = localStorage.getItem(`goal-progress-${goalId}`);
  if (!saved) return null;

  const { tasks, timestamp } = JSON.parse(saved);
  // 检查是否过期（24小时）
  if (Date.now() - timestamp > 24 * 60 * 60 * 1000) {
    localStorage.removeItem(`goal-progress-${goalId}`);
    return null;
  }
  return tasks;
};

// 页面加载时恢复
useEffect(() => {
  const savedTasks = loadProgress(goalId);
  if (savedTasks && savedTasks.length > 0) {
    setGeneratedTasks(savedTasks);
    setPhase('preview');
  }
}, [goalId]);
```

**References**:
- Web Storage API best practices

---

## 5. 错误处理和重试策略

**Decision**: 实现指数退避重试，最多 3 次，提供友好错误信息

**Rationale**:
- AI 服务可能暂时不可用
- 网络波动可能导致请求失败
- 用户不应因暂时性错误而丢失进度

**Implementation Approach**:
```go
// 后端重试逻辑
func (s *LearningGoalService) GeneratePlanWithRetry(goalID string, maxRetries int) ([]Task, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        tasks, err := s.GeneratePlan(goalID)
        if err == nil {
            return tasks, nil
        }
        lastErr = err

        // 指数退避：1s, 2s, 4s
        backoff := time.Duration(1<<uint(i)) * time.Second
        time.Sleep(backoff)
    }
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

**Frontend Error Display**:
```typescript
const errorMessages: Record<string, string> = {
  'AI_SERVICE_UNAVAILABLE': 'AI 服务暂时不可用，请稍后重试',
  'INVALID_RESPONSE': 'AI 返回了无效的响应，已自动重试',
  'NETWORK_ERROR': '网络连接中断，请检查网络后重试',
  'TIMEOUT': '请求超时，请重试',
};

const getErrorMessage = (error: string): string => {
  return errorMessages[error] || '生成失败，请重试';
};
```

**References**:
- Exponential backoff best practices
- User-friendly error messages guidelines
