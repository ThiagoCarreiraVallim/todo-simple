DROP INDEX IF EXISTS lists_user_id_idx;

ALTER TABLE lists DROP COLUMN IF EXISTS user_id;
