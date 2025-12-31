-- Remove performance indexes

DROP INDEX IF EXISTS idx_bookmarks_project_deleted;
DROP INDEX IF EXISTS idx_bookmarks_action_deleted;
DROP INDEX IF EXISTS idx_bookmarks_project_id;
DROP INDEX IF EXISTS idx_bookmarks_timestamp;
DROP INDEX IF EXISTS idx_bookmarks_deleted;
DROP INDEX IF EXISTS idx_bookmarks_topic;
DROP INDEX IF EXISTS idx_bookmarks_action;
