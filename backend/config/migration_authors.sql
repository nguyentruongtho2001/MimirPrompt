-- Migration: Add Authors table
-- Run this in MySQL to add author support

USE mimirprompt_db;

-- Authors table to store creator information
CREATE TABLE IF NOT EXISTS authors (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  username VARCHAR(100),                  -- Social media username (e.g., @username)
  platform VARCHAR(50) DEFAULT 'twitter', -- Platform: twitter, instagram, etc.
  profile_url VARCHAR(500),               -- Link to author's profile
  avatar_url VARCHAR(500),                -- Author's avatar image
  bio TEXT,                               -- Short bio/description
  prompt_count INT DEFAULT 0,             -- Number of prompts by this author
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_username (username),
  INDEX idx_platform (platform),
  INDEX idx_prompt_count (prompt_count)
);

-- Add author_id to prompts table
ALTER TABLE prompts 
ADD COLUMN author_id INT DEFAULT NULL AFTER source_url,
ADD INDEX idx_author_id (author_id),
ADD FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL;
