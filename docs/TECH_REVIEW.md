# AIAOS 技术评审报告

> **评审人：** Tech Lead Agent  
> **评审日期：** 2026-03-24  
> **评审文档版本：** PRD v1.0 / ARCHITECTURE v1.0 / schema.sql v1.0  
> **状态：** 完成

---

## 评审总结

### 🟡 PASS WITH CONDITIONS

整体架构设计质量较高，分层清晰、技术选型合理、API 规范统一。在内部工具 + 小团队的场景下，复杂度控制得当，没有明显的过度设计。

但存在 **4 个 Blocker** 和 **12 个 Suggestion**，需在开发前修正 Blocker 项。

---

## 各维度评分

| 维度 | 评分 | 说明 |
|------|:----:|------|
| 1. 架构合理性 | ⭐⭐⭐⭐ (4/5) | 分层清晰，职责单一，依赖方向正确。单进程 API+Worker 的决策务实 |
| 2. 数据库设计 | ⭐⭐⭐⭐ (4/5) | 表结构覆盖全面，索引设计合理，有几处需要补充 |
| 3. API 设计 | ⭐⭐⭐⭐⭐ (5/5) | RESTful 规范统一，错误码体系完整，接口覆盖全面 |
| 4. 安全性 | ⭐⭐⭐⭐ (4/5) | 认证授权方案完善，但 Token 刷新和用户修改密码流程有缺失 |
| 5. 可扩展性 | ⭐⭐⭐⭐ (4/5) | AI Provider interface 设计良好，扩展性强 |
| 6. 技术选型 | ⭐⭐⭐⭐⭐ (5/5) | Go + Next.js + PostgreSQL 成熟可靠，中间件选择务实 |
| 7. 部署方案 | ⭐⭐⭐⭐ (4/5) | Docker Compose 覆盖完整，但缺少监控和日志采集方案 |

---

## Blocker（必须修改）

### B1. duration 字段类型不合理 — 应改为数值类型

**问题：** `storyboard_shots.duration` 定义为 `VARCHAR(16)`（存储 `"4s"` 这样的字符串），在排序、求和、比较时无法直接参与数值运算，成片预览需要计算总时长时需要额外解析。

**建议修改：**

```sql
-- 修改前
duration VARCHAR(16),  -- 如 "4s"

-- 修改后
duration_ms INT NOT NULL DEFAULT 4000,  -- 毫秒级精度，方便计算总时长
```

前端显示时做 `ms → "4s"` 的转换即可。API 层可以接受秒或毫秒，Domain 层统一用毫秒。

---

### B2. 缺少用户修改自身密码的接口

**问题：** PRD 6.9.1 提到管理员可以重置密码，但普通用户修改自己密码的能力完全缺失。对于内部工具来说，用户至少应该能修改自己的密码（初次登录后修改初始密码、定期更换密码等）。

**建议补充接口：**

```
PUT /api/v1/auth/password
Body: { "old_password": "xxx", "new_password": "yyy" }
```

---

### B3. JWT 刷新机制需明确 — 当前设计有安全隐患

**问题：** 架构文档提到了 `POST /api/v1/auth/refresh` 接口，但没有说明 Refresh Token 的存储和验证机制。如果用同一个 JWT 去刷新自身（自我刷新），那 Token 被盗后攻击者可以无限续期。

**建议：** 考虑到这是内部工具、并发用户不多，两种方案任选：

**方案 A（简单，推荐）：** 去掉 refresh 接口，JWT 有效期改为 7 天，过期后重新登录。内部工具可以接受。

**方案 B（标准）：** 引入 Refresh Token（存 Redis，与用户绑定），Access Token 有效期 1 小时，Refresh Token 有效期 7 天。刷新时返回新的 Access Token 并轮转 Refresh Token。

```go
// 方案 B 的 Redis 存储
// key: refresh_token:{token_hash} → value: user_id, TTL: 7d
```

---

### B4. 资产(assets)的 episode_id ON DELETE SET NULL 导致跨集资产孤立

**问题：** `assets.episode_id` 设置了 `ON DELETE SET NULL`，当删除集时资产的 `episode_id` 被置空。但 `assets.project_id` 的外键是 `ON DELETE CASCADE`——即删除项目时会级联删除资产。这导致一个矛盾：

- 删除集 → 资产变成"孤立资产"（episode_id = NULL），可通过资产库复用 ✅
- 但这些资产仍然属于项目（project_id 不变），删除项目时会级联删除 ✅
- **问题在于：** 没有机制区分"有意保留在资产库中的资产"和"被删除集遗留的垃圾资产"

**建议：** 增加 `is_library` 字段标识是否已入库，或在确认(confirm)时复制一份到资产库独立记录。

```sql
ALTER TABLE assets ADD COLUMN is_library BOOLEAN NOT NULL DEFAULT FALSE;
-- confirmed 且 is_library = true 的资产在集被删除后仍保留
-- 未入库的资产在集被删除后应清理
```

