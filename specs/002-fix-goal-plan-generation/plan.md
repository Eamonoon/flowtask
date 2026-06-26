# Implementation Plan: 修复学习目标生成学习计划功能

**Branch**: `002-fix-goal-plan-generation` | **Date**: 2026-06-23 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-fix-goal-plan-generation/spec.md`

## Summary

修复学习目标模块的两个核心问题：1) 前端"生成学习计划"按钮缺少完整的状态反馈流程，用户点击后无响应；2) 后端 AI 响应解析失败，无法处理 markdown 代码块格式的 JSON。需要完善前端流式显示体验和增强后端 JSON 解析容错能力。

## Technical Context

**Language/Version**: Go 1.21+ (后端), TypeScript 5.x (前端)

**Primary Dependencies**: 
- 后端: Gin, GORM, JWT (golang-jwt), go-redis, OpenAI Go SDK
- 前端: Next.js 14 App Router, shadcn/ui, TanStack Query, Zustand

**Storage**: PostgreSQL 15+ (主数据库), Redis 7+ (缓存/会话)

**Testing**: Go testing + testify (后端), Vitest + Testing Library (前端)

**Target Platform**: Linux/macOS 开发环境，Web 浏览器访问

**Project Type**: Web Application (前后端分离)

**Performance Goals**: 
- 首个任务显示时间 < 3 秒 (SC-001)
- 流式实时显示任务生成 (SC-002)
- API 解析错误率 < 1% (SC-003)

**Constraints**: 
- JWT 双令牌认证 (Access + Refresh)
- SSE (Server-Sent Events) 流式传输
- AI 请求指数退避重试

**Scale/Scope**: 修复现有功能，不涉及新增用户规模或数据量

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Spec Before Build: ✅ `spec.md` 已完成，所有澄清已集成到 spec 中
- Testable Delivery: ✅ P1 用户故事（US1: 即时反馈, US2: 解析修复）都有独立测试路径和可测量成功标准
- Contract-First Integration: ✅ `contracts/api.md` 已定义错误码规范（40001, 40002, 40003, 50001, 50002），任务中包含错误码验证测试
- Operability By Default: ✅ Phase 7 包含健康检查（T056）、结构化日志（T057）、监控指标（T058）、Redis 缓存（T059）
- Keep Scope Honest: ✅ 无范围扩展，仅修复现有功能

## Project Structure

### Documentation (this feature)

```text
specs/002-fix-goal-plan-generation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command)
```

### Source Code (repository root)

```text
server/
├── cmd/
│   ├── server/main.go
│   ├── migrate/main.go
│   └── cleanup/main.go             # 新增：定时清理过期 session
├── internal/
│   ├── handler/
│   │   ├── learning_goal.go        # 修复 JSON 解析
│   │   └── health.go               # 新增：健康检查端点
│   ├── service/
│   │   ├── learning_goal.go        # 增强 AI 响应解析容错
│   │   └── generation_session.go   # 新增：session 生命周期管理
│   ├── ai/
│   │   ├── client.go               # 新增：结构化日志
│   │   ├── parser.go               # 新增：多层 JSON 解析
│   │   ├── stream.go               # 修复 SSE 流式传输
│   │   └── retry.go                # 新增：指数退避重试
│   ├── repository/
│   │   ├── generation_session.go   # 新增：session repository
│   │   └── generated_task.go       # 新增：临时任务 repository
│   └── model/
│       ├── learning_goal.go
│       ├── generation_session.go   # 新增：生成会话模型
│       └── generated_task.go       # 新增：临时任务模型
├── tests/
│   ├── integration/
│   │   ├── test_learning_goal_create.go
│   │   ├── test_learning_goal_stream.go
│   │   ├── test_learning_goal_confirm.go
│   │   ├── test_learning_goal_parse.go
│   │   ├── test_learning_goal_errors.go   # 新增：错误码验证测试
│   │   └── test_learning_goal_resume.go
│   ├── unit/
│   │   ├── test_ai_parser.go
│   │   └── test_ai_retry.go
│   └── performance/
│       ├── test_learning_goal_perf.go     # 新增：性能测试
│       └── test_learning_goal_load.go     # 新增：负载测试
└── migrations/
    └── 002_add_generation_sessions.sql

web/
├── src/
│   ├── app/(dashboard)/goals/
│   │   └── page.tsx               # 重构状态反馈流程
│   ├── components/goal/
│   │   ├── streaming-plan-viewer.tsx  # 增强流式显示和确认保存
│   │   └── error-boundary.tsx         # 新增：错误边界组件
│   ├── hooks/
│   │   └── use-goal-stream.ts     # 新增：封装生成状态逻辑
│   ├── stores/
│   │   └── goal-store.ts          # 新增：管理生成状态
│   └── lib/
│       └── goal-api.ts            # 新增：目标相关 API 调用封装
└── tests/
    ├── integration/
    │   ├── test-goal-generation.spec.tsx
    │   ├── test-goal-error-scenarios.spec.tsx  # 新增：错误场景测试
    │   └── test-goal-resume.spec.tsx
    └── unit/
        └── test-goal-store.ts
```

**Structure Decision**: 采用现有 monorepo 结构，后端 `server/` 和前端 `web/` 分离。主要修改集中在 handler/service 层（后端）和 goals 页面组件（前端）。新增 hooks 和 stores 用于状态管理。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

无违反项，所有 Constitution 原则均满足。
