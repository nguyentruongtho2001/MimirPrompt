package repository

import (
	"database/sql"
	"strings"

	"mimirprompt/internal/models"
)

// CategoryRepository handles category database operations
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List returns all categories ordered by display_order
func (r *CategoryRepository) List() ([]models.Category, error) {
	query := `
		SELECT id, name, slug, COALESCE(description, ''), COALESCE(icon, ''), 
		       COALESCE(color, ''), parent_id, display_order, prompt_count, 
		       is_active, created_at, updated_at
		FROM categories
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&cat.Icon,
			&cat.Color,
			&cat.ParentID,
			&cat.DisplayOrder,
			&cat.PromptCount,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// GetByID returns a category by its ID
func (r *CategoryRepository) GetByID(id int) (*models.Category, error) {
	query := `
		SELECT id, name, slug, COALESCE(description, ''), COALESCE(icon, ''), 
		       COALESCE(color, ''), parent_id, display_order, prompt_count, 
		       is_active, created_at, updated_at
		FROM categories
		WHERE id = ?
	`

	var cat models.Category
	err := r.db.QueryRow(query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Slug,
		&cat.Description,
		&cat.Icon,
		&cat.Color,
		&cat.ParentID,
		&cat.DisplayOrder,
		&cat.PromptCount,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &cat, nil
}

// GetBySlug returns a category by its slug
func (r *CategoryRepository) GetBySlug(slug string) (*models.Category, error) {
	query := `
		SELECT id, name, slug, COALESCE(description, ''), COALESCE(icon, ''), 
		       COALESCE(color, ''), parent_id, display_order, prompt_count, 
		       is_active, created_at, updated_at
		FROM categories
		WHERE slug = ?
	`

	var cat models.Category
	err := r.db.QueryRow(query, slug).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Slug,
		&cat.Description,
		&cat.Icon,
		&cat.Color,
		&cat.ParentID,
		&cat.DisplayOrder,
		&cat.PromptCount,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &cat, nil
}

// Create inserts a new category
func (r *CategoryRepository) Create(input models.CategoryInput) (int64, error) {
	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `
		INSERT INTO categories (name, slug, description, icon, color, parent_id, display_order, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		input.Name,
		slug,
		input.Description,
		input.Icon,
		input.Color,
		input.ParentID,
		input.DisplayOrder,
		isActive,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update updates an existing category
func (r *CategoryRepository) Update(id int, input models.CategoryInput) error {
	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Name)
	}

	query := `
		UPDATE categories 
		SET name = ?, slug = ?, description = ?, icon = ?, color = ?, 
		    parent_id = ?, display_order = ?
		WHERE id = ?
	`

	args := []interface{}{
		input.Name,
		slug,
		input.Description,
		input.Icon,
		input.Color,
		input.ParentID,
		input.DisplayOrder,
		id,
	}

	// Handle is_active if provided
	if input.IsActive != nil {
		query = `
			UPDATE categories 
			SET name = ?, slug = ?, description = ?, icon = ?, color = ?, 
			    parent_id = ?, display_order = ?, is_active = ?
			WHERE id = ?
		`
		args = []interface{}{
			input.Name,
			slug,
			input.Description,
			input.Icon,
			input.Color,
			input.ParentID,
			input.DisplayOrder,
			*input.IsActive,
			id,
		}
	}

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete removes a category by ID
func (r *CategoryRepository) Delete(id int) error {
	// First, set category_id to NULL for all prompts in this category
	_, err := r.db.Exec("UPDATE prompts SET category_id = NULL WHERE category_id = ?", id)
	if err != nil {
		return err
	}

	// Then delete the category
	_, err = r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

// UpdatePromptCount updates the prompt_count for a category
func (r *CategoryRepository) UpdatePromptCount(categoryID int) error {
	query := `
		UPDATE categories 
		SET prompt_count = (SELECT COUNT(*) FROM prompts WHERE category_id = ?)
		WHERE id = ?
	`
	_, err := r.db.Exec(query, categoryID, categoryID)
	return err
}

// GetCount returns total number of categories
func (r *CategoryRepository) GetCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	return count, err
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	var result strings.Builder
	for _, c := range slug {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}