---

## Suggestion（建议修改）

### S1. 缺少 `created_by` 字段的数据归属追踪

**问题：** PRD 说"所有用户共享全部项目数据"，但 `seasons`、`episodes`、`storyboard_shots`、`assets` 等表缺少 `created_by` 字段。虽然不需要权限隔离，但对审计追踪和后续可能的权限扩展有价值。

**建议：** 至少在 `seasons` 和 `episodes` 表添加 `created_by BIGINT REFERENCES users(id)`。

---

### S2. `storyboard_shots.shot_number` 与 `sort_order` 职责重叠

**问题：** `shot_number` 是业务展示用的镜号，`sort_order` 是排序字段。两者可能不一致（用户拖拽排序后）。

**建议：** 保留两个字段，但明确语义：
- `shot_number`：展示用，拖拽排序后自动重新编号
- `sort_order`：纯排序字段，用于 ORDER BY

在 API 的 reorder 接口中，同时更新两个字段：

```go
// 排序后重新编号
for i, shotID := range newOrder {
    repo.Update(shotID, map[string]any{
        "sort_order":  i,
        "shot_number": i + 1,
    })
}
```

---

### S3. 建议给 `projects` 表增加 `name` 的模糊搜索索引

**问题：** PRD 要求项目列表支持按名称模糊搜索，当前只有 `idx_projects_list (deleted_at, updated_at DESC)` 索引，模糊搜索会走全表扫描。

**建议：**

```sql
-- 方案 A：pg_trgm 扩展（支持 LIKE '%keyword%'）
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_projects_name_trgm ON projects USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- 方案 B：如果只需要前缀搜索
CREATE INDEX idx_projects_name ON projects (name varchar_pattern_ops) WHERE deleted_at IS NULL;
```

考虑到内部工具项目数不会很多（几十到几百），可以先不加，等确认性能问题后再加。**低优先级。**

---

### S4. `exports` 表缺少状态字段

**问题：** `exports` 表有 `task_id` 关联到 `tasks` 表，但自身没有状态字段。前端查询导出列表时需要 JOIN `tasks` 表才能获取状态，不便。

**建议：** 增加冗余状态字段：

```sql
ALTER TABLE exports ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT 'pending';
-- pending | processing | completed | failed
```

---

### S5. 缺少 `GET /api/v1/auth/me` 在架构文档 API 路由表中已有，但 PRD 中未提及

**说明：** 架构文档补充了 `GET /api/v1/auth/me`，这是合理的。仅标注 PRD 和架构文档应保持同步。不需要修改。

---

### S6. Redis 缓存策略未明确

**问题：** 架构文档提到 Redis 用于 Cache / Session / Stream，但没有说明具体的缓存策略：
- 哪些数据需要缓存？（模型配置列表？系统设置？）
- TTL 多少？
- 缓存更新策略？（写穿透 / 旁路缓存？）

**建议：** 内部工具初期不需要复杂缓存。明确以下即可：
1. **登录失败计数：** Redis INCR，TTL 15min ✅（已有）
2. **系统设置/组件参数：** Cache-Aside，TTL 5min，更新时主动失效
3. **AI 模型配置：** Cache-Aside，TTL 5min，更新时主动失效
4. **其他数据不缓存**，直接查 DB（50 并发场景 PostgreSQL 完全承受得住）

---

### S7. AI 改写/续写不应走异步任务队列

**问题：** 架构文档 6.4 节将 `ai_rewrite` 和 `ai_continue` 列为异步任务类型，但 7.4 节又说"AI 改写/续写支持 SSE 流式返回"。这两者矛盾——如果走 Redis Stream 异步队列，就无法直接 SSE 流式返回给发起请求的用户。

**建议：** AI 改写/续写应走 **同步 SSE** 而非异步队列：

```
POST /api/v1/ai/rewrite (Accept: text/event-stream)
  → Handler 直接调用 AI Provider.GenerateText(stream=true)
  → 逐 chunk 通过 SSE 返回客户端
```

只有耗时较长的操作（生图、生视频、成片导出）才走异步队列。

**修改：** 将 `ai_rewrite`、`ai_continue`、`style_inference` 从异步任务类型列表中移除，改为同步 SSE 处理。`tasks` 表中不记录这些短任务。

---

### S8. 软删除的 `projects` 级联删除存在逻辑冲突

**问题：** `projects` 表有 `deleted_at` 软删除字段，但 `seasons.project_id` 设置了 `ON DELETE CASCADE`。如果软删除项目（UPDATE deleted_at），外键级联不会触发，seasons 不会被删除。如果硬删除（DELETE），级联才会生效。

需要明确：项目删除是软删除还是硬删除？

