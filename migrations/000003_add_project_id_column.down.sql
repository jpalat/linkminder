-- Note: SQLite doesn't support DROP COLUMN directly
-- This would require recreating the table without the column
-- For now, we'll leave this as a no-op since it's complex in SQLite
-- In production, you'd typically recreate the table

-- CREATE TABLE bookmarks_new AS SELECT id, timestamp, url, title, description, content, action, shareTo, topic FROM bookmarks;
-- DROP TABLE bookmarks;
-- ALTER TABLE bookmarks_new RENAME TO bookmarks;