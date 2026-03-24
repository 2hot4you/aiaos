-- ============================================================================
-- AIAOS — 数据库 DDL
-- PostgreSQL 16+
-- 生成日期: 2026-03-24
-- ============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. 用户表
-- ============================================================================
CREATE TABLE users (
    id              BIGINT PRIMARY KEY,
    username        VARCHAR(64)  NOT NULL,
    display_name    VARCHAR(128) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,           -- bcrypt hash
    role            VARCHAR(16)  NOT NULL DEFAULT 'user',  -- admin | user
    enabled         BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ                       -- 软删除
);

CREATE UNIQUE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users (role) WHERE deleted_at IS NULL;

-- ============================================================================
-- 2. 项目表
-- ============================================================================
CREATE TABLE projects (
    id              BIGINT PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    created_by      BIGINT       NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    -- [FIXED S8] 去掉 deleted_at，统一硬删除 + 级联。仅 users 表保留软删除
);

CREATE INDEX idx_projects_list ON projects (updated_at DESC); -- [FIXED S8] 去掉 deleted_at 条件

-- ============================================================================
-- 3. 季表
-- ============================================================================
CREATE TABLE seasons (
    id              BIGINT PRIMARY KEY,
    project_id      BIGINT       NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      BIGINT       NOT NULL REFERENCES users(id), -- [FIXED S1] 补充数据归属追踪
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_seasons_project ON seasons (project_id, sort_order);

-- ============================================================================
-- 4. 集表
-- ============================================================================
CREATE TABLE episodes (
    id              BIGINT PRIMARY KEY,
    season_id       BIGINT       NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    script_content  TEXT,                              -- 剧本内容（富文本 HTML/Markdown）
    config          JSONB        NOT NULL DEFAULT '{}', -- 项目配置（创作模式、语言、比例、时长、风格等）
    sort_order      INT          NOT NULL DEFAULT 0,
    created_by      BIGINT       NOT NULL REFERENCES users(id), -- [FIXED S1] 补充数据归属追踪
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_episodes_season ON episodes (season_id, sort_order);

-- ============================================================================
-- 5. 分镜表
-- ============================================================================
CREATE TABLE storyboard_shots (
    id                BIGINT PRIMARY KEY,
    episode_id        BIGINT       NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    shot_number       INT          NOT NULL,            -- [FIXED S2] 展示用镜号，reorder 时同步重新编号
    sort_order        INT          NOT NULL DEFAULT 0,  -- [FIXED S2] 纯排序字段，用于 ORDER BY
    scene_description TEXT,
    camera_movement   VARCHAR(64),                     -- 推/拉/摇/移/跟/固定/自定义
    dialogue          TEXT,
    action            TEXT,
    duration_ms       INTEGER      NOT NULL DEFAULT 4000, -- [FIXED B1] 毫秒级精度，方便计算总时长（原 VARCHAR(16) 无法参与数值运算）
    script_prompt     TEXT,
    visual_prompt     TEXT,
    status            VARCHAR(32)  NOT NULL DEFAULT 'pending',  -- pending | processing | completed | failed
    video_url         TEXT,                             -- S3 key
    thumbnail_url     TEXT,                             -- S3 key
    video_config      JSONB        NOT NULL DEFAULT '{}', -- 视频生成参数（模型、比例、时长、配音等）
    created_by        BIGINT       NOT NULL REFERENCES users(id), -- [FIXED S1] 补充数据归属追踪
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shots_episode ON storyboard_shots (episode_id, sort_order);
CREATE INDEX idx_shots_status ON storyboard_shots (status);

-- ============================================================================
-- 6. 资产表
-- ============================================================================
CREATE TABLE assets (
    id              BIGINT PRIMARY KEY,
    project_id      BIGINT       NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    episode_id      BIGINT       REFERENCES episodes(id) ON DELETE SET NULL,  -- 来源集，可为空（资产库复用）
    is_library      BOOLEAN      NOT NULL DEFAULT FALSE, -- [FIXED B4] 是否已入资产库（confirmed 后 is_library=true，资产独立存在不依赖分镜）
    type            VARCHAR(32)  NOT NULL,             -- character | scene | prop
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    image_prompt    TEXT,
    image_url       TEXT,                              -- S3 key
    thumbnail_url   TEXT,                              -- S3 key
    status          VARCHAR(32)  NOT NULL DEFAULT 'pending',  -- pending | confirmed
    metadata        JSONB        NOT NULL DEFAULT '{}', -- 附加信息（变体 URL、九宫格 URL 等）
    created_by      BIGINT       NOT NULL REFERENCES users(id), -- [FIXED S1] 补充数据归属追踪
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assets_project_type ON assets (project_id, type);
CREATE INDEX idx_assets_episode_type ON assets (episode_id, type);
CREATE INDEX idx_assets_status ON assets (status);

-- ============================================================================
-- 7. 分镜-资产关联表（多对多）
-- ============================================================================
CREATE TABLE shot_asset_relations (
    shot_id         BIGINT NOT NULL REFERENCES storyboard_shots(id) ON DELETE CASCADE,
    asset_id        BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    PRIMARY KEY (shot_id, asset_id)
);

CREATE INDEX idx_shot_assets_asset ON shot_asset_relations (asset_id);

-- ============================================================================
-- 8. 异步任务表
-- ============================================================================
CREATE TABLE tasks (
    id              BIGINT PRIMARY KEY,
    type            VARCHAR(64)  NOT NULL,             -- [FIXED S7] storyboard_generate | asset_image_generate | video_generate | export_compose（AI 改写/续写/风格反推改为同步 SSE，不走异步队列）
    status          VARCHAR(32)  NOT NULL DEFAULT 'pending',  -- pending | processing | completed | failed
    payload         JSONB        NOT NULL DEFAULT '{}', -- 任务输入参数
    result          JSONB,                             -- 任务输出结果
    error_message   TEXT,
    progress        INT          NOT NULL DEFAULT 0,   -- 0-100
    retry_count     INT          NOT NULL DEFAULT 0,
    max_retries     INT          NOT NULL DEFAULT 3,
    created_by      BIGINT       NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_tasks_status ON tasks (status, created_at);
CREATE INDEX idx_tasks_creator ON tasks (created_by, created_at DESC);
CREATE INDEX idx_tasks_type_status ON tasks (type, status);

-- ============================================================================
-- 9. 提示词模板表
-- ============================================================================
CREATE TABLE prompt_templates (
    id              BIGINT PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    type            VARCHAR(64)  NOT NULL,             -- storyboard | character | scene | prop | video | style | rewrite
    source          VARCHAR(32)  NOT NULL DEFAULT 'custom',  -- system | custom
    content         TEXT         NOT NULL,              -- 提示词内容，支持 {{variable}} 插值
    user_id         BIGINT       REFERENCES users(id), -- 创建者（system 预设为 NULL）
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prompts_type_source ON prompt_templates (type, source);
CREATE INDEX idx_prompts_user ON prompt_templates (user_id);

-- ============================================================================
-- 10. AI 模型配置表
-- ============================================================================
CREATE TABLE ai_model_configs (
    id               BIGINT PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    model_type       VARCHAR(32)  NOT NULL,            -- text | image | video
    provider         VARCHAR(64)  NOT NULL,            -- openai | google | sora | custom
    endpoint         TEXT         NOT NULL,             -- API 端点 URL
    api_key_enc      TEXT         NOT NULL,             -- AES-256 加密后的 API Key
    model_identifier VARCHAR(128) NOT NULL,             -- 模型标识，如 gpt-5.4、gemini-3.1-pro
    max_concurrency  INT          NOT NULL DEFAULT 5,
    timeout_seconds  INT          NOT NULL DEFAULT 60,
    is_default       BOOLEAN      NOT NULL DEFAULT FALSE,
    enabled          BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_models_type_default ON ai_model_configs (model_type, is_default) WHERE enabled = TRUE;

-- ============================================================================
-- 11. 系统设置表（KV 存储）
-- ============================================================================
CREATE TABLE system_settings (
    key             VARCHAR(128) PRIMARY KEY,
    value           JSONB        NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 12. 导出记录表
-- ============================================================================
CREATE TABLE exports (
    id              BIGINT PRIMARY KEY,
    episode_id      BIGINT       NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    task_id         BIGINT       REFERENCES tasks(id),
    status          VARCHAR(32)  NOT NULL DEFAULT 'pending', -- [FIXED S4] 冗余状态字段，避免 JOIN tasks 表 (pending | processing | completed | failed)
    resolution      VARCHAR(16)  NOT NULL DEFAULT '1080p',  -- 720p | 1080p | 4k
    format          VARCHAR(16)  NOT NULL DEFAULT 'mp4',
    file_url        TEXT,                              -- S3 key
    file_size       BIGINT,                            -- 字节数
    duration        INT,                               -- 总时长（秒）
    config          JSONB        NOT NULL DEFAULT '{}', -- 片头片尾、字幕、背景音乐等配置
    expires_at      TIMESTAMPTZ,                       -- 下载链接过期时间
    created_by      BIGINT       NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exports_episode ON exports (episode_id, created_at DESC);

-- ============================================================================
-- 13. 操作日志表（审计）
-- ============================================================================
CREATE TABLE audit_logs (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users(id),
    action          VARCHAR(128) NOT NULL,             -- login | create_user | delete_project | ...
    resource_type   VARCHAR(64),                       -- user | project | episode | ...
    resource_id     BIGINT,
    detail          JSONB,                             -- 操作详情
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs (user_id, created_at DESC);
CREATE INDEX idx_audit_action ON audit_logs (action, created_at DESC);

-- ============================================================================
-- 触发器：自动更新 updated_at
-- ============================================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为所有需要 updated_at 的表创建触发器
DO $$
DECLARE
    t TEXT;
BEGIN
    FOR t IN
        SELECT unnest(ARRAY[
            'users', 'projects', 'seasons', 'episodes',
            'storyboard_shots', 'assets', 'tasks',
            'prompt_templates', 'ai_model_configs', 'system_settings'
        ])
    LOOP
        EXECUTE format(
            'CREATE TRIGGER set_updated_at BEFORE UPDATE ON %I
             FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at()',
            t
        );
    END LOOP;
END;
$$;

-- ============================================================================
-- Seed 数据：初始管理员账号
-- ============================================================================
-- 密码: Admin@2026 (bcrypt hash, cost=12)
-- 生成命令: htpasswd -nbBC 12 "" "Admin@2026" | cut -d: -f2
INSERT INTO users (id, username, display_name, password_hash, role, enabled, created_at, updated_at)
VALUES (
    1,
    'admin',
    '系统管理员',
    '$2a$12$LJ3m4ys3Grl3CfGaVPq7/.QqF5Bv3FMqMIw0jPR9RGM8OfjBLxeda',
    'admin',
    TRUE,
    NOW(),
    NOW()
);

-- ============================================================================
-- Seed 数据：默认系统设置
-- ============================================================================
INSERT INTO system_settings (key, value, updated_at) VALUES
    ('platform_name', '"AIAOS"', NOW()),
    ('platform_logo', '""', NOW()),
    ('session_timeout_hours', '24', NOW()),
    ('max_upload_size_mb', '500', NOW()),
    ('default_language', '"zh"', NOW()),
    ('default_aspect_ratio', '"16:9"', NOW()),
    ('storage_type', '"s3"', NOW());

-- ============================================================================
-- Seed 数据：默认组件参数
-- ============================================================================
INSERT INTO system_settings (key, value, updated_at) VALUES
    ('component_storyboard', '{"default_mode": "standard", "max_shots": 50}', NOW()),
    ('component_image', '{"default_resolution": "1024x1024", "count": 1, "quality": "high"}', NOW()),
    ('component_video', '{"default_duration": 4, "default_aspect_ratio": "16:9", "max_concurrency": 3}', NOW()),
    ('component_voiceover', '{"enabled": false, "default_voice": "female_1", "default_speed": 1.0}', NOW()),
    ('component_export', '{"default_resolution": "1080p", "default_format": "mp4", "watermark": false}', NOW()),
    ('component_editor', '{"auto_save_interval_ms": 3000, "max_chars": 50000}', NOW());

-- ============================================================================
-- Seed 数据：系统预设提示词模板
-- ============================================================================
INSERT INTO prompt_templates (id, name, type, source, content, user_id, created_at, updated_at) VALUES
(
    100001,
    '标准模式 - 分镜生成',
    'storyboard',
    'system',
    '你是一个专业的影视分镜编剧。请根据以下剧本内容，生成分镜脚本。

## 要求
- 视觉风格：{{style}}
- 画面比例：{{aspect_ratio}}
- 目标时长：{{target_duration}}
- 输出语言：{{language}}

## 剧本内容
{{script_content}}

## 输出格式
请以 JSON 格式输出，包含 storyboard 数组和 identified_assets 对象。
每个分镜包含：shot_number, scene_description, camera_movement, dialogue, action, duration, script_prompt, visual_prompt。
identified_assets 包含：characters, scenes, props 数组，每个元素包含 name, description, image_prompt。',
    NULL,
    NOW(),
    NOW()
),
(
    100002,
    '高级模式 - 分镜生成',
    'storyboard',
    'system',
    '你是一个专业的影视分镜编剧。请根据以下已有的分镜脚本，优化并补充完整信息。

## 要求
- 视觉风格：{{style}}
- 画面比例：{{aspect_ratio}}
- 目标时长：{{target_duration}}
- 输出语言：{{language}}

## 分镜脚本
{{script_content}}

## 输出格式
请以 JSON 格式输出，保留原有镜头顺序，补充 visual_prompt 和 camera_movement。
包含 storyboard 数组和 identified_assets 对象。',
    NULL,
    NOW(),
    NOW()
),
(
    100003,
    '角色生图提示词',
    'character',
    'system',
    '{{style}} style, character design sheet, {{character_name}}, {{description}}, full body, front view, clean background, high detail, concept art',
    NULL,
    NOW(),
    NOW()
),
(
    100004,
    '场景生图提示词',
    'scene',
    'system',
    '{{style}} style, environment concept art, {{scene_description}}, wide shot, detailed background, cinematic lighting, {{aspect_ratio}}',
    NULL,
    NOW(),
    NOW()
),
(
    100005,
    '道具生图提示词',
    'prop',
    'system',
    '{{style}} style, prop design, {{name}}, {{description}}, isolated on clean background, multiple angles, detailed render',
    NULL,
    NOW(),
    NOW()
),
(
    100006,
    '视频生成提示词',
    'video',
    'system',
    '{{visual_prompt}}, {{camera_movement}} camera movement, {{action}}, cinematic, {{style}} style, {{aspect_ratio}}',
    NULL,
    NOW(),
    NOW()
),
(
    100007,
    'AI 改写指令',
    'rewrite',
    'system',
    '请根据以下指令改写文本。保持原文风格和语气，除非指令另有要求。

## 改写指令
{{instruction}}

## 原文
{{text}}

## 输出
直接输出改写后的文本，不要添加任何解释。',
    NULL,
    NOW(),
    NOW()
);
