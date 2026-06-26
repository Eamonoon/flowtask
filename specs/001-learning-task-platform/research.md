# Research: FlowTask - AI Learning Plan & Task Management Platform

**Branch**: `001-learning-task-platform` | **Date**: 2026-06-17

## Research Findings

### Decision 1: JWT 认证方案

**Decision**: 使用 Access Token + Refresh Token 双令牌机制

**Rationale**:
- Access Token 短期有效（15 分钟），减少泄露风险
- Refresh Token 长期有效（7 天），存储在 HttpOnly Cookie 中，防止 XSS
- 后端 Gin 中间件统一校验，前端 Axios 拦截器自动刷新
- 用户无感知的 Token 轮换，体验流畅

**Alternatives considered**:
- Session-based 认证：需要服务端状态存储，不适合前后端分离架构
- OAuth2/SSO：v1 仅支持邮箱注册，无需第三方登录

---

### Decision 2: AI 流式输出方案

**Decision**: 后端使用 Server-Sent Events (SSE) 转发 AI 流式响应

**Rationale**:
- OpenAI Compatible API 原生支持 SSE 流式输出
- 后端 Gin 作为中间层透传 SSE 流，同时处理重试和错误
- 前端使用 EventSource 或 fetch + ReadableStream 接收
- SSE 比 WebSocket 更轻量，单向推送场景完全满足需求

**Alternatives considered**:
- WebSocket：双向通信，但本场景只需服务端向客户端推送，SSE 更简单
- 轮询：延迟高，实时性差，浪费资源
- 后端生成后一次性返回：用户等待时间长，体验差

---

### Decision 3: 数据库设计策略

**Decision**: PostgreSQL 作为主数据库，Redis 用于缓存和会话管理

**Rationale**:
- PostgreSQL 支持 JSONB 字段，适合存储 AI 生成的灵活结构数据（学习计划、推荐资源）
- 支持全文搜索（tsvector），满足任务搜索需求
- Redis 缓存热点数据（用户会话、Dashboard 统计、频繁查询的任务列表）
- Redis 存储 Refresh Token 白名单，支持主动吊销

**Alternatives considered**:
- MongoDB：文档型更适合灵活数据，但关系型更适合用户-任务-计划的复杂关联查询
- Elasticsearch：全文搜索更强，但 v1 规模下 PostgreSQL 全文搜索足够

---

### Decision 4: 前端状态管理

**Decision**: Zustand + React Query (TanStack Query)

**Rationale**:
- Zustand 轻量、无 boilerplate，适合管理客户端状态（UI 状态、用户偏好）
- React Query 管理服务端状态，自动处理缓存、重新获取、乐观更新
- 无限滚动与 React Query 的 useInfiniteQuery 天然集成
- 避免 Redux 的繁重模板代码

**Alternatives considered**:
- Redux Toolkit：功能强大但模板代码多，对本项目规模过度
- Jotai/Recoil：原子化状态管理，适合小型项目但缺乏服务端状态管理
- SWR：类似 React Query 但功能较少，无限滚动支持不够好

---

### Decision 5: 图表库选择

**Decision**: Recharts

**Rationale**:
- React 原生组件，与 Next.js 和 shadcn/ui 生态完美集成
- 声明式 API，学习成本低
- 支持响应式和暗黑模式定制
- 足够覆盖 Dashboard 需要的折线图、柱状图、饼图

**Alternatives considered**:
- Chart.js：功能更全但需额外封装 React 组件
- ECharts：功能最全但包体积大，中文社区强但本项目图表需求简单
- Nivo：美观但文档和社区活跃度不如 Recharts

---

### Decision 6: 项目结构

**Decision**: Monorepo 前后端分离结构

**Rationale**:
- 前端 Next.js 和后端 Go 各自独立构建和部署
- 通过 API 契约解耦，前端可独立开发
- 使用 Makefile/docker-compose 统一开发环境启动

**Structure**:
```
flowtask/
├── web/                    # 前端 (Next.js)
│   ├── src/
│   │   ├── app/            # App Router 路由
│   │   ├── components/     # UI 组件
│   │   ├── lib/            # 工具函数
│   │   ├── hooks/          # 自定义 Hooks
│   │   ├── stores/         # Zustand stores
│   │   └── types/          # TypeScript 类型
│   └── tests/
├── server/                 # 后端 (Go)
│   ├── cmd/                # 入口
│   ├── internal/
│   │   ├── handler/        # HTTP handlers
│   │   ├── service/        # 业务逻辑
│   │   ├── repository/     # 数据访问
│   │   ├── model/          # 数据模型
│   │   ├── middleware/     # 中间件
│   │   └── config/         # 配置
│   └── tests/
├── docker-compose.yml
└── Makefile
```
