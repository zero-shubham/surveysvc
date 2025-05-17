-- Down migration for users
DROP TABLE users;

-- Down migration for apps
DROP TABLE apps;

-- Down migration for app_groups
DROP TABLE app_groups;

-- DROP FUNCTION update_modified_column;