package repository

import (
	"database/sql"
	"mimirprompt/internal/models"
	"strings"
)

// TagRepository handles database operations for tags
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// List returns all tags ordered by prompt count
func (r *TagRepository) List() ([]models.Tag, error) {
	rows, err := r.db.Query(`
		SELECT id, name, slug, COALESCE(description,''), prompt_count 
		FROM prompt_tags 
		ORDER BY prompt_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.PromptCount)
		tags = append(tags, t)
	}

	return tags, nil
}

// GetByID returns a single tag by ID
func (r *TagRepository) GetByID(id string) (*models.Tag, error) {
	var t models.Tag
	err := r.db.QueryRow(`
		SELECT id, name, slug, COALESCE(description,''), prompt_count 
		FROM prompt_tags WHERE id = ?
	`, id).Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.PromptCount)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

// Create creates a new tag
func (r *TagRepository) Create(input models.TagInput) (int64, error) {
	slug := input.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))
	}

	result, err := r.db.Exec(`
		INSERT INTO prompt_tags (name, slug, description) VALUES (?, ?, ?)
	`, input.Name, slug, input.Description)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update updates an existing tag
func (r *TagRepository) Update(id string, input models.TagInput) error {
	updates := []string{}
	args := []interface{}{}

	if input.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, input.Name)
	}
	if input.Slug != "" {
		updates = append(updates, "slug = ?")
		args = append(args, input.Slug)
	}
	if input.Description != "" {
		updates = append(updates, "description = ?")
		args = append(args, input.Description)
	}

	if len(updates) == 0 {
		return nil
	}

	args = append(args, id)
	query := "UPDATE prompt_tags SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err := r.db.Exec(query, args...)
	return err
}

// Delete deletes a tag
func (r *TagRepository) Delete(id string) (int64, error) {
	// Delete tag relations first
	r.db.Exec("DELETE FROM prompt_tag_relations WHERE tag_id = ?", id)

	result, err := r.db.Exec("DELETE FROM prompt_tags WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns database statistics
func (r *TagRepository) GetStats(db *sql.DB) (*models.Stats, error) {
	var stats models.Stats

	db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&stats.TotalPrompts)
	db.QueryRow("SELECT COUNT(*) FROM prompt_tags").Scan(&stats.TotalTags)
	db.QueryRow("SELECT COUNT(*) FROM prompt_images").Scan(&stats.TotalImages)

	return &stats, nil
}
