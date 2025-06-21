-- Rollback: Clear project_id references and remove migrated projects
UPDATE bookmarks SET project_id = NULL WHERE project_id IS NOT NULL;
DELETE FROM projects WHERE description LIKE 'Migrated from topic:%';