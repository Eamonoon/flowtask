# Implementation Plan: FlowTask - AI Learning Plan & Task Management Platform

**Branch**: `001-learning-task-platform` | **Date**: 2026-06-17 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-learning-task-platform/spec.md`

## Summary

开发一个 AI 驱动的学习计划和任务管理平台。用户可以用自然语言描述学习目标，AI 自动生成结构化学习路线图和任务列表。平台提供完整的任务管理、Dashboard 数据可视化、AI 聊天助手和每日学习总结功能。

前后端完全分离：前端 Next.js App Router + shadcn/ui，后端 Go + Gin + GORM，通过 REST API + SSE 流式通信连接。

## Technical Context

**Language/Version**: Go 1.21+ (后端), TypeScript 5.x (前端)

**Primary Dependencies**:
- 后端: Gin, GORM, JWT (golang-jwt), go-redis
- 前端: Next.js 14 App Router, shadcn/ui, React Hook Form, Zod, TanStack Query, Zustand, Recharts

**Storage**: PostgreSQL 15+ (主数据库), Redis 7+ (缓存/会话)

**Testing**: Go testing + testify (后端), Vitest + Testing Library (前端)

**Target Platform**: Linux/macOS 开发环境，Web 浏览器访问

**Project Type**: Web Application (前后端分离)

**Performance Goals**: 
- 页面加载 < 2s
- AI 流式响应首 token < 3s
- 搜索响应 < 1s (1000 条任务内)
- 支持 100 并发用户

**Constraints**:
- JWT 双令牌认证 (Access + Refresh)
- AI 请求指数退避重试 (1s/2s/4s, 最多 3 次)
- 支持暗黑模式和响应式

**Scale/Scope**:
- 7 个用户故事 (P1: 3, P2: 2, P3: 2)
- 22 个功能需求
- 9 个数据表
- 约 32 个 API 端点

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Spec Before Build: 已完成澄清并同步 `spec.md`
- Testable Delivery: `tasks.md` 已补充基础联调、P1 故事验证、性能与并发验证、前端 E2E 验证 (T108)
- Contract-First Integration: `contracts/api.md`、`data-model.md`、`tasks.md`
  已统一字段命名 (`parent_task_id`)、补充完成率趋势端点、明确 Refresh Token 传输方式
- Operability By Default: 保留健康检查（含 AI 服务降级观测）、统一错误处理、AI 重试与 Redis/PostgreSQL
  依赖恢复策略
- Keep Scope Honest: v1 不包含邮件/推送通知和密码重置；个人资料仅包含昵称、头像和学习偏好

## Project Structure

### Documentation (this feature)

```text
specs/001-learning-task-platform/
├── plan.md              # 本文件
├── research.md          # 技术研究决策
├── data-model.md        # 数据模型设计
├── quickstart.md        # 快速启动指南
├── contracts/           # API 契约
│   └── api.md           # REST API 定义
└── tasks.md             # 任务清单 (/speckit-tasks 命令生成)
```

### Source Code (repository root)

```text
flowtask/
├── web/                        # 前端 (Next.js)
│   ├── src/
│   │   ├── app/                # App Router 路由页面
│   │   │   ├── (auth)/         # 认证相关页面 (登录/注册)
│   │   │   ├── (dashboard)/    # Dashboard 页面
│   │   │   ├── (tasks)/        # 任务管理页面
│   │   │   ├── (goals)/        # 学习目标页面
│   │   │   ├── (chat)/         # AI 聊天页面
│   │   │   ├── (profile)/      # 个人资料页面
│   │   │   └── layout.tsx      # 根布局
│   │   ├── components/         # UI 组件
│   │   │   ├── ui/             # shadcn/ui 基础组件
│   │   │   ├── task/           # 任务相关组件
│   │   │   ├── goal/           # 学习目标相关组件
│   │   │   ├── dashboard/      # Dashboard 组件
│   │   │   ├── chat/           # AI 聊天组件
│   │   │   ├── profile/        # 个人资料组件
│   │   │   └── layout/         # 导航与布局组件
│   │   ├── lib/                # 工具函数
│   │   │   ├── api.ts          # API 客户端
│   │   │   ├── auth.ts         # 认证工具
│   │   │   └── utils.ts        # 通用工具
│   │   ├── hooks/              # 自定义 Hooks
│   │   ├── stores/             # Zustand stores
│   │   └── types/              # TypeScript 类型定义
│   ├── public/
│   ├── tests/                  # Vitest + Testing Library 测试
│   │   ├── integration/
│   │   └── components/
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   └── tsconfig.json
│
├── server/                     # 后端 (Go)
│   ├── cmd/
│   │   ├── server/main.go      # 服务入口
│   │   └── migrate/main.go     # 数据库迁移入口
│   ├── internal/
│   │   ├── handler/            # HTTP 处理器
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   ├── learning_goal.go
│   │   │   ├── dashboard.go
│   │   │   ├── ai_chat.go
│   │   │   ├── label.go
│   │   │   ├── study_session.go
│   │   │   └── health.go
│   │   ├── service/            # 业务逻辑层
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   ├── learning_goal.go
│   │   │   ├── dashboard.go
│   │   │   ├── ai_chat.go
│   │   │   ├── label.go
│   │   │   └── study_session.go
│   │   ├── repository/         # 数据访问层
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   ├── task_dependency.go
│   │   │   ├── learning_goal.go
│   │   │   ├── study_session.go
│   │   │   ├── label.go
│   │   │   └── ai_conversation.go
│   │   ├── model/              # 数据模型
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   ├── task_dependency.go
│   │   │   ├── learning_goal.go
│   │   │   ├── study_session.go
│   │   │   ├── label.go
│   │   │   └── ai_conversation.go
│   │   ├── middleware/         # 中间件
│   │   │   ├── auth.go         # JWT 认证中间件
│   │   │   ├── cors.go
│   │   │   └── logger.go
│   │   ├── ai/                 # AI 客户端封装
│   │   │   ├── client.go       # OpenAI Compatible API 客户端
│   │   │   ├── prompts.go      # Prompt 模板
│   │   │   └── stream.go       # SSE 流式处理
│   │   ├── config/             # 配置管理
│   │   │   └── config.go
│   │   └── response/           # 统一响应格式
│   │       └── response.go
│   ├── migrations/             # 数据库迁移文件
│   ├── tests/                  # Go 集成与性能测试
│   │   ├── integration/
│   │   └── performance/
│   ├── go.mod
│   └── go.sum
│
├── docker-compose.yml          # 开发环境容器编排
├── Makefile                    # 开发命令
└── README.md
```

**Structure Decision**: 采用前后端分离的 Monorepo 结构。前端 `web/` 基于 Next.js App Router，后端 `server/` 采用 Go 分层架构 (Handler → Service → Repository → Model)，AI 客户端封装为独立包。两个项目通过 REST API + SSE 流式通信连接。
