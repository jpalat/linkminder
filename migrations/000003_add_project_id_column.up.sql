-- Add project_id foreign key to bookmarks table
ALTER TABLE bookmarks ADD COLUMN project_id INTEGER REFERENCES projects(id);