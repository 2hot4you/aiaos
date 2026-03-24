# AIAOS — 系统架构设计文档

> **版本：** v1.0
> **日期：** 2026-03-24
> **作者：** Architect Agent
> **状态：** Draft

---

## 目录

1. [整体架构](#1-整体架构)
2. [前端架构](#2-前端架构)
3. [后端架构](#3-后端架构)
4. [数据库设计](#4-数据库设计)
5. [API 设计规范](#5-api-设计规范)
6. [异步任务系统](#6-异步任务系统)
7. [AI 模型调用层](#7-ai-模型调用层)
8. [文件存储设计](#8-文件存储设计)
9. [认证授权设计](#9-认证授权设计)
10. [部署架构](#10-部署架构)

---

## 1. 整体架构

### 1.1 架构总览

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Nginx (反向代理)                          │
│                    :80 / :443  TLS Termination                      │
└────────────┬──────────────────────────────┬──────────────────────────┘
             │ /                            │ /api/v1/*
             ▼                              ▼
┌────────────────────────┐    ┌──────────────────────────────────────┐
│   Next.js 15 (SSR)     │    │         Go Backend (:8080)           │
│   ─────────────────    │    │   ───────────────────────────────    │
│   • App Router         │    │   ┌─────────┐  ┌──────────────┐    │
│   • Ant Design 5       │    │   │ Handler │→│   Service    │    │
│   • Tailwind CSS       │    │   │ (HTTP)  │  │ (Business)   │    │
│   • Zustand            │    │   └─────────┘  └──────┬───────┘    │
│   • SSE Client         │    │                       │             │
└────────────────────────┘    │   ┌───────────────────┼──────────┐  │
                              │   │        ┌──────────┴────────┐ │  │
                              │   │        │    Repository     │ │  │
                              │   │        └──────────┬────────┘ │  │
                              │   └───────────────────┼──────────┘  │
                              └───────────┬───────────┼─────────────┘
                                          │           │
                    ┌─────────────────────┤           │
                    │                     │           │
                    ▼                     ▼           ▼
          ┌──────────────┐    ┌────────────┐  ┌──────────────┐
          │ Redis 7      │    │ PostgreSQL │  │ S3 / COS     │
          │ ───────────  │    │ 16         │  │ (文件存储)    │
          │ • Cache      │    │ (持久化)    │  └──────────────┘
          │ • Session    │    └────────────┘
          │ • Stream     │
          │   (任务队列)  │
          └──────┬───────┘
                 │ Stream Consumer
                 ▼
          ┌──────────────────────────────────────┐
          │       Task Worker (Go 协程池)         │
          │   ┌──────────┬──────────┬──────────┐ │
          │   │  GPT     │ Gemini  │  Sora    │ │
          │   │ (生文)    │ (生图)   │ (生视频)  │ │
          │   └──────────┴──────────┴──────────┘ │
          │         ↓ SSE Push / DB Update        │
          └──────────────────────────────────────┘
```

### 1.2 核心设计原则

| 原则 | 说明 |
|------|------|
| **Clean Architecture** | Handler → Service → Repository 单向依赖，业务逻辑不依赖框架 |
| **接口隔离** | AI 调用、存储等外部依赖全部通过 interface 抽象 |
| **异步优先** | 所有 AI 操作走 Redis Stream 异步任务队列 |
| **共享数据** | 所有用户共享全部项目数据，无用户级数据隔离 |
| **12-Factor** | 配置走环境变量，无状态服务，容器化部署 |

### 1.3 数据流向

```
用户操作 → Next.js → Go API → 同步响应（CRUD）
                            → 异步任务（AI 操作）
                                → Redis Stream XADD
                                → Worker XREADGROUP
                                → 调用 AI API
                                → 更新 DB + 上传文件到 S3
                                → SSE 推送前端
```

---

## 2. 前端架构

### 2.1 目录结构

```
frontend/
├── public/
│   ├── fonts/                    # Space Grotesk, DM Sans, JetBrains Mono
│   └── images/                   # 静态图片、Logo
├── src/
│   ├── app/                      # Next.js App Router
│   │   ├── layout.tsx            # 根布局（Dark Theme Provider）
│   │   ├── page.tsx              # 根页面 → redirect /projects
│   │   ├── login/
│   │   │   └── page.tsx          # 登录页
│   │   ├── projects/
│   │   │   ├── page.tsx          # 项目库首页
│   │   │   └── [id]/
│   │   │       └── page.tsx      # 剧集管理页
│   │   ├── workspace/            # 工作台（嵌套布局）
│   │   │   ├── layout.tsx        # 工作台布局（左侧菜单）
│   │   │   └── [pid]/[sid]/[eid]/
│   │   │       ├── page.tsx      # 剧本与故事（默认）
│   │   │       ├── assets/
│   │   │       │   └── page.tsx  # 角色与场景
│   │   │       ├── director/
│   │   │       │   └── page.tsx  # 导演工作台
│   │   │       ├── export/
│   │   │       │   └── page.tsx  # 成片与导出
│   │   │       └── prompts/
│   │   │           └── page.tsx  # 提示词管理
│   │   └── admin/
│   │       ├── layout.tsx        # 管理员布局
│   │       ├── users/
│   │       │   └── page.tsx      # 用户管理
│   │       ├── models/
│   │       │   └── page.tsx      # 模型管理
│   │       ├── settings/
│   │       │   └── page.tsx      # 系统设置
│   │       └── component-settings/
│   │           └── page.tsx      # 组件参数设置
│   ├── components/               # 通用组件
│   │   ├── ui/                   # 基础 UI 组件（封装 Ant Design）
│   │   │   ├── Button.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Card.tsx
│   │   │   └── ...
│   │   ├── layout/               # 布局组件
│   │   │   ├── AppHeader.tsx
│   │   │   ├── WorkspaceSidebar.tsx
│   │   │   └── AdminSidebar.tsx
│   │   ├── editor/               # 编辑器组件
│   │   │   ├── ScriptEditor.tsx
│   │   │   └── PromptEditor.tsx
│   │   ├── project/              # 项目相关组件
│   │   │   ├── ProjectCard.tsx
│   │   │   ├── StatsCards.tsx
│   │   │   └── CreateProjectModal.tsx
│   │   ├── storyboard/           # 分镜相关组件
│   │   │   ├── StoryboardCard.tsx
│   │   │   ├── StoryboardTimeline.tsx
│   │   │   └── StoryboardDetail.tsx
│   │   ├── asset/                # 资产相关组件
│   │   │   ├── AssetCard.tsx
│   │   │   ├── AssetLibraryModal.tsx
│   │   │   └── AssetUploader.tsx
│   │   └── video/                # 视频相关组件
│   │       ├── VideoPlayer.tsx
│   │       └── ExportPanel.tsx
│   ├── stores/                   # Zustand 状态管理
│   │   ├── useAuthStore.ts       # 认证状态
│   │   ├── useProjectStore.ts    # 项目列表状态
│   │   ├── useEpisodeStore.ts    # 当前集状态
│   │   ├── useStoryboardStore.ts # 分镜状态
│   │   ├── useAssetStore.ts      # 资产状态
│   │   ├── useTaskStore.ts       # 异步任务状态
│   │   └── useSettingsStore.ts   # 全局设置
│   ├── hooks/                    # 自定义 Hooks
│   │   ├── useSSE.ts             # SSE 连接管理
│   │   ├── useAutoSave.ts        # 自动保存
│   │   ├── useAuth.ts            # 认证逻辑
│   │   └── usePagination.ts      # 分页
│   ├── services/                 # API 调用层
│   │   ├── api.ts                # Axios 实例 + 拦截器
│   │   ├── authService.ts
│   │   ├── projectService.ts
│   │   ├── episodeService.ts
│   │   ├── storyboardService.ts
│   │   ├── assetService.ts
│   │   ├── taskService.ts
│   │   ├── promptService.ts
│   │   └── adminService.ts
│   ├── types/                    # TypeScript 类型定义
│   │   ├── auth.ts
│   │   ├── project.ts
│   │   ├── episode.ts
│   │   ├── storyboard.ts
│   │   ├── asset.ts
│   │   ├── task.ts
│   │   ├── prompt.ts
│   │   ├── admin.ts
│   │   └── api.ts                # 通用 API Response 类型
│   ├── lib/                      # 工具函数
│   │   ├── constants.ts          # 常量
│   │   ├── utils.ts              # 通用工具
│   │   └── format.ts             # 格式化（日期、相对时间等）
│   └── styles/
│       └── globals.css           # Tailwind 指令 + CSS 变量
├── tailwind.config.ts
├── next.config.ts
├── tsconfig.json
└── package.json
```

### 2.2 路由设计

| 路由 | 页面 | 权限 | 布局 |
|------|------|------|------|
| `/login` | 登录页 | Public | 无布局 |
| `/projects` | 项目库首页 | User/Admin | AppHeader |
| `/projects/:id` | 剧集管理 | User/Admin | AppHeader |
| `/workspace/:pid/:sid/:eid` | 剧本与故事 | User/Admin | WorkspaceSidebar |
| `/workspace/:pid/:sid/:eid/assets` | 角色与场景 | User/Admin | WorkspaceSidebar |
| `/workspace/:pid/:sid/:eid/director` | 导演工作台 | User/Admin | WorkspaceSidebar |
| `/workspace/:pid/:sid/:eid/export` | 成片与导出 | User/Admin | WorkspaceSidebar |
| `/workspace/:pid/:sid/:eid/prompts` | 提示词管理 | User/Admin | WorkspaceSidebar |
| `/admin/users` | 用户管理 | Admin | AdminSidebar |
| `/admin/models` | 模型管理 | Admin | AdminSidebar |
| `/admin/settings` | 系统设置 | Admin | AdminSidebar |
| `/admin/component-settings` | 组件参数 | Admin | AdminSidebar |

> **说明：** PRD 中工作台路径为 `/projects/:pid/seasons/:sid/episodes/:eid`，前端路由简化为 `/workspace/:pid/:sid/:eid` 以减少嵌套深度。API 路径保持与 PRD 一致。

### 2.3 状态管理方案

采用 **Zustand** 分片管理，每个业务域一个 store：

```typescript
// stores/useAuthStore.ts
interface AuthState {
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
}

// stores/useTaskStore.ts — 管理异步任务状态
interface TaskState {
  tasks: Record<string, Task>;        // taskId → Task
  subscribe: (taskId: string) => void; // 开始 SSE 监听
  updateTask: (task: Task) => void;
}
```

**状态管理原则：**

- 服务端数据用 **SWR 模式**（Zustand + API 调用），不做全局缓存
- 仅 UI 状态（侧边栏折叠、当前选中分镜等）放 Zustand
- 异步任务状态通过 SSE 实时更新
- 认证状态全局共享

### 2.4 SSE 客户端设计

```typescript
// hooks/useSSE.ts
function useSSE(taskId: string) {
  useEffect(() => {
    const source = new EventSource(`/api/v1/tasks/${taskId}/stream`);
    source.onmessage = (event) => {
      const data = JSON.parse(event.data);
      useTaskStore.getState().updateTask(data);
    };
    source.onerror = () => { source.close(); /* 重连逻辑 */ };
    return () => source.close();
  }, [taskId]);
}
```

### 2.5 Dark Theme 配置

通过 Ant Design 的 `ConfigProvider` + Tailwind CSS 变量实现：

```typescript
// app/layout.tsx
<ConfigProvider theme={{
  algorithm: theme.darkAlgorithm,
  token: {
    colorPrimary: '#6366F1',
    colorBgContainer: '#121828',
    colorBgElevated: '#1E293B',
    colorBgLayout: '#0A0E1A',
    borderRadius: 8,
  }
}}>
  {children}
</ConfigProvider>
```

---

## 3. 后端架构

### 3.1 目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go               # 入口：HTTP Server + Worker 启动
├── internal/
│   ├── domain/                   # 领域模型（纯结构体，无外部依赖）
│   │   ├── user.go               # User, Role
│   │   ├── project.go            # Project
│   │   ├── season.go             # Season
│   │   ├── episode.go            # Episode, EpisodeConfig, Script
│   │   ├── storyboard.go         # StoryboardShot
│   │   ├── asset.go              # Asset, AssetType, AssetStatus
│   │   ├── task.go               # Task, TaskType, TaskStatus
│   │   ├── prompt.go             # PromptTemplate, PromptType
│   │   ├── model_config.go       # AIModelConfig, ModelType
│   │   ├── setting.go            # SystemSetting, ComponentSetting
│   │   └── errors.go             # 业务错误定义
│   ├── service/                  # 业务逻辑层
│   │   ├── auth_service.go       # 登录、Token、密码
│   │   ├── project_service.go    # 项目 CRUD
│   │   ├── season_service.go     # 季 CRUD
│   │   ├── episode_service.go    # 集 CRUD + 配置 + 剧本
│   │   ├── storyboard_service.go # 分镜 CRUD + 排序
│   │   ├── asset_service.go      # 资产 CRUD + 确认 + 替换
│   │   ├── task_service.go       # 任务提交 + 状态查询 + SSE
│   │   ├── prompt_service.go     # 提示词管理
│   │   ├── admin_service.go      # 用户管理 + 模型管理 + 设置
│   │   └── export_service.go     # 成片导出
│   ├── repository/               # 数据访问层（接口 + 实现）
│   │   ├── interfaces.go         # 所有 Repository 接口定义
│   │   ├── postgres/             # PostgreSQL 实现
│   │   │   ├── user_repo.go
│   │   │   ├── project_repo.go
│   │   │   ├── season_repo.go
│   │   │   ├── episode_repo.go
│   │   │   ├── storyboard_repo.go
│   │   │   ├── asset_repo.go
│   │   │   ├── task_repo.go
│   │   │   ├── prompt_repo.go
│   │   │   ├── model_config_repo.go
│   │   │   └── setting_repo.go
│   │   └── redis/                # Redis 实现
│   │       ├── cache.go          # 通用缓存
│   │       └── session.go        # 登录失败计数、黑名单
│   ├── handler/                  # HTTP Handler（Gin/Chi）
│   │   ├── router.go             # 路由注册
│   │   ├── auth_handler.go
│   │   ├── project_handler.go
│   │   ├── season_handler.go
│   │   ├── episode_handler.go
│   │   ├── storyboard_handler.go
│   │   ├── asset_handler.go
│   │   ├── task_handler.go
│   │   ├── prompt_handler.go
│   │   ├── admin_handler.go
│   │   ├── export_handler.go
│   │   ├── sse_handler.go        # SSE 推送端点
│   │   └── response.go           # 统一响应格式
│   ├── middleware/                # 中间件
│   │   ├── auth.go               # JWT 验证
│   │   ├── rbac.go               # 角色权限校验
│   │   ├── cors.go               # CORS
│   │   ├── ratelimit.go          # 登录限流
│   │   ├── logger.go             # 请求日志
│   │   ├── recovery.go           # Panic 恢复
│   │   └── requestid.go          # 请求 ID 追踪
│   ├── ai/                       # AI 模型调用层
│   │   ├── provider.go           # Provider 接口定义
│   │   ├── registry.go           # Provider 注册表
│   │   ├── openai/               # OpenAI (GPT) 实现
│   │   │   └── client.go
│   │   ├── gemini/               # Google Gemini 实现
│   │   │   └── client.go
│   │   ├── sora/                 # Sora 实现
│   │   │   └── client.go
│   │   └── types.go              # 统一请求/响应类型
│   ├── storage/                  # 文件存储抽象
│   │   ├── storage.go            # Storage 接口
│   │   ├── s3/                   # S3 / COS 实现
│   │   │   └── client.go
│   │   └── local/                # 本地文件系统实现（开发用）
│   │       └── client.go
│   ├── queue/                    # Redis Stream 消费者
│   │   ├── producer.go           # 任务发布
│   │   ├── consumer.go           # Consumer Group 管理
│   │   ├── worker.go             # Worker 协程池
│   │   └── handlers/             # 各类任务处理器
│   │       ├── storyboard_gen.go # 分镜生成任务
│   │       ├── image_gen.go      # 图片生成任务
│   │       ├── video_gen.go      # 视频生成任务
│   │       ├── export.go         # 导出合成任务
│   │       └── ai_rewrite.go     # AI 改写/续写任务
│   └── config/                   # 配置管理
│       └── config.go             # 环境变量解析
├── pkg/                          # 公共工具包
│   ├── crypto/
│   │   ├── bcrypt.go             # 密码哈希
│   │   └── aes.go                # AES-256 加密（API Key）
│   ├── jwt/
│   │   └── jwt.go                # JWT 签发/验证
│   ├── validator/
│   │   └── validator.go          # 请求参数校验
│   ├── logger/
│   │   └── logger.go             # 结构化日志（zerolog/slog）
│   ├── httputil/
│   │   └── response.go           # 统一 HTTP 响应
│   └── idgen/
│       └── snowflake.go          # ID 生成（Snowflake 或 UUID）
├── migrations/                   # 数据库迁移
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── scripts/
│   ├── seed.sql                  # 初始数据
│   └── wait-for-it.sh            # Docker 启动等待脚本
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── go.sum
```

### 3.2 分层职责

```
┌─────────────────────────────────────────────┐
│  Handler (HTTP 层)                           │
│  • 解析请求参数、校验                          │
│  • 调用 Service                              │
│  • 构造 HTTP 响应                             │
│  • 不含业务逻辑                               │
├─────────────────────────────────────────────┤
│  Service (业务逻辑层)                         │
│  • 编排业务流程                               │
│  • 调用 Repository 持久化                     │
│  • 调用 Queue 提交异步任务                     │
│  • 业务规则校验                               │
├─────────────────────────────────────────────┤
│  Repository (数据访问层)                      │
│  • 数据库 CRUD                               │
│  • SQL 查询                                  │
│  • 不含业务逻辑                               │
├─────────────────────────────────────────────┤
│  Domain (领域模型)                            │
│  • 纯数据结构                                 │
│  • 枚举/常量                                  │
│  • 无外部依赖                                 │
└─────────────────────────────────────────────┘

外部依赖（全部通过 interface 隔离）：
  • AI Provider    → internal/ai/provider.go
  • Storage        → internal/storage/storage.go
  • Queue          → internal/queue/producer.go
  • Cache          → internal/repository/redis/cache.go
```

### 3.3 依赖注入

使用构造函数注入（不引入 DI 框架，保持简单）：

```go
// cmd/server/main.go 伪代码
func main() {
    cfg := config.Load()

    // 基础设施
    db := postgres.Connect(cfg.DB)
    rdb := redis.Connect(cfg.Redis)
    store := s3.NewClient(cfg.Storage)

    // Repository
    userRepo := postgres.NewUserRepo(db)
    projectRepo := postgres.NewProjectRepo(db)
    // ...

    // AI Provider
    aiRegistry := ai.NewRegistry()
    // 运行时从 DB 加载模型配置，动态注册 provider

    // Queue
    producer := queue.NewProducer(rdb)
    consumer := queue.NewConsumer(rdb, taskHandlers)

    // Service
    authSvc := service.NewAuthService(userRepo, rdb, cfg.JWT)
    projectSvc := service.NewProjectService(projectRepo)
    taskSvc := service.NewTaskService(taskRepo, producer)
    // ...

    // Handler + Router
    r := handler.NewRouter(authSvc, projectSvc, taskSvc, ...)

    // 启动 Worker
    go consumer.Start(ctx)

    // 启动 HTTP Server
    r.Run(":8080")
}
```

---

## 4. 数据库设计

### 4.1 ER 图

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌────────────────┐
│  users   │     │ projects │     │ seasons  │     │   episodes     │
│──────────│     │──────────│     │──────────│     │────────────────│
│ id (PK)  │     │ id (PK)  │     │ id (PK)  │     │ id (PK)        │
│ username │     │ name     │     │ project_id│───▶│ season_id      │
│ display_ │     │ created_ │     │ title    │     │ title          │
│ password │     │ updated_ │     │ sort_order│    │ script_content │
│ role     │     │ deleted_ │     │ created_ │     │ config (JSONB) │
│ enabled  │     └──────────┘     └──────────┘     │ created_at     │
│ created_ │           │                │           └────────────────┘
│ last_login│          │                │                  │
└──────────┘           │                │                  │
                       │                │                  │
                       ▼                ▼                  ▼
              ┌──────────────────────────────────────────────────┐
              │              storyboard_shots                     │
              │──────────────────────────────────────────────────│
              │ id (PK)  │ episode_id  │ shot_number │ sort_order│
              │ scene_description │ camera_movement │ dialogue   │
              │ action │ duration │ script_prompt │ visual_prompt│
              │ status │ video_url │ thumbnail_url │ created_at  │
              └──────────────────────────────────────────────────┘
                                       │
                                       ▼
              ┌──────────────────────────────────────────────────┐
              │                    assets                         │
              │──────────────────────────────────────────────────│
              │ id (PK) │ project_id │ episode_id │ type         │
              │ name │ description │ image_prompt │ image_url    │
              │ status │ confirmed │ metadata (JSONB) │ created_ │
              └──────────────────────────────────────────────────┘
                                       │
              ┌──────────────────────────────────────────────────┐
              │            shot_asset_relations                   │
              │──────────────────────────────────────────────────│
              │ shot_id (FK) │ asset_id (FK) │ (复合主键)         │
              └──────────────────────────────────────────────────┘

┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│      tasks       │    │ prompt_templates  │    │  ai_model_configs│
│──────────────────│    │──────────────────│    │──────────────────│
│ id (PK)          │    │ id (PK)          │    │ id (PK)          │
│ type             │    │ name             │    │ name             │
│ status           │    │ type             │    │ model_type       │
│ payload (JSONB)  │    │ source           │    │ provider         │
│ result (JSONB)   │    │ content          │    │ endpoint         │
│ error_message    │    │ user_id          │    │ api_key_enc      │
│ progress         │    │ created_at       │    │ model_identifier │
│ created_by       │    │ updated_at       │    │ max_concurrency  │
│ created_at       │    └──────────────────┘    │ timeout_seconds  │
│ updated_at       │                            │ is_default       │
└──────────────────┘                            │ enabled          │
                                                └──────────────────┘

┌─────────────────────────┐
│    system_settings      │
│─────────────────────────│
│ key (PK) │ value (JSONB)│
│ updated_at              │
└─────────────────────────┘
```

### 4.2 表结构详细设计

> 完整 DDL 见 `schema.sql`，此处列出设计要点。

#### 核心设计决策

| 决策 | 说明 |
|------|------|
| **主键策略** | 全部使用 `BIGINT` + Snowflake 算法生成，避免 UUID 索引性能问题 |
| **时间字段** | 统一 `TIMESTAMPTZ`，使用 UTC 存储 |
| **软删除** | `users`、`projects` 表有 `deleted_at` 字段 |
| **JSON 字段** | 配置/元数据使用 `JSONB`，便于灵活扩展 |
| **无用户隔离** | 项目/资产等表不按用户过滤，所有用户共享 |
| **审计字段** | 所有表包含 `created_at`、`updated_at` |

#### 索引设计要点

| 表 | 索引 | 类型 | 用途 |
|----|------|------|------|
| `users` | `username` | UNIQUE | 登录查找 |
| `projects` | `deleted_at, updated_at` | BTREE | 列表排序 |
| `seasons` | `project_id, sort_order` | BTREE | 项目下季排序 |
| `episodes` | `season_id` | BTREE | 季下集查找 |
| `storyboard_shots` | `episode_id, sort_order` | BTREE | 分镜排序 |
| `assets` | `project_id, type` | BTREE | 按项目+类型筛选 |
| `assets` | `episode_id, type` | BTREE | 按集+类型筛选 |
| `tasks` | `status, created_at` | BTREE | 任务队列查询 |
| `tasks` | `created_by, created_at` | BTREE | 用户任务列表 |
| `prompt_templates` | `type, source` | BTREE | 按类型+来源筛选 |
| `ai_model_configs` | `model_type, is_default` | BTREE | 查默认模型 |

---

## 5. API 设计规范

### 5.1 基本约定

| 项目 | 规范 |
|------|------|
| **Base URL** | `/api/v1` |
| **版本策略** | URL Path 版本控制 (`/api/v1`, `/api/v2`) |
| **命名风格** | 资源名用复数名词，snake_case |
| **HTTP 方法** | GET=查, POST=创建/操作, PUT=全量更新, DELETE=删除 |
| **请求格式** | `Content-Type: application/json` |
| **认证方式** | `Authorization: Bearer <JWT>` |
| **分页** | `?page=1&page_size=20` |
| **排序** | `?sort_by=updated_at&sort_order=desc` |
| **搜索** | `?keyword=xxx` |

### 5.2 统一响应格式

```json
// 成功
{
  "code": 0,
  "message": "ok",
  "data": { ... }
}

// 成功（分页）
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}

// 失败
{
  "code": 40101,
  "message": "用户名或密码错误",
  "data": null
}
```

### 5.3 错误码设计

采用 5 位数错误码：`XXYYY`，前两位表示模块，后三位表示具体错误。

| 范围 | 模块 | 说明 |
|------|------|------|
| `400xx` | 通用 | 参数校验、请求格式错误 |
| `401xx` | 认证 | 登录、Token 相关 |
| `403xx` | 权限 | RBAC 权限不足 |
| `404xx` | 资源 | 资源不存在 |
| `409xx` | 冲突 | 数据冲突 |
| `500xx` | 服务端 | 内部错误 |

**具体错误码：**

| 错误码 | 说明 |
|--------|------|
| `40001` | 请求参数无效 |
| `40002` | 请求体解析失败 |
| `40101` | 用户名或密码错误 |
| `40102` | Token 无效或过期 |
| `40103` | 账号已被禁用 |
| `40104` | 账号已被锁定（暴力破解防护） |
| `40301` | 权限不足（需要 Admin 角色） |
| `40401` | 项目不存在 |
| `40402` | 季不存在 |
| `40403` | 集不存在 |
| `40404` | 分镜不存在 |
| `40405` | 资产不存在 |
| `40406` | 任务不存在 |
| `40407` | 用户不存在 |
| `40408` | 模型配置不存在 |
| `40409` | 提示词不存在 |
| `40901` | 用户名已存在 |
| `50001` | 内部服务器错误 |
| `50002` | AI 模型调用失败 |
| `50003` | 文件上传失败 |
| `50004` | 任务队列错误 |

### 5.4 完整 API 路由表

```
# ── 认证 ──────────────────────────────────────
POST   /api/v1/auth/login                           # 登录
POST   /api/v1/auth/logout                          # 登出
POST   /api/v1/auth/refresh                         # 刷新 Token
GET    /api/v1/auth/me                               # 当前用户信息

# ── 项目 ──────────────────────────────────────
GET    /api/v1/projects                              # 项目列表（分页、搜索、排序）
GET    /api/v1/projects/stats                        # 项目统计
POST   /api/v1/projects                              # 创建项目
GET    /api/v1/projects/:id                          # 项目详情
PUT    /api/v1/projects/:id                          # 更新项目
DELETE /api/v1/projects/:id                          # 删除项目（软删除）
GET    /api/v1/projects/:id/stats                    # 单项目统计

# ── 季 ────────────────────────────────────────
GET    /api/v1/projects/:id/seasons                  # 季列表（含集信息）
POST   /api/v1/projects/:id/seasons                  # 创建季
PUT    /api/v1/seasons/:id                           # 更新季
DELETE /api/v1/seasons/:id                           # 删除季（级联）

# ── 集 ────────────────────────────────────────
POST   /api/v1/seasons/:id/episodes                  # 创建集
GET    /api/v1/episodes/:id                          # 集详情（含剧本、配置）
PUT    /api/v1/episodes/:id                          # 更新集
DELETE /api/v1/episodes/:id                          # 删除集（级联）
PUT    /api/v1/episodes/:id/script                   # 保存剧本
PUT    /api/v1/episodes/:id/config                   # 保存配置

# ── 分镜 ──────────────────────────────────────
POST   /api/v1/episodes/:id/storyboard/generate      # 生成分镜（异步）
GET    /api/v1/episodes/:id/storyboard               # 分镜列表
PUT    /api/v1/episodes/:id/storyboard/reorder       # 分镜排序
GET    /api/v1/storyboard/:id                        # 分镜详情
PUT    /api/v1/storyboard/:id                        # 更新分镜
POST   /api/v1/storyboard/:id/assets                 # 关联资产
POST   /api/v1/storyboard/:id/video/generate         # 生成视频（异步）
GET    /api/v1/storyboard/:id/video                  # 获取视频
POST   /api/v1/episodes/:id/video/batch-generate     # 批量生成视频

# ── 资产 ──────────────────────────────────────
GET    /api/v1/episodes/:id/assets                   # 资产列表（?type=character|scene|prop）
PUT    /api/v1/assets/:id                            # 更新资产
POST   /api/v1/assets/:id/generate                   # 生成资产图片（异步）
POST   /api/v1/assets/:id/variants                   # 生成服装变体（异步）
POST   /api/v1/assets/:id/poses                      # 生成造型九宫格（异步）
POST   /api/v1/assets/:id/reference-generate         # 参考图生图（异步）
POST   /api/v1/assets/:id/confirm                    # 确认资产
POST   /api/v1/assets/:id/upload                     # 手动上传
POST   /api/v1/assets/:id/replace                    # 从资产库替换
GET    /api/v1/asset-library                         # 资产库搜索

# ── AI 工具 ───────────────────────────────────
POST   /api/v1/ai/rewrite                            # AI 改写
POST   /api/v1/ai/continue                           # AI 续写
POST   /api/v1/ai/style-inference                    # 风格反推

# ── 任务 ──────────────────────────────────────
GET    /api/v1/tasks/:id                             # 查询任务状态
GET    /api/v1/tasks/:id/stream                      # SSE 任务进度流

# ── 导出 ──────────────────────────────────────
GET    /api/v1/episodes/:id/preview                  # 预览成片
POST   /api/v1/episodes/:id/export                   # 导出成片（异步）
GET    /api/v1/episodes/:id/download                 # 下载成片

# ── 提示词 ────────────────────────────────────
GET    /api/v1/prompts                               # 提示词列表
GET    /api/v1/prompts/:id                           # 提示词详情
POST   /api/v1/prompts                               # 创建提示词
PUT    /api/v1/prompts/:id                           # 更新提示词
DELETE /api/v1/prompts/:id                           # 删除提示词
POST   /api/v1/prompts/:id/duplicate                 # 复制提示词

# ── 管理员 ────────────────────────────────────
GET    /api/v1/admin/users                           # 用户列表
POST   /api/v1/admin/users                           # 创建用户
PUT    /api/v1/admin/users/:id                       # 更新用户
DELETE /api/v1/admin/users/:id                       # 删除用户（软删除）
POST   /api/v1/admin/users/:id/reset-password        # 重置密码
PUT    /api/v1/admin/users/:id/status                # 禁用/启用

GET    /api/v1/admin/models                          # 模型列表
POST   /api/v1/admin/models                          # 创建模型配置
PUT    /api/v1/admin/models/:id                      # 更新模型配置
DELETE /api/v1/admin/models/:id                      # 删除模型配置
POST   /api/v1/admin/models/:id/test                 # 测试连接

GET    /api/v1/admin/settings                        # 获取系统设置
PUT    /api/v1/admin/settings                        # 更新系统设置
GET    /api/v1/admin/component-settings              # 获取组件参数
PUT    /api/v1/admin/component-settings              # 更新组件参数
```

---

## 6. 异步任务系统

### 6.1 架构概览

```
┌──────────────┐         ┌─────────────────────────────┐
│  HTTP Handler│         │       Redis Stream           │
│  (API 请求)  │──XADD──▶│  stream: aiaos:tasks        │
└──────────────┘         │  consumer group: workers     │
                         └──────────────┬──────────────┘
                                        │ XREADGROUP
                         ┌──────────────┴──────────────┐
                         │     Worker Pool (N 个协程)    │
                         │  ┌────────┐ ┌────────┐      │
                         │  │Worker 1│ │Worker 2│ ...  │
                         │  └───┬────┘ └───┬────┘      │
                         │      │          │            │
                         │  ┌───▼──────────▼───┐       │
                         │  │  Task Dispatcher │       │
                         │  │  (按 type 分发)   │       │
                         │  └───┬──────────┬───┘       │
                         │      │          │            │
                         │  ┌───▼───┐  ┌───▼───┐       │
                         │  │ AI    │  │Export │  ...  │
                         │  │Handler│  │Handler│       │
                         │  └───┬───┘  └───┬───┘       │
                         └──────┼──────────┼───────────┘
                                │          │
                         ┌──────▼──────────▼───────────┐
                         │   更新 DB + SSE 推送          │
                         └─────────────────────────────┘
```

### 6.2 Redis Stream 设计

| 配置项 | 值 | 说明 |
|--------|-----|------|
| Stream Key | `aiaos:tasks` | 任务流 |
| Consumer Group | `workers` | 消费者组 |
| Consumer Name | `worker-{hostname}-{pid}` | 每个进程唯一 |
| 最大长度 | `MAXLEN ~10000` | 自动裁剪历史消息 |
| 读取策略 | `XREADGROUP ... COUNT 1 BLOCK 5000` | 每次读 1 条，阻塞 5s |

### 6.3 任务消息格式

```json
{
  "task_id": "1234567890",
  "type": "storyboard_generate",
  "payload": {
    "episode_id": "xxx",
    "config": { ... },
    "script_content": "..."
  },
  "created_by": "user_123",
  "created_at": "2026-03-24T10:00:00Z"
}
```

### 6.4 任务类型

| Type | 说明 | AI Provider | 预计耗时 |
|------|------|-------------|----------|
| `storyboard_generate` | 分镜生成 | GPT (生文) | 10-30s |
| `asset_image_generate` | 资产图片生成 | Gemini (生图) | 10-60s |
| `asset_variant_generate` | 服装变体生成 | Gemini (生图) | 10-60s |
| `asset_pose_generate` | 造型九宫格 | Gemini (生图) | 20-60s |
| `asset_reference_generate` | 参考图生图 | Gemini (生图) | 10-60s |
| `video_generate` | 视频生成 | Sora (生视频) | 30-180s |
| `video_batch_generate` | 批量视频生成 | Sora (生视频) | N × 30-180s |
| `export_compose` | 成片合成 | FFmpeg | 60-300s |
| `ai_rewrite` | AI 改写 | GPT (生文) | 5-15s |
| `ai_continue` | AI 续写 | GPT (生文) | 5-15s |
| `style_inference` | 风格反推 | Gemini (生图) | 10-30s |

### 6.5 任务状态机

```
pending → processing → completed
                    → failed → (可手动重试) → pending
```

| 状态 | 说明 |
|------|------|
| `pending` | 已入队，等待消费 |
| `processing` | 正在处理 |
| `completed` | 完成 |
| `failed` | 失败 |

### 6.6 SSE 推送

```
GET /api/v1/tasks/:id/stream

event: progress
data: {"task_id":"xxx","status":"processing","progress":50,"message":"正在生成分镜..."}

event: progress
data: {"task_id":"xxx","status":"processing","progress":80,"message":"解析 AI 返回..."}

event: complete
data: {"task_id":"xxx","status":"completed","result":{...}}

event: error
data: {"task_id":"xxx","status":"failed","error":"模型调用超时"}
```

**实现方式：** Worker 处理任务时，将进度写入 Redis Pub/Sub channel `task:{taskId}`。SSE Handler 订阅该 channel，转发给客户端。

```
Worker → Redis PUBLISH task:{taskId} progress
SSE Handler → Redis SUBSCRIBE task:{taskId} → EventSource → Client
```

### 6.7 失败处理与重试

| 策略 | 说明 |
|------|------|
| **自动重试** | AI 调用失败时，自动重试最多 3 次，指数退避 (1s, 4s, 16s) |
| **手动重试** | 最终失败后，用户可通过 UI 手动重试 |
| **超时保护** | 每个任务有最大执行时间（根据类型不同），超时自动标记 failed |
| **死信处理** | 超过重试次数的消息移入 `aiaos:tasks:dead` stream |
| **ACK 机制** | 处理完成后 XACK，未 ACK 的消息会被重新投递 |

---

## 7. AI 模型调用层

### 7.1 统一接口设计

```go
// internal/ai/provider.go

// Provider 是所有 AI 模型调用的统一接口
type Provider interface {
    // 文本生成（流式返回）
    GenerateText(ctx context.Context, req *TextRequest) (<-chan TextChunk, error)

    // 文本生成（一次性返回）
    GenerateTextSync(ctx context.Context, req *TextRequest) (*TextResponse, error)

    // 图片生成
    GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResponse, error)

    // 视频生成
    GenerateVideo(ctx context.Context, req *VideoRequest) (*VideoResponse, error)

    // 健康检查 / 连接测试
    HealthCheck(ctx context.Context) error
}

// TextRequest 文本生成请求
type TextRequest struct {
    Model       string            `json:"model"`
    SystemPrompt string           `json:"system_prompt"`
    UserPrompt  string            `json:"user_prompt"`
    MaxTokens   int               `json:"max_tokens"`
    Temperature float64           `json:"temperature"`
    Stream      bool              `json:"stream"`
    Extra       map[string]any    `json:"extra,omitempty"`
}

// TextChunk SSE 流式返回的文本块
type TextChunk struct {
    Content  string `json:"content"`
    Done     bool   `json:"done"`
}

// TextResponse 完整文本响应
type TextResponse struct {
    Content     string `json:"content"`
    TokensUsed  int    `json:"tokens_used"`
}

// ImageRequest 图片生成请求
type ImageRequest struct {
    Model       string         `json:"model"`
    Prompt      string         `json:"prompt"`
    Width       int            `json:"width"`
    Height      int            `json:"height"`
    Count       int            `json:"count"`        // 生成数量
    ReferenceImage []byte      `json:"reference_image,omitempty"`
    Extra       map[string]any `json:"extra,omitempty"`
}

// ImageResponse 图片生成响应
type ImageResponse struct {
    Images []GeneratedImage `json:"images"`
}

type GeneratedImage struct {
    Data     []byte `json:"data"`      // 图片二进制
    MimeType string `json:"mime_type"` // image/png
}

// VideoRequest 视频生成请求
type VideoRequest struct {
    Model       string         `json:"model"`
    Prompt      string         `json:"prompt"`
    ImageRef    string         `json:"image_ref,omitempty"`  // 参考图 URL
    Width       int            `json:"width"`
    Height      int            `json:"height"`
    Duration    int            `json:"duration"`   // 秒
    Extra       map[string]any `json:"extra,omitempty"`
}

// VideoResponse 视频生成响应
type VideoResponse struct {
    VideoData []byte `json:"video_data"`
    MimeType  string `json:"mime_type"`  // video/mp4
    Duration  int    `json:"duration"`
}
```

### 7.2 Provider 注册表

```go
// internal/ai/registry.go

type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider  // key = model_config_id
}

func (r *Registry) Register(configID string, p Provider) { ... }
func (r *Registry) Get(configID string) (Provider, error) { ... }
func (r *Registry) Remove(configID string) { ... }

// 从 DB 加载模型配置，动态创建 Provider
func (r *Registry) LoadFromDB(configs []domain.AIModelConfig) error {
    for _, cfg := range configs {
        var p Provider
        switch cfg.Provider {
        case "openai":
            p = openai.NewClient(cfg.Endpoint, decrypt(cfg.APIKeyEnc), cfg.Timeout)
        case "google":
            p = gemini.NewClient(cfg.Endpoint, decrypt(cfg.APIKeyEnc), cfg.Timeout)
        case "sora":
            p = sora.NewClient(cfg.Endpoint, decrypt(cfg.APIKeyEnc), cfg.Timeout)
        }
        r.Register(cfg.ID, p)
    }
    return nil
}
```

### 7.3 重试与超时策略

| 策略 | 参数 |
|------|------|
| **HTTP 超时** | 连接 10s，生文 60s，生图 120s，生视频 300s |
| **重试次数** | 最多 3 次 |
| **退避策略** | 指数退避：1s, 4s, 16s |
| **可重试错误** | 5xx、超时、连接拒绝 |
| **不可重试** | 4xx（参数错误、认证失败） |
| **并发控制** | 每个模型配置可设置最大并发数，使用 semaphore 限制 |

```go
// 伪代码
func (c *OpenAIClient) GenerateTextSync(ctx context.Context, req *TextRequest) (*TextResponse, error) {
    var lastErr error
    for attempt := 0; attempt <= c.maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(math.Pow(4, float64(attempt-1))) * time.Second
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        resp, err := c.doRequest(ctx, req)
        if err == nil {
            return resp, nil
        }
        if !isRetryable(err) {
            return nil, err
        }
        lastErr = err
    }
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### 7.4 SSE 流式返回（AI 改写/续写）

对于 AI 改写、续写等需要实时反馈的场景，支持 SSE 流式返回：

```
POST /api/v1/ai/rewrite
Accept: text/event-stream

→ 调用 GPT Stream API
→ 逐 chunk 转发给前端

event: chunk
data: {"content": "从前有"}

event: chunk
data: {"content": "一座山"}

event: done
data: {"content": "从前有一座山，山上有座庙..."}
```

---

## 8. 文件存储设计

### 8.1 存储接口

```go
// internal/storage/storage.go

type Storage interface {
    // 上传文件，返回存储路径（key）
    Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)

    // 获取文件预签名 URL（有效期）
    GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error)

    // 删除文件
    Delete(ctx context.Context, key string) error

    // 检查文件是否存在
    Exists(ctx context.Context, key string) (bool, error)
}
```

### 8.2 存储路径规范

```
aiaos/
├── assets/
│   ├── characters/
│   │   └── {asset_id}/
│   │       ├── original.png        # 原图
│   │       ├── thumbnail.png       # 缩略图
│   │       ├── variant_{n}.png     # 服装变体
│   │       └── poses.png           # 九宫格
│   ├── scenes/
│   │   └── {asset_id}/
│   │       ├── original.png
│   │       └── thumbnail.png
│   └── props/
│       └── {asset_id}/
│           ├── original.png
│           └── thumbnail.png
├── videos/
│   └── {shot_id}/
│       ├── raw.mp4                 # 原始生成视频
│       └── thumbnail.jpg           # 视频封面
├── exports/
│   └── {episode_id}/
│       └── {export_id}.mp4         # 导出成片
├── uploads/
│   ├── references/                 # 用户上传的参考图
│   │   └── {upload_id}.{ext}
│   └── logos/                      # Logo 上传
│       └── {filename}.{ext}
└── styles/
    └── {style_id}.png              # 风格参考图
```

### 8.3 文件处理流程

```
用户上传 → Handler 接收 multipart/form-data
         → 校验文件类型（白名单）+ 大小限制
         → 生成唯一 key
         → storage.Upload(key, file)
         → 返回 key，存入 DB

前端展示 → 请求资源 URL
         → storage.GetSignedURL(key, 1h)
         → 返回预签名 URL（1 小时有效）

AI 生成图片/视频 → Worker 调用 AI API 获取二进制
                 → storage.Upload(key, data)
                 → 更新 DB 记录 URL
```

### 8.4 文件类型白名单

| 类型 | 允许的 MIME | 最大大小 |
|------|------------|----------|
| 图片 | `image/png`, `image/jpeg`, `image/webp` | 20MB |
| 视频 | `video/mp4`, `video/webm` | 500MB |
| 音频 | `audio/mp3`, `audio/wav` | 50MB |

---

## 9. 认证授权设计

### 9.1 JWT 设计

```
┌──────────────────────────────────────────────┐
│  JWT Payload                                  │
│  {                                            │
│    "sub": "user_123",       // 用户 ID        │
│    "username": "zhangsan",  // 用户名          │
│    "role": "admin",         // 角色            │
│    "exp": 1711324800,       // 过期时间         │
│    "iat": 1711238400        // 签发时间         │
│  }                                            │
│                                               │
│  签名算法: HS256                               │
│  密钥: 环境变量 JWT_SECRET                      │
│  有效期: 24 小时（可配置）                       │
└──────────────────────────────────────────────┘
```

### 9.2 认证流程

```
1. POST /api/v1/auth/login {username, password}
2. 检查账号锁定状态（Redis key: login_fail:{username}）
3. 查 DB → bcrypt.CompareHashAndPassword
4. 失败 → Redis INCR login_fail:{username} (TTL 15min)
   → 累计 5 次 → 锁定（set lock:{username} TTL 15min）
5. 成功 → 清除失败计数
   → 签发 JWT
   → 更新 last_login_at
   → 返回 {token, user}
6. 后续请求 → Authorization: Bearer <token>
   → middleware 验证签名 + 过期时间
   → 注入 user 信息到 context
```

### 9.3 RBAC 权限中间件

```go
// 角色定义
const (
    RoleAdmin = "admin"
    RoleUser  = "user"
)

// 中间件链
router.Use(middleware.Auth())           // 所有需认证路由
router.Group("/admin").Use(middleware.RequireRole(RoleAdmin))  // Admin 路由

// middleware/rbac.go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        for _, role := range roles {
            if user.Role == role {
                c.Next()
                return
            }
        }
        c.JSON(403, ErrorResponse(40301, "权限不足"))
        c.Abort()
    }
}
```

### 9.4 密码与密钥安全

| 场景 | 方案 |
|------|------|
| **用户密码** | bcrypt 哈希，cost = 12 |
| **AI API Key** | AES-256-GCM 加密，密钥从环境变量 `ENCRYPTION_KEY` 读取 |
| **JWT Secret** | 环境变量 `JWT_SECRET`，至少 32 字节随机字符串 |

```go
// pkg/crypto/aes.go
func Encrypt(plaintext string, key []byte) (string, error) { ... }
func Decrypt(ciphertext string, key []byte) (string, error) { ... }
// 存储格式: base64(nonce + ciphertext + tag)
```

---

## 10. 部署架构

### 10.1 Docker Compose 架构

```
┌─────────────────────────────────────────────────┐
│                  Docker Network                  │
│                  (aiaos-network)                 │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │  nginx   │  │ frontend │  │   backend    │  │
│  │  :80/443 │─▶│  :3000   │  │   :8080      │  │
│  │          │─▶│          │  │ (API+Worker) │  │
│  └──────────┘  └──────────┘  └──────┬───────┘  │
│                                      │          │
│                        ┌─────────────┼────────┐ │
│                        │             │        │ │
│                   ┌────▼─────┐  ┌────▼─────┐  │ │
│                   │ postgres │  │  redis   │  │ │
│                   │  :5432   │  │  :6379   │  │ │
│                   │ (volume) │  │ (volume) │  │ │
│                   └──────────┘  └──────────┘  │ │
│                                               │ │
│                                               │ │
└───────────────────────────────────────────────┘ │
```

### 10.2 docker-compose.yml 结构

```yaml
version: "3.9"

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/certs:/etc/nginx/certs:ro
    depends_on:
      - frontend
      - backend

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    environment:
      - NEXT_PUBLIC_API_BASE_URL=/api/v1
    depends_on:
      - backend

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=aiaos
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
      - S3_ENDPOINT=${S3_ENDPOINT}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
      - S3_BUCKET=${S3_BUCKET}
      - S3_REGION=${S3_REGION}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=aiaos
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
```

### 10.3 环境变量清单

```bash
# .env.example

# ── Database ──
DB_USER=aiaos
DB_PASSWORD=your_secure_password
DB_NAME=aiaos

# ── Security ──
JWT_SECRET=your_jwt_secret_at_least_32_chars
ENCRYPTION_KEY=your_aes256_key_32_bytes_hex

# ── S3 Storage ──
S3_ENDPOINT=https://cos.ap-guangzhou.myqcloud.com
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=aiaos-assets
S3_REGION=ap-guangzhou

# ── Server ──
SERVER_PORT=8080
LOG_LEVEL=info
GIN_MODE=release

# ── Worker ──
WORKER_CONCURRENCY=5
TASK_STREAM_KEY=aiaos:tasks
TASK_CONSUMER_GROUP=workers
```

### 10.4 Nginx 配置要点

```nginx
server {
    listen 80;
    server_name aiaos.internal;

    # 前端静态资源 + SSR
    location / {
        proxy_pass http://frontend:3000;
        proxy_set_header Host $host;
    }

    # API 代理
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Request-ID $request_id;

        # SSE 支持
        proxy_set_header Connection '';
        proxy_http_version 1.1;
        chunked_transfer_encoding off;
        proxy_buffering off;
        proxy_cache off;
    }

    # 文件上传大小
    client_max_body_size 500M;
}
```

---

## 附录 A：Go 后端推荐依赖

| 库 | 用途 | 选型理由 |
|----|------|----------|
| `gin-gonic/gin` | HTTP 框架 | 性能好、生态成熟 |
| `jackc/pgx/v5` | PostgreSQL 驱动 | 原生支持、性能优 |
| `redis/go-redis/v9` | Redis 客户端 | 功能全面、Stream 支持好 |
| `aws/aws-sdk-go-v2` | S3 客户端 | 兼容腾讯 COS |
| `golang-jwt/jwt/v5` | JWT 处理 | 标准实现 |
| `rs/zerolog` | 结构化日志 | 零分配、高性能 |
| `golang-migrate/migrate` | 数据库迁移 | 主流方案 |
| `go-playground/validator` | 参数校验 | 功能丰富 |
| `swaggo/swag` | Swagger 文档 | 注解自动生成 |

## 附录 B：关键设计决策记录

| # | 决策 | 原因 |
|---|------|------|
| 1 | **无用户级数据隔离** | 内部平台，所有用户共享项目数据，简化查询逻辑 |
| 2 | **Snowflake ID 而非 UUID** | BIGINT 索引性能优于 UUID，有序性有助于 B-tree |
| 3 | **Redis Stream 而非 RabbitMQ** | 基础设施已有 Redis，无需引入新组件 |
| 4 | **单进程 API + Worker** | 初期规模小（≥50 并发），单进程足够；后续可拆 |
| 5 | **SSE 而非 WebSocket** | 只需要服务端→客户端推送，SSE 更简单 |
| 6 | **JSONB 存配置** | 配置字段灵活多变，JSONB 避免频繁改表 |
| 7 | **前端路由简化** | `/workspace/:pid/:sid/:eid` 而非完整层级，减少嵌套 |
| 8 | **bcrypt cost=12** | 安全与性能的平衡点，单次哈希约 250ms |