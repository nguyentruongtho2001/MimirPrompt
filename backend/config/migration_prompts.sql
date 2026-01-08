-- Database setup for MimirPrompt
-- AI Prompt Gallery Database

-- Create database (skip if already created via phpMyAdmin)
-- CREATE DATABASE IF NOT EXISTS mimirprompt_db;
USE mimirprompt_db;

-- Prompt Categories/Tags table
CREATE TABLE IF NOT EXISTS prompt_tags (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  slug VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  prompt_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_slug (slug),
  INDEX idx_prompt_count (prompt_count)
);

-- AI Prompts table
CREATE TABLE IF NOT EXISTS prompts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  case_number INT,                    -- Case number like "案例 857"
  title VARCHAR(500) NOT NULL,
  prompt_text LONGTEXT NOT NULL,      -- The actual AI prompt text
  source_url VARCHAR(1000),           -- Original source URL
  thumbnail VARCHAR(1000),            -- Thumbnail image URL
  is_hidden TINYINT(1) DEFAULT 0,
  view_count INT DEFAULT 0,
  prompt_count INT DEFAULT 1,         -- Number of prompt variations
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_case_number (case_number),
  INDEX idx_is_hidden (is_hidden),
  INDEX idx_view_count (view_count),
  INDEX idx_created_at (created_at),
  FULLTEXT idx_ft_title (title),
  FULLTEXT idx_ft_prompt_text (prompt_text)
);

-- Prompt-Tag relationship (many-to-many)
CREATE TABLE IF NOT EXISTS prompt_tag_relations (
  id INT AUTO_INCREMENT PRIMARY KEY,
  prompt_id INT NOT NULL,
  tag_id INT NOT NULL,
  FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id) REFERENCES prompt_tags(id) ON DELETE CASCADE,
  UNIQUE KEY unique_prompt_tag (prompt_id, tag_id),
  INDEX idx_prompt_id (prompt_id),
  INDEX idx_tag_id (tag_id)
);

-- Prompt Images table (for example images)
CREATE TABLE IF NOT EXISTS prompt_images (
  id INT AUTO_INCREMENT PRIMARY KEY,
  prompt_id INT NOT NULL,
  image_url VARCHAR(1000) NOT NULL,
  display_order INT DEFAULT 0,
  FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
  INDEX idx_prompt_id (prompt_id)
);

-- Admin users table
CREATE TABLE IF NOT EXISTS admin_users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default admin user
-- Password: admin123 (bcrypt hash)
INSERT INTO admin_users (username, password_hash) 
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.wv5vSg.SzCwXzXDdcu')
ON DUPLICATE KEY UPDATE username = username;

-- Insert default tags based on crawled data
INSERT INTO prompt_tags (name, slug, description) VALUES
  ('3D', '3d', 'Thiết kế và render 3D'),
  ('Animal', 'animal', 'Động vật'),
  ('Architecture', 'architecture', 'Kiến trúc'),
  ('Branding', 'branding', 'Thương hiệu'),
  ('Cartoon', 'cartoon', 'Hoạt hình'),
  ('Character', 'character', 'Nhân vật'),
  ('Clay', 'clay', 'Phong cách đất sét'),
  ('Creative', 'creative', 'Sáng tạo'),
  ('Data-Viz', 'data-viz', 'Trực quan hóa dữ liệu'),
  ('Emoji', 'emoji', 'Biểu tượng cảm xúc'),
  ('Fantasy', 'fantasy', 'Giả tưởng'),
  ('Fashion', 'fashion', 'Thời trang'),
  ('Felt', 'felt', 'Phong cách len dạ'),
  ('Food', 'food', 'Đồ ăn'),
  ('Futuristic', 'futuristic', 'Tương lai'),
  ('Gaming', 'gaming', 'Game'),
  ('Illustration', 'illustration', 'Minh họa'),
  ('Infographic', 'infographic', 'Đồ họa thông tin'),
  ('Interior', 'interior', 'Nội thất'),
  ('Landscape', 'landscape', 'Phong cảnh'),
  ('Logo', 'logo', 'Thiết kế logo'),
  ('Minimalist', 'minimalist', 'Tối giản'),
  ('Nature', 'nature', 'Thiên nhiên'),
  ('Neon', 'neon', 'Phong cách neon'),
  ('Paper-Craft', 'paper-craft', 'Nghệ thuật giấy'),
  ('Photography', 'photography', 'Nhiếp ảnh'),
  ('Pixel', 'pixel', 'Pixel art'),
  ('Portrait', 'portrait', 'Chân dung'),
  ('Poster', 'poster', 'Áp phích'),
  ('Product', 'product', 'Sản phẩm'),
  ('Retro', 'retro', 'Phong cách cổ điển'),
  ('Sci-Fi', 'sci-fi', 'Khoa học viễn tưởng'),
  ('Sculpture', 'sculpture', 'Điêu khắc'),
  ('Toy', 'toy', 'Đồ chơi'),
  ('Typography', 'typography', 'Typography'),
  ('UI', 'ui', 'Giao diện người dùng'),
  ('Vehicle', 'vehicle', 'Phương tiện')
ON DUPLICATE KEY UPDATE name = name;
