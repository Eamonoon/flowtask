# Quickstart: FlowTask

**Branch**: `001-learning-task-platform` | **Date**: 2026-06-17

## 前置条件

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选，推荐用于数据库)

## 快速启动

### 1. 克隆并启动数据库

```bash
# 启动 PostgreSQL 和 Redis
docker-compose up -d postgres redis

# 或手动确保 PostgreSQL 和 Redis 运行在默认端口
```

### 2. 后端 (Go)

```bash
cd server

# 复制配置文件
cp config.example.yaml config.yaml

# 编辑 config.yaml，配置数据库连接和 AI API 密钥

# 安装依赖
go mod tidy

# 运行数据库迁移
go run cmd/migrate/main.go

# 启动服务
go run cmd/server/main.go
```

后端默认运行在 `http://localhost:8080`

### 3. 前端 (Next.js)

```bash
cd web

# 安装依赖
npm install

# 复制环境变量
cp .env.example .env.local

# 编辑 .env.local，配置 API 地址
# NEXT_PUBLIC_API_URL=http://localhost:8080/api

# 启动开发服务器
npm run dev
```

前端默认运行在 `http://localhost:3000`

### 4. 使用 Makefile (推荐)

```bash
# 启动所有服务
make dev

# 仅启动后端
make dev-server

# 仅启动前端
make dev-web

# 运行测试
make test

# 构建生产版本
make build
```

## 环境变量

### 后端 (server/config.yaml)

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  user: flowtask
  password: flowtask
  dbname: flowtask
  sslmode: disable

redis:
  host: localhost
  port: 6379
  db: 0

jwt:
  access_secret: your-access-secret-key
  refresh_secret: your-refresh-secret-key
  access_ttl: 15m
  refresh_ttl: 168h  # 7 days

ai:
  api_key: your-openai-api-key
  base_url: https://api.openai.com/v1
  model: gpt-4o
  max_retries: 3
  timeout: 60s
```

### 前端 (web/.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

## 验证安装

```bash
# 检查后端健康状态
curl http://localhost:8080/api/health

# 预期返回
# {"code":0,"data":{"status":"ok","database":"ok","redis":"ok"}}
```

打开浏览器访问 `http://localhost:3000`，注册账号开始使用。
