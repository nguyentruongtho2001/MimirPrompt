-- Admin Dashboard Improvements Migration
-- Adds: Audit Logs, Categories, Bulk Operations support

USE mimirprompt_db;

-- ==========================================
-- 1. AUDIT LOGS TABLE
-- Tracks all changes to prompts, tags, authors, categories
-- ==========================================
CREATE TABLE IF NOT EXISTS audit_logs (
  id INT AUTO_INCREMENT PRIMARY KEY,
  action VARCHAR(50) NOT NULL COMMENT 'CREATE, UPDATE, DELETE',
  entity_type VARCHAR(50) NOT NULL COMMENT 'prompt, tag, author, category',
  entity_id INT NOT NULL,
  entity_title VARCHAR(500) COMMENT 'Title/name for reference',
  old_values JSON COMMENT 'Previous values before change',
  new_values JSON COMMENT 'New values after change',
  user_id INT COMMENT 'Admin user who made the change',
  username VARCHAR(50) COMMENT 'Username for display',
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_action (action),
  INDEX idx_entity_type (entity_type),
  INDEX idx_entity_id (entity_id),
  INDEX idx_user_id (user_id),
  INDEX idx_created_at (created_at)
);

-- ==========================================
-- 2. CATEGORIES TABLE
-- Organize prompts into categories (separate from tags)
-- ==========================================
CREATE TABLE IF NOT EXISTS categories (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  slug VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  icon VARCHAR(50) COMMENT 'Icon name or emoji',
  color VARCHAR(20) COMMENT 'Hex color code',
  parent_id INT DEFAULT NULL COMMENT 'For nested categories',
  display_order INT DEFAULT 0,
  prompt_count INT DEFAULT 0,
  is_active TINYINT(1) DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL,
  INDEX idx_slug (slug),
  INDEX idx_parent_id (parent_id),
  INDEX idx_display_order (display_order),
  INDEX idx_is_active (is_active)
);

-- ==========================================
-- 3. ADD CATEGORY TO PROMPTS
-- ==========================================
ALTER TABLE prompts ADD COLUMN IF NOT EXISTS category_id INT DEFAULT NULL;
ALTER TABLE prompts ADD INDEX IF NOT EXISTS idx_category_id (category_id);
-- Note: Foreign key may fail if column already exists, ignore error
-- ALTER TABLE prompts ADD FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- ==========================================
-- 4. INSERT DEFAULT CATEGORIES
-- ==========================================
INSERT INTO categories (name, slug, description, icon, color, display_order) VALUES
  ('AI Art', 'ai-art', 'Tạo hình ảnh nghệ thuật với AI', '🎨', '#8B5CF6', 1),
  ('Photography', 'photography', 'Prompt cho ảnh chân thực', '📷', '#3B82F6', 2),
  ('Design', 'design', 'Thiết kế đồ họa và UI/UX', '🎯', '#F59E0B', 3),
  ('Character', 'character', 'Thiết kế nhân vật', '👤', '#10B981', 4),
  ('Landscape', 'landscape', 'Phong cảnh và môi trường', '🌄', '#6366F1', 5),
  ('Product', 'product', 'Chụp ảnh sản phẩm', '📦', '#EC4899', 6),
  ('Abstract', 'abstract', 'Nghệ thuật trừu tượng', '🔮', '#8B5CF6', 7),
  ('Other', 'other', 'Các prompt khác', '📁', '#6B7280', 99)
ON DUPLICATE KEY UPDATE name = name;

-- ==========================================
-- 5. UPDATE PROMPT COUNT FOR CATEGORIES (trigger alternative)
-- Run this after importing prompts to a category
-- ==========================================
-- UPDATE categories c SET prompt_count = (
--   SELECT COUNT(*) FROM prompts p WHERE p.category_id = c.id
-- );
