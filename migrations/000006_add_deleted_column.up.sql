-- Add deleted column to bookmarks table for soft delete functionality
ALTER TABLE bookmarks ADD COLUMN deleted BOOLEAN DEFAULT FALSE;