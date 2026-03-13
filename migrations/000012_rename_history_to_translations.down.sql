DROP INDEX IF EXISTS idx_translations_deleted_at;

ALTER TABLE translations
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE translations RENAME TO history;