**建议：** 内部工具，软删除价值不大。建议统一为 **硬删除 + 级联**，减少复杂度。如果需要保留软删除：
- 在 Service 层实现"软删除级联"（手动标记所有子资源的 deleted_at）
- 或者只在 `users` 表保留软删除（因为用户有登录关联），`projects` 改为硬删除

---

### S9. 建议统一 HTTP 框架选型

**问题：** 架构文档在多处提到 `gin-gonic/gin`，handler 代码示例用了 Gin 语法（`gin.HandlerFunc`, `gin.Context`），但 3.1 目录结构注释写了"Gin/Chi"。建议明确选一个。

**建议：** 选 **Gin**。原因：
- 团队规模小，Gin 生态更成熟、社区资料多
- 中间件生态丰富（JWT、CORS、RateLimit 都有现成的）
- 附录 A 已经推荐了 Gin

删除对 Chi 的提及即可。

---

### S10. 缺少 API 版本迁移策略

**问题：** API 路由用 `/api/v1` 前缀，提到了"URL Path 版本控制"，但没有说明 v1 → v2 的迁移策略。

**建议：** 对于内部工具，简单说明即可：
> v1 为当前唯一版本。如需 Breaking Change，直接在 v1 上修改，前后端同步发布。暂不考虑多版本共存。

---

### S11. 补充数据库连接池配置建议

**建议在 `.env.example` 中补充：**

```bash
# ── Database Pool ──
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
```

pgx 默认连接池配置对于小规模内部工具够用，但显式配置更可控。

---

### S12. 建议补充监控和日志采集的最小方案

**问题：** PRD 8.4 提到需要"API 响应时间、错误率、AI 调用统计"监控，但架构文档没有给出方案。

**建议最小方案（适合内部工具）：**

1. **结构化日志：** zerolog → stdout → Docker logs（已有）
2. **健康检查端点：** `GET /healthz`（检查 DB + Redis 连接）
3. **基础 Metrics：** Prometheus 中间件 + `/metrics` 端点
4. **可选：** 后续加 Grafana + Loki（Docker Compose 中额外加两个服务）

```yaml
# docker-compose.yml 补充（可选，Phase 2+再加）
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
```

---

## 架构亮点

以下是设计中做得好的地方，值得肯定：

1. **Clean Architecture 分层** — Handler → Service → Repository 单向依赖，Interface 隔离外部依赖，可测试性好
2. **AI Provider Registry 动态注册** — 从 DB 加载模型配置、运行时动态注册，模型替换/新增零代码改动
3. **Redis Stream 异步任务** — 利用已有基础设施，不引入额外消息队列，复杂度可控
4. **SSE 推送方案** — 比 WebSocket 简单，满足服务端→客户端单向推送需求
5. **Snowflake ID** — 比 UUID 更适合 B-tree 索引，有序且不暴露业务信息
6. **JSONB 存配置** — 避免频繁 DDL 变更，适合配置类数据
7. **单进程 API + Worker** — 初期规模务实，避免微服务过早拆分
8. **触发器自动更新 updated_at** — 减少应用层遗漏

---

## 修改优先级总览

| 编号 | 类型 | 描述 | 优先级 | 工作量 |
|------|------|------|:------:|:------:|
| B1 | Blocker | duration 改为数值类型 | 🔴 高 | 小 |
| B2 | Blocker | 补充用户修改密码接口 | 🔴 高 | 小 |
| B3 | Blocker | 明确 JWT 刷新机制 | 🔴 高 | 中 |
| B4 | Blocker | 资产库标识与孤立资产处理 | 🔴 高 | 小 |
| S1 | Suggestion | 补充 created_by 字段 | 🟡 中 | 小 |
| S2 | Suggestion | 明确 shot_number vs sort_order 语义 | 🟡 中 | 小 |
| S3 | Suggestion | 项目名称搜索索引 | 🟢 低 | 小 |
| S4 | Suggestion | exports 补充 status 字段 | 🟡 中 | 小 |
| S6 | Suggestion | 明确 Redis 缓存策略 | 🟡 中 | 小 |
| S7 | Suggestion | AI 改写/续写改同步 SSE | 🟡 中 | 中 |
| S8 | Suggestion | 统一软删除/硬删除策略 | 🟡 中 | 中 |
| S9 | Suggestion | 明确 Gin 框架选型 | 🟢 低 | 小 |
| S10 | Suggestion | 补充 API 版本策略说明 | 🟢 低 | 小 |
| S11 | Suggestion | 补充 DB 连接池配置 | 🟢 低 | 小 |
| S12 | Suggestion | 补充最小监控方案 | 🟡 中 | 中 |

---

## 结论

架构设计整体扎实，适合内部工具 + 小团队的场景。解决 4 个 Blocker 后即可进入开发阶段。Suggestion 项可在开发过程中逐步补充。

**建议：修复 B1-B4 后，更新架构文档和 schema.sql，即可 PASS。**

---

> **评审完成** — Tech Lead Agent  
> 如有异议请在 Task Board 中标注讨论。
