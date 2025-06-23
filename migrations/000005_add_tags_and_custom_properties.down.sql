-- Remove tags and custom_properties columns from bookmarks table

ALTER TABLE bookmarks DROP COLUMN tags;
ALTER TABLE bookmarks DROP COLUMN custom_properties;