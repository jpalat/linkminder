-- Migrate existing topics to projects table and update bookmarks
INSERT OR IGNORE INTO projects (name, description, status, created_at, updated_at)
SELECT DISTINCT 
    topic,
    'Migrated from topic: ' || topic,
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM bookmarks 
WHERE topic IS NOT NULL AND topic != '' AND topic NOT IN (SELECT name FROM projects);

-- Update bookmarks to reference project_id
UPDATE bookmarks 
SET project_id = (
    SELECT p.id 
    FROM projects p 
    WHERE p.name = bookmarks.topic
)
WHERE topic IS NOT NULL AND topic != '' AND project_id IS NULL;