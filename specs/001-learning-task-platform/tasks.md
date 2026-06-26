# Tasks: FlowTask - AI Learning Plan & Task Management Platform

**Input**: Design documents from `/specs/001-learning-task-platform/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 项目初始化、Monorepo 结构、开发环境

- [x] T001 创建 Monorepo 项目结构 (web/, server/, docker-compose.yml, Makefile) 按 plan.md 定义的目录结构
- [x] T002 初始化 Go 后端项目在 server/ 目录，配置 go.mod，安装 Gin, GORM, golang-jwt, go-redis 等依赖
- [x] T003 [P] 初始化 Next.js 前端项目在 web/ 目录，配置 TypeScript 严格模式、Tailwind CSS、shadcn/ui
- [x] T004 [P] 创建 docker-compose.yml 配置 PostgreSQL 15 和 Redis 7 容器
- [x] T005 [P] 创建 Makefile 包含 dev, dev-server, dev-web, test, build 等常用命令
- [x] T006 [P] 配置前端 ESLint、Prettier，后端 golangci-lint

**Checkpoint**: 项目骨架就绪，可通过 `make dev` 启动开发环境

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 所有用户故事的核心依赖基础设施

**⚠️ CRITICAL**: 必须完成此阶段后才能开始任何用户故事

### 后端基础

- [x] T007 实现配置管理模块 server/internal/config/config.go 支持 YAML 配置文件读取 (数据库、Redis、JWT、AI API 配置)
- [x] T008 实现统一响应格式 server/internal/response/response.go 定义成功/错误响应结构和错误码规范
- [x] T009 实现数据库连接和 GORM 初始化 server/internal/config/database.go，支持 PostgreSQL 连接池
- [x] T010 实现 Redis 连接初始化 server/internal/config/redis.go
- [x] T011 实现数据库迁移框架 server/cmd/migrate/main.go 和 server/migrations/ 目录结构
- [x] T012 实现 User 模型 server/internal/model/user.go 按 data-model.md 定义字段和约束
- [x] T013 实现 JWT 认证中间件 server/internal/middleware/auth.go，支持 Access Token 校验和自动解析用户信息
- [x] T014 实现 CORS 中间件 server/internal/middleware/cors.go 支持前后端分离跨域
- [x] T015 [P] 实现日志中间件 server/internal/middleware/logger.go 记录请求日志
- [x] T016 实现 Gin 路由初始化 server/cmd/server/main.go 配置路由分组和中间件挂载
- [x] T017 实现 Health API server/internal/handler/health.go 提供 GET /api/health 端点检查数据库、Redis 连接状态和 AI 服务可用性（AI 为降级项，不可用时返回 degraded 而非阻塞）

### 前端基础

- [x] T018 实现 API 客户端 web/src/lib/api.ts 封装 Axios 实例，配置请求/响应拦截器，实现自动 Token 刷新逻辑
- [x] T019 实现认证状态管理 web/src/stores/auth-store.ts 使用 Zustand 管理用户登录状态、Token 存储
- [x] T020 [P] 实现 shadcn/ui 主题配置 web/src/app/globals.css 支持亮色/暗色模式切换
- [x] T021 [P] 实现根布局 web/src/app/layout.tsx 配置字体、主题 Provider、全局样式
- [x] T022 [P] 实现 TypeScript 类型定义 web/src/types/api.ts 对齐 contracts/api.md 中的响应格式
- [x] T023 实现通用工具函数 web/src/lib/utils.ts 包含日期格式化、Token 解析等
- [x] T100 [P] 实现基础设施联调验证 server/tests/integration/foundation_health_test.go，检查数据库迁移、Redis 连接、统一响应格式和 `/api/health` 健康检查（含 AI 服务降级状态）

**Checkpoint**: 基础设施就绪，前后端可通信，数据库和 Redis 可连接

---

## Phase 3: User Story 1 - 用户注册与认证 (Priority: P1) 🎯 MVP

**Goal**: 用户可以注册、登录、自动刷新 Token，访问受保护页面

**Independent Test**: 注册新账号 → 登录 → 访问受保护页面 → 刷新页面验证会话持久 → Token 过期自动刷新

### 后端实现

- [x] T024 [US1] 实现 UserRepository server/internal/repository/user.go 提供 Create, FindByEmail, FindByID, Update 方法
- [x] T025 [US1] 实现 AuthService server/internal/service/auth.go 包含 Register, Login, RefreshToken, Logout 业务逻辑，密码使用 bcrypt 加密
- [x] T026 [US1] 实现 AuthHandler server/internal/handler/auth.go 实现 POST /api/auth/register, /login, /refresh, /logout 四个端点
- [x] T027 [US1] 实现 Refresh Token 管理逻辑，存储在 Redis 中，支持主动吊销（Logout 时清除）

### 前端实现

- [x] T028 [US1] 实现注册表单组件 web/src/components/auth/register-form.tsx 使用 React Hook Form + Zod 校验邮箱格式和密码强度
- [x] T029 [US1] 实现登录表单组件 web/src/components/auth/login-form.tsx 使用 React Hook Form + Zod 校验
- [x] T030 [US1] 实现注册页面 web/src/app/(auth)/register/page.tsx 调用 register API，成功后自动登录并跳转
- [x] T031 [US1] 实现登录页面 web/src/app/(auth)/login/page.tsx 调用 login API，存储 Token 到 auth store
- [x] T032 [US1] 实现认证路由保护 web/src/middleware.ts 使用 Next.js Middleware 检查 Token，未登录重定向到 /login
- [x] T033 [US1] 实现 Token 自动刷新机制集成到 API 客户端，401 响应时自动调用 refresh API 并重试原请求
- [x] T101 [P] [US1] 实现认证集成验证 server/tests/integration/auth_flow_test.go，覆盖注册、登录、受保护路由访问和 Token 自动刷新

**Checkpoint**: 用户可完成注册、登录全流程，会话自动持久化

---

## Phase 4: User Story 2 - AI 学习目标生成 (Priority: P1)

**Goal**: 用户输入自然语言学习目标，AI 自动生成结构化学习计划（含任务树、时间、资源），并可在生成后持续编辑计划

**Independent Test**: 输入"我想两个月学会 RAG" → 验证流式生成完整学习计划 → 查看任务层级结构 → 点击任务查看详情 → 新增、删除、重排并编辑任务

### 后端实现

- [x] T034 [US2] 实现 AI 客户端封装 server/internal/ai/client.go 封装 OpenAI Compatible API 调用，支持配置 endpoint、model、api_key
- [x] T035 [US2] 实现 SSE 流式处理 server/internal/ai/stream.go 封装 AI 流式响应读取和转发逻辑
- [x] T036 [US2] 实现 Prompt 模板 server/internal/ai/prompts.go 定义学习计划生成的 system prompt 和 user prompt 模板
- [x] T037 [US2] 实现 LearningGoal 模型 server/internal/model/learning_goal.go 按 data-model.md 定义，包含状态枚举 (active/paused/completed/archived)
- [x] T038 [US2] 实现 Task 模型 server/internal/model/task.go 按 data-model.md 定义，包含子任务自引用和依赖关系
- [x] T039 [US2] 实现 TaskDependency 模型 server/internal/model/task_dependency.go 定义任务依赖关系表
- [x] T040 [US2] 实现 LearningGoalRepository server/internal/repository/learning_goal.go 提供 CRUD 和按用户/状态查询
- [x] T041 [US2] 实现 TaskRepository server/internal/repository/task.go 提供 CRUD、按条件查询、子任务查询、依赖关系查询
- [x] T042 [US2] 实现 LearningGoalService server/internal/service/learning_goal.go 包含 Create（读取用户学习偏好并调用 AI 生成计划，同时持久化任务）、Update、List、GetByID，以及学习计划任务新增/删除/重排业务逻辑，集成指数退避重试 (1s/2s/4s, 最多 3 次)
- [x] T043 [US2] 实现 LearningGoalHandler server/internal/handler/learning_goal.go 实现 POST /api/learning-goals (SSE 流式返回)、GET /api/learning-goals、PUT /api/learning-goals/:id、PUT /api/learning-goals/:id/tasks/reorder、POST /api/learning-goals/:id/tasks、DELETE /api/learning-goals/:id/tasks/:taskId
- [x] T102 [P] [US2] 实现学习计划编辑验证 server/tests/integration/learning_goal_plan_edit_test.go，覆盖生成后任务新增、删除、重排、详情编辑和流式首包校验

### 前端实现

- [x] T044 [US2] 实现学习目标列表页 web/src/app/(goals)/goals/page.tsx 展示用户所有学习目标及进度
- [x] T045 [US2] 实现学习目标创建页面 web/src/app/(goals)/goals/new/page.tsx 包含自然语言输入框和提交按钮
- [x] T046 [US2] 实现 SSE 流式接收组件 web/src/components/goal/streaming-plan-viewer.tsx 实时展示 AI 生成的学习计划任务树
- [x] T047 [US2] 实现任务树组件 web/src/components/goal/task-tree.tsx 以层级树形结构展示任务，支持展开/折叠、拖拽重排，显示依赖关系
- [x] T048 [US2] 实现任务详情弹窗 web/src/components/task/task-detail-dialog.tsx 展示任务标题、描述、预计时长、推荐资源、状态
- [x] T049 [US2] 实现任务编辑功能 web/src/components/task/task-edit-form.tsx 支持编辑标题、描述、时长，以及手动新增/删除学习计划任务

**Checkpoint**: 用户可创建学习目标，AI 流式生成完整计划，并完成任务查看、新增、删除、重排与编辑

---

## Phase 5: User Story 3 - 任务管理 (Priority: P1)

**Goal**: 用户通过任务看板管理所有任务，支持创建、编辑、删除、搜索、筛选、排序，子任务和标签管理

**Independent Test**: 手动创建任务 → 拖拽切换状态 → 搜索关键词 → 按优先级/标签筛选 → 管理子任务 → 删除任务

### 后端实现

- [x] T050 [US3] 实现 Label 模型 server/internal/model/label.go 和 TaskLabel 关联模型
- [x] T051 [US3] 实现 LabelRepository server/internal/repository/label.go 提供 CRUD 和按用户查询
- [x] T052 [US3] 实现 TaskService server/internal/service/task.go 扩展 Create、Update、Delete（含子任务级联删除）、搜索、筛选、排序、游标分页逻辑
- [x] T053 [US3] 实现 TaskHandler server/internal/handler/task.go 实现 POST /api/tasks、GET /api/tasks (含筛选/搜索/排序/游标分页、deadline_from/deadline_to)、GET /api/tasks/:id、PUT /api/tasks/:id、DELETE /api/tasks/:id、POST /api/tasks/:id/dependencies
- [x] T054 [US3] 实现 LabelService 和 LabelHandler server/internal/service/label.go server/internal/handler/label.go 实现 POST /api/labels、GET /api/labels、PUT /api/labels/:id、DELETE /api/labels/:id

### 前端实现

- [x] T055 [US3] 实现任务看板页面 web/src/app/(tasks)/tasks/page.tsx 包含看板视图 (Todo/Doing/Done 三列) 和列表视图切换
- [x] T056 [US3] 实现看板列组件 web/src/components/task/kanban-column.tsx 展示单列任务卡片，支持拖拽
- [x] T057 [US3] 实现任务卡片组件 web/src/components/task/task-card.tsx 展示标题、优先级、标签、截止日期、子任务进度
- [x] T058 [US3] 实现任务创建表单 web/src/components/task/task-create-form.tsx 使用 React Hook Form + Zod 校验，支持设置标题、描述、优先级、标签、截止日期
- [x] T059 [US3] 实现搜索和筛选栏 web/src/components/task/task-filters.tsx 包含关键词搜索、状态/优先级/标签筛选器、截止日期范围筛选 (deadline_from/deadline_to)、排序选择
- [x] T060 [US3] 实现无限滚动加载 web/src/hooks/use-infinite-tasks.ts 使用 TanStack Query useInfiniteQuery 实现游标分页
- [x] T061 [US3] 实现标签管理组件 web/src/components/task/label-manager.tsx 支持创建、编辑、删除标签（含颜色选择）
- [x] T062 [US3] 实现子任务列表组件 web/src/components/task/subtask-list.tsx 展示子任务，支持添加、完成、删除子任务
- [x] T106 [P] [US3] 实现任务列表前端验证 web/tests/integration/task-filters.spec.tsx，覆盖关键词搜索、状态/优先级/标签/截止日期范围筛选与无限滚动联动
- [x] T107 [P] [US3] 实现标签管理前端验证 web/tests/components/label-manager.test.tsx，覆盖标签创建、编辑、删除和颜色更新
- [x] T103 [P] [US3] 实现任务管理验证 server/tests/integration/task_dependency_and_search_test.go，覆盖搜索/筛选/排序、无限滚动、依赖循环校验和未完成前置任务的 Done 拦截

**Checkpoint**: 任务管理全流程完成，支持看板拖拽、搜索筛选、无限滚动、标签和子任务管理

---

## Phase 6: User Story 4 - Dashboard 概览 (Priority: P2)

**Goal**: 用户打开 Dashboard 立即看到学习进度、统计数据、图表可视化

**Independent Test**: 完成多个任务 → 打开 Dashboard → 验证今日任务、完成率、学习时长、图表数据准确

### 后端实现

- [x] T063 [US4] 实现 StudySession 模型 server/internal/model/study_session.go 按 data-model.md 定义
- [x] T064 [US4] 实现 StudySessionRepository server/internal/repository/study_session.go 提供 CRUD 和按日期/用户聚合查询
- [x] T065 [US4] 实现 DashboardService server/internal/service/dashboard.go 聚合今日任务、完成率、学习时长、即将截止任务、最近活动
- [x] T066 [US4] 实现 DashboardHandler server/internal/handler/dashboard.go 实现 GET /api/dashboard/stats、GET /api/dashboard/charts/study-time、GET /api/dashboard/charts/category-stats、GET /api/dashboard/charts/completion-rate (FR-012)
- [x] T109 [US4] 实现完成率趋势查询在 DashboardService 中，GET /api/dashboard/charts/completion-rate 端点返回按日/周聚合的完成率数据 (FR-012)

### 前端实现

- [x] T067 [US4] 实现 Dashboard 页面 web/src/app/(dashboard)/dashboard/page.tsx 整体布局，响应式网格排列各统计卡片和图表
- [x] T068 [US4] 实现统计卡片组件 web/src/components/dashboard/stats-card.tsx 展示今日任务数、已完成数、完成率、学习时长
- [x] T069 [US4] 实现学习时长折线图 web/src/components/dashboard/study-time-chart.tsx 使用 Recharts 展示每日学习时长趋势 (支持 week/month 切换)
- [x] T070 [US4] 实现分类统计饼图 web/src/components/dashboard/category-chart.tsx 使用 Recharts 展示任务分类完成情况
- [x] T110 [US4] 实现完成率趋势折线图 web/src/components/dashboard/completion-rate-chart.tsx 使用 Recharts 展示每日/每周完成率变化趋势 (FR-012)
- [x] T071 [US4] 实现即将截止任务列表 web/src/components/dashboard/upcoming-deadlines.tsx 展示近期到期任务
- [x] T072 [US4] 实现最近活动列表 web/src/components/dashboard/recent-activity.tsx 展示最近的任务完成、计划创建等活动

**Checkpoint**: Dashboard 完整展示统计数据和图表

---

## Phase 7: User Story 5 - AI 聊天助手 (Priority: P2)

**Goal**: 用户可通过 AI 聊天助手提问学习问题，AI 基于上下文流式回答，并可将聊天拆解结果保存为任务树

**Independent Test**: 打开聊天 → 提问"Docker 怎么学习？" → 验证流式回答 → 追问"那具体怎么学？" → 验证上下文记忆 → 请求任务拆解并保存到学习目标

### 后端实现

- [x] T073 [US5] 实现 AIConversation 和 AIMessage 模型 server/internal/model/ai_conversation.go 按 data-model.md 定义
- [x] T074 [US5] 实现 AIConversationRepository server/internal/repository/ai_conversation.go 提供对话和消息的 CRUD，支持按对话加载完整消息历史
- [x] T075 [US5] 实现 AIChatService server/internal/service/ai_chat.go 实现 Chat、TaskBreakdown（读取用户学习偏好并生成可保存任务树）、ListConversations、GetMessages 逻辑，集成指数退避重试
- [x] T076 [US5] 实现 AIChatHandler server/internal/handler/ai_chat.go 实现 POST /api/ai/chat (SSE 流式返回)、POST /api/ai/task-breakdown、GET /api/ai/conversations、GET /api/ai/conversations/:id/messages

### 前端实现

- [x] T077 [US5] 实现聊天页面 web/src/app/(chat)/chat/page.tsx 左侧对话列表 + 右侧聊天窗口布局
- [x] T078 [US5] 实现对话列表组件 web/src/components/chat/conversation-list.tsx 展示历史对话，支持新建对话
- [x] T079 [US5] 实现聊天消息组件 web/src/components/chat/chat-message.tsx 渲染用户/AI 消息，AI 消息支持 Markdown 渲染
- [x] T080 [US5] 实现聊天输入框 web/src/components/chat/chat-input.tsx 支持 Enter 发送，Shift+Enter 换行
- [x] T081 [US5] 实现流式消息接收逻辑 web/src/hooks/use-chat-stream.ts 接收 SSE 流式响应并逐字渲染
- [x] T104 [P] [US5] 实现聊天任务拆解保存流程 web/src/hooks/use-chat-stream.ts 与 web/src/app/(chat)/chat/page.tsx，支持将 AI 拆解结果转换为可保存的任务树并关联学习目标

**Checkpoint**: AI 聊天助手可正常对话，支持多轮上下文、流式输出，以及任务拆解保存

---

## Phase 8: User Story 6 - AI 每日总结 (Priority: P3)

**Goal**: 用户可请求 AI 生成当日学习总结和明日建议

**Independent Test**: 完成任务 + 记录学习时长 → 请求每日总结 → 验证总结内容准确反映当日活动

### 后端实现

- [x] T082 [US6] 实现 StudySessionService server/internal/service/study_session.go 提供 Create、List、按日期聚合等业务逻辑
- [x] T083 [US6] 实现 StudySessionHandler server/internal/handler/study_session.go 实现 POST /api/study-sessions、GET /api/study-sessions
- [x] T084 [US6] 实现每日总结生成逻辑在 AIChatService 中，POST /api/ai/daily-summary 端点，聚合当日任务完成情况和学习时长，结合用户学习偏好调用 AI 生成总结 (SSE 流式返回)

### 前端实现

- [x] T085 [US6] 实现每日总结卡片 web/src/components/dashboard/daily-summary.tsx 在 Dashboard 中展示，支持点击生成总结
- [x] T086 [US6] 实现学习记录表单 web/src/components/task/study-session-form.tsx 支持手动记录学习时长和笔记

**Checkpoint**: 用户可记录学习时长，请求 AI 生成每日学习总结

---

## Phase 9: User Story 7 - 用户资料管理 (Priority: P3)

**Goal**: 用户可查看和编辑个人资料

**Independent Test**: 编辑显示名称、头像和学习偏好 → 保存 → 刷新页面验证持久化

### 后端实现

- [x] T087 [US7] 实现 UserService server/internal/service/user.go 提供 GetProfile、UpdateProfile 业务逻辑
- [x] T088 [US7] 实现 UserHandler server/internal/handler/user.go 实现 GET /api/user/profile、PUT /api/user/profile

### 前端实现

- [x] T089 [US7] 实现个人资料页面 web/src/app/(profile)/profile/page.tsx 展示和编辑用户信息
- [x] T090 [US7] 实现资料编辑表单 web/src/components/profile/profile-form.tsx 使用 React Hook Form + Zod，支持编辑名称、头像 URL、学习偏好设置（主题、语言、learning_style、weekly_study_hours、preferred_session_minutes）
- [x] T091 [US7] 实现暗黑模式切换组件 web/src/components/ui/theme-toggle.tsx 集成到导航栏，同步更新用户偏好
- [x] T108 [P] [US7] 实现资料页面前端验证 web/tests/components/profile-form.test.tsx，覆盖个人资料保存、偏好持久化和推荐偏好字段校验

**Checkpoint**: 用户可管理个人资料，暗黑模式正常切换

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: 横跨多个用户故事的优化和完善

- [x] T092 [P] 实现响应式导航栏 web/src/components/layout/navbar.tsx 支持移动端汉堡菜单，包含 Dashboard、任务、学习目标、聊天、个人资料入口
- [x] T093 [P] 实现侧边栏布局 web/src/components/layout/sidebar.tsx 桌面端侧边栏导航，移动端折叠
- [x] T094 统一全局错误处理前端 web/src/components/ui/error-boundary.tsx 实现 React Error Boundary 和 Toast 错误提示
- [x] T095 统一全局错误处理后端，完善所有 Handler 的错误响应，确保遵循 contracts/api.md 错误码规范
- [x] T096 [P] 实现前端加载状态 web/src/components/ui/loading-skeleton.tsx 为 Dashboard、任务列表、聊天等页面添加骨架屏
- [x] T097 性能优化：后端添加 Redis 缓存层用于 Dashboard 统计数据和频繁查询的任务列表
- [ ] T098 [P] 运行 quickstart.md 验证，确保 `docker-compose up`、`make dev`、健康检查端点全部正常工作
- [ ] T099 全流程集成验证：注册 → 创建学习目标 → AI 生成计划 → 管理任务 → 查看 Dashboard → AI 聊天 → 每日总结 → 编辑资料
- [x] T111 [P] 实现前端 E2E 验证覆盖 P1 UI 验收路径：注册/登录表单交互、路由保护重定向、流式计划生成 UI 渲染、任务创建 <15s 操作验证
- [x] T105 [P] 实现性能与并发验证 server/tests/performance/critical_paths_test.go，覆盖注册/登录 <60s (SC-001)、任务创建 <15s (SC-003)、任务搜索 <1s、Dashboard <2s、100 并发用户与 AI 首 token <3s，以及 AI 首次请求成功率 ≥95% (SC-009)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖，立即开始
- **Foundational (Phase 2)**: 依赖 Phase 1 完成，阻塞所有用户故事
- **US1 (Phase 3)**: 依赖 Phase 2，MVP 起点
- **US2 (Phase 4)**: 依赖 Phase 2 + US1 (需要认证)
- **US3 (Phase 5)**: 依赖 Phase 2 + US1 (需要认证)，与 US2 部分并行
- **US4 (Phase 6)**: 依赖 Phase 2 + US1 + US3 (需要任务数据)
- **US5 (Phase 7)**: 依赖 Phase 2 + US1 + US2 的 AI 基础设施
- **US6 (Phase 8)**: 依赖 Phase 2 + US1 + US3 + US5 的 AI 基础设施
- **US7 (Phase 9)**: 依赖 Phase 2 + US1
- **Polish (Phase 10)**: 依赖所有用户故事完成

### User Story Dependencies

- **US1 (P1)**: 无其他故事依赖，Phase 2 后立即开始
- **US2 (P1)**: 依赖 US1 (认证)，但 AI 客户端基础设施可与 US1 并行开发
- **US3 (P1)**: 依赖 US1 (认证)，可与 US2 并行开发
- **US4 (P2)**: 依赖 US3 (需要任务数据进行统计)
- **US5 (P2)**: 依赖 US1 (认证) + US2 (共享 AI 客户端与流式基础)
- **US6 (P3)**: 依赖 US3 (任务数据) + US5 (AI 客户端)
- **US7 (P3)**: 依赖 US1 (认证)，可与其他故事并行

### Within Each User Story

- Models → Repository → Service → Handler (后端分层顺序)
- 类型定义 → Hooks → 组件 → 页面 (前端分层顺序)
- 先后端再前端（可部分并行）

---

## Parallel Example: User Story 1

```bash
# 后端并行：
Task: "实现 UserRepository (T024)"
Task: "实现 AuthService (T025)"  # 部分依赖 T024，可先写接口

# 前端并行（后端完成后）：
Task: "实现注册表单 (T028)"
Task: "实现登录表单 (T029)"     # 不同文件，无依赖
```

## Parallel Example: User Story 2 + User Story 3

```bash
# US2 和 US3 可并行开发（不同文件，共享认证基础）：
Developer A: US2 (AI 学习目标生成)
Developer B: US3 (任务管理)

# US3 内部并行：
Task: "实现 Label 模型 (T050)"
Task: "实现 TaskRepository 扩展 (T041)"  # 不同文件
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational
3. 完成 Phase 3: User Story 1 (用户认证)
4. **STOP and VALIDATE**: 测试注册、登录、Token 刷新全流程
5. 可部署的最小可用版本

### Incremental Delivery

1. Setup + Foundational → 基础就绪
2. + US1 → 认证完成 → 可注册登录 (MVP!)
3. + US2 → AI 学习目标 → 核心差异化功能
4. + US3 → 任务管理 → 日常使用核心
5. + US4 → Dashboard → 数据可视化
6. + US5 → AI 聊天 → 增强体验
7. + US6 → 每日总结 → 锦上添花
8. + US7 → 用户资料 → 完善体验
9. + Polish → 生产就绪

---

## Notes

- [P] tasks = 不同文件，无依赖
- [Story] 标签标记任务所属用户故事
- 每个用户故事可独立完成和测试
- 每个 Checkpoint 后验证故事独立可用
- 后端严格遵循分层架构: Handler → Service → Repository → Model
- 前端严格遵循分层: 页面 → 组件 → Hooks → API 客户端
- Commit after each task or logical group
