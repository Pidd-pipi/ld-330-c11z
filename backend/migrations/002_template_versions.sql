-- 002_template_versions.sql：病历模板版本化
-- 模板内容迁入 record_template_versions，每次更新只追加新版本行，历史版本不可变；
-- 病历表冗余保存模板名称、版本号与当时内容快照，归档病历不随模板更新变化。
-- 开发环境由 GORM AutoMigrate 自动建表，本文件供生产环境受控迁移使用。

CREATE TABLE IF NOT EXISTS record_template_versions (
    id          BIGSERIAL PRIMARY KEY,
    template_id BIGINT      NOT NULL REFERENCES record_templates (id) ON DELETE CASCADE,
    version     INTEGER     NOT NULL,
    content     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT idx_template_version UNIQUE (template_id, version)
);

-- 将既有模板内容迁移为第 1 版（仅当旧 content 列仍存在时执行）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'record_templates' AND column_name = 'content') THEN
        INSERT INTO record_template_versions (template_id, version, content, created_at)
        SELECT id, 1, content, COALESCE(updated_at, NOW())
        FROM record_templates t
        WHERE NOT EXISTS (SELECT 1 FROM record_template_versions v WHERE v.template_id = t.id);
        ALTER TABLE record_templates DROP COLUMN content;
    END IF;
END $$;

ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS template_id BIGINT;
ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS template_name VARCHAR(128);
ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS template_version INTEGER;
ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS template_content TEXT;
