-- Add indexes on frequently filtered columns for better query performance

-- Index on action column (used for filtering by bookmark type)
CREATE INDEX IF NOT EXISTS idx_bookmarks_action ON bookmarks(action);

-- Index on topic column (used for legacy topic filtering)
CREATE INDEX IF NOT EXISTS idx_bookmarks_topic ON bookmarks(topic);

-- Index on deleted column (used to filter out soft-deleted items)
CREATE INDEX IF NOT EXISTS idx_bookmarks_deleted ON bookmarks(deleted);

-- Index on timestamp column (used for sorting and date filtering)
CREATE INDEX IF NOT EXISTS idx_bookmarks_timestamp ON bookmarks(timestamp DESC);

-- Index on project_id column (used for project-bookmark relationships)
CREATE INDEX IF NOT EXISTS idx_bookmarks_project_id ON bookmarks(project_id);

-- Composite index for common query pattern: active working bookmarks
CREATE INDEX IF NOT EXISTS idx_bookmarks_action_deleted ON bookmarks(action, deleted) WHERE deleted = 0;

-- Composite index for project queries
CREATE INDEX IF NOT EXISTS idx_bookmarks_project_deleted ON bookmarks(project_id, deleted) WHERE deleted = 0;
