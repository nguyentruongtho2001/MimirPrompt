-- MimirPrompt Performance Indexes
-- Run this script to add indexes for common queries

-- Performance indexes for prompts table
CREATE INDEX IF NOT EXISTS idx_prompts_view_count ON prompts(view_count DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_created_at ON prompts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_is_hidden ON prompts(is_hidden);
CREATE INDEX IF NOT EXISTS idx_prompts_category_id ON prompts(category_id);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_prompts_hidden_views ON prompts(is_hidden, view_count DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_hidden_created ON prompts(is_hidden, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_category_hidden ON prompts(category_id, is_hidden);

-- Full-text index for search (if not exists)
-- ALTER TABLE prompts ADD FULLTEXT INDEX ft_prompts_search (title, prompt_text);

-- Analyze tables to update statistics
ANALYZE TABLE prompts;
ANALYZE TABLE tags;
ANALYZE TABLE categories;
ANALYZE TABLE authors;

-- Show created indexes
SHOW INDEX FROM prompts;
