# HaruhiServer 开发计划

| 阶段  | 主题            | 当前目标                | 你要完成的内容                                                     | 阶段产出        | 完成标准                                        |
| --- | ------------- | ------------------- | ----------------------------------------------------------- | ----------- | ------------------------------------------- |
| P0  | 项目初始化         | 搭好最小工程              | 建目录、`go mod init`、主程序入口、基础 README、Makefile                  | 一个能启动的空项目   | `go run` 能启动，不报错                            |
| P1  | 最小 HTTP 服务    | 先让 server 活起来       | `main.go`、`http.Server`、基础 router、`/healthz`                | 第一个可访问接口    | 浏览器或 curl 能访问 `/healthz`                    |
| P2  | 配置与日志         | 先让程序像一个服务           | 配置结构体、默认值、环境变量读取、`slog` 日志初始化                               | 可配置的服务程序    | 能通过环境变量修改端口、日志正常输出                          |
| P3  | 统一响应与错误       | 先统一“说话方式”           | `response helper`、统一 JSON 格式、业务错误定义、HTTP 错误映射               | 返回风格统一      | 所有接口都能按统一 JSON 输出                           |
| P4  | 路由与中间件骨架      | 建立请求处理链路            | router 集中注册、request id、recover、logging、cors 中间件             | 有组织的 HTTP 层 | 请求经过中间件链，日志可看到 request id                   |
| P5  | Domain 层建模    | 把核心对象先定义清楚          | `User`、`Project`、`Task`、`Note`、`Session`、`AuditLog`、枚举、领域错误 | 核心模型稳定      | domain 文件齐全，命名统一                            |
| P6  | Repository 抽象 | 先做“数据访问边界”          | 定义 repository 接口，写 memory 版 `map + sync.RWMutex` 实现         | 可替换的数据层     | memory repository 可正常增删改查                   |
| P7  | Service 分层    | 建立业务层，不让 handler 变脏 | 定义 service 接口与默认实现，先写最基础逻辑                                  | 分层正式成立      | handler 不直接操作 repository                    |
| P8  | System 模块     | 做最简单但完整的业务模块        | `/healthz`、`/readyz`、`/api/v1/system/info`                  | 第一组完整模块     | 能返回服务名、版本、启动时间等                             |
| P9  | Project 模块    | 完成第一个完整 CRUD 模块     | DTO、handler、service、repository 接口串起来，实现 projects 增删改查       | 第一条完整业务链路   | `POST/GET/PATCH/DELETE /api/v1/projects` 跑通 |
| P10 | Task 模块       | 学习嵌套资源设计            | 按 `projectId` 管理任务，支持列表、详情、更新、删除                            | 第二个业务模块     | `/projects/{projectId}/tasks/...` 跑通        |
| P11 | Note 模块       | 重复并巩固模块化实现          | 按 `projectId` 管理笔记，完成 CRUD                                  | 第三个业务模块     | `/projects/{projectId}/notes/...` 跑通        |
| P12 | Auth 基础       | 打通登录认证链路            | password hash、JWT 骨架、register、login、me                      | 最小认证系统      | 能注册、登录、带 token 访问 `/me`                     |
| P13 | Auth 中间件      | 把用户身份接入请求上下文        | Bearer Token 解析、用户信息写入 context、受保护路由                        | 有鉴权能力的服务    | 未登录访问受保护接口会被拒绝                              |
| P14 | Users 模块      | 完成用户信息访问链路          | `GET /users`、`GET /users/{id}`、`PATCH`、`DELETE` 骨架          | 用户管理模块      | users 路由结构完整，可正常返回                          |
| P15 | 测试骨架          | 建立基本验证能力            | handler 测试、service 测试、memory repo 测试                        | 基础测试体系      | `go test ./...` 可运行                         |
| P16 | 优雅关闭与清理       | 让项目更像真实服务           | `SIGINT/SIGTERM`、graceful shutdown、超时设置、资源清理                | 可正常停止的服务    | Ctrl+C 后能平稳退出                               |
| P17 | 收尾与整理         | 形成一份合格骨架            | README、目录说明、配置示例、TODO 标注、代码整理                               | 可继续迭代的项目骨架  | 项目结构清晰，别人能读懂                                |

## P9 Project 模块接口

已实现 `/api/v1/projects` 的完整 CRUD：

- `POST /api/v1/projects`：创建项目
- `GET /api/v1/projects`：项目列表
- `GET /api/v1/projects/{id}`：项目详情
- `PATCH /api/v1/projects/{id}`：更新项目（支持 `name`、`description`、`visibility`、`archived`）
- `DELETE /api/v1/projects/{id}`：删除项目
