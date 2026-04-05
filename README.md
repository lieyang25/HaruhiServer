# HaruhiServer

HaruhiServer 是一个面向个人学习与小型团队协作的后端服务骨架项目，主题是「项目、任务、笔记协作管理服务」。

本仓库当前提供：
- 可编译、可运行的分层工程结构
- `net/http` + `slog` + 优雅关闭
- 统一 JSON 响应与统一错误映射
- `repository` 接口抽象与并发安全 `memory` 实现（`map + sync.RWMutex`）
- Auth/Users/Projects/Tasks/Notes/System 路由骨架

> 说明：本项目刻意不实现完整复杂业务，关键未完成能力均通过 `TODO` 标注。

## 快速启动

1. 准备配置：
```bash
cp configs/config.example.env .env
```

2. 导出环境变量（示例）
```bash
export APP_ENV=dev
export HTTP_PORT=8080
export JWT_SECRET=change-me-in-production
```

3. 启动服务：
```bash
make run
```

4. 健康检查：
```bash
curl http://127.0.0.1:8080/healthz
```

## 目录结构

```text
.
├── cmd/haruhiserver/main.go
├── configs/config.example.env
├── internal
│   ├── app
│   │   └── bootstrap.go
│   ├── auth
│   │   ├── jwt.go
│   │   └── password.go
│   ├── config
│   │   └── config.go
│   ├── domain
│   │   ├── errors.go
│   │   └── models.go
│   ├── repository
│   │   ├── interfaces.go
│   │   └── memory
│   │       ├── auditlog_repository.go
│   │       ├── interfaces_assert.go
│   │       ├── note_repository.go
│   │       ├── project_repository.go
│   │       ├── session_repository.go
│   │       ├── task_repository.go
│   │       └── user_repository.go
│   ├── server
│   │   └── http_server.go
│   ├── service
│   │   ├── auth_service.go
│   │   ├── id.go
│   │   ├── note_service.go
│   │   ├── project_service.go
│   │   ├── system_service.go
│   │   ├── task_service.go
│   │   └── user_service.go
│   └── transport/http
│       ├── dto
│       ├── handler
│       ├── middleware
│       ├── response
│       └── router.go
├── Makefile
└── go.mod
```

## API 路由

### Auth
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

### Users
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

### Projects
- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{id}`
- `PATCH /api/v1/projects/{id}`
- `DELETE /api/v1/projects/{id}`

### Tasks
- `POST /api/v1/projects/{projectId}/tasks`
- `GET /api/v1/projects/{projectId}/tasks`
- `GET /api/v1/projects/{projectId}/tasks/{taskId}`
- `PATCH /api/v1/projects/{projectId}/tasks/{taskId}`
- `DELETE /api/v1/projects/{projectId}/tasks/{taskId}`

### Notes
- `POST /api/v1/projects/{projectId}/notes`
- `GET /api/v1/projects/{projectId}/notes`
- `GET /api/v1/projects/{projectId}/notes/{noteId}`
- `PATCH /api/v1/projects/{projectId}/notes/{noteId}`
- `DELETE /api/v1/projects/{projectId}/notes/{noteId}`

### System
- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/system/info`

## 工程约定

- Handler 只做参数解析、调用 service、返回响应。
- Service 依赖 repository 接口，不依赖具体实现。
- 默认仓储实现是内存版，后续可替换为 MySQL/PostgreSQL 等实现。
- 统一错误通过 `internal/domain/errors.go` 定义，并在 HTTP 层映射状态码。
- DTO 与 domain model 分离。

## 测试

```bash
make test
```

当前提供基础测试骨架，后续可围绕 service 与 middleware 增补单测、集成测试。
# HaruhiServer
