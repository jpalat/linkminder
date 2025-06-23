-- Add tags and custom_properties columns to bookmarks table
-- Tags will be stored as JSON array of strings
-- Custom properties will be stored as JSON object with string keys and values

ALTER TABLE bookmarks ADD COLUMN tags TEXT DEFAULT '[]';
ALTER TABLE bookmarks ADD COLUMN custom_properties TEXT DEFAULT '{}';