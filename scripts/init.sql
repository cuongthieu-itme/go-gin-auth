-- Create database if not exists
CREATE DATABASE IF NOT EXISTS authdb;

-- Use the database
USE authdb;

-- Grant permissions
GRANT ALL PRIVILEGES ON authdb.* TO 'appuser'@'%';
FLUSH PRIVILEGES;
