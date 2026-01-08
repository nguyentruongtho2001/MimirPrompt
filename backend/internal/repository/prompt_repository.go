package repository

import (
	"database/sql"
	"mimirprompt/internal/models"
)

// PromptRepository handles database operations for prompts
type PromptRepository struct {
	db *sql.DB
}

// NewPromptRepository creates a new prompt repository
func NewPromptRepository(db *sql.DB) *PromptRepository {
	return &PromptRepository{db: db}
}

// List returns a paginated list of prompts
func (r *PromptRepository) List(page, perPage int, showHidden bool) ([]models.Prompt, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM prompts"
	if !showHidden {
		countQuery += " WHERE is_hidden = 0"
	}
	if err := r.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get prompts with author info
	query := `SELECT p.id, p.case_number, p.title, p.prompt_text, COALESCE(p.source_url,''), 
			  COALESCE(p.thumbnail,''), p.is_hidden, p.view_count, p.prompt_count, p.created_at,
			  COALESCE(p.author_id, 0), COALESCE(a.name, ''), COALESCE(a.username, '')
			  FROM prompts p
			  LEFT JOIN authors a ON p.author_id = a.id`
	if !showHidden {
		query += " WHERE p.is_hidden = 0"
	}
	query += " ORDER BY p.case_number DESC LIMIT ? OFFSET ?"

	rows, err := r.db.Query(query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		var authorID int
		var authorName, authorUsername string
		if err := rows.Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
			&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt,
			&authorID, &authorName, &authorUsername); err != nil {
			return nil, 0, err
		}

		if authorID > 0 {
			p.Author = &models.Author{
				ID:       authorID,
				Name:     authorName,
				Username: authorUsername,
			}
		}
		prompts = append(prompts, p)
	}

	return prompts, total, nil
}

// GetByID returns a single prompt by ID with its images and tags
func (r *PromptRepository) GetByID(id string) (*models.Prompt, error) {
	var p models.Prompt
	err := r.db.QueryRow(`
		SELECT id, case_number, title, prompt_text, COALESCE(source_url,''), 
		COALESCE(thumbnail,''), is_hidden, view_count, prompt_count, created_at
		FROM prompts WHERE id = ?
	`, id).Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
		&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Get images
	imgRows, err := r.db.Query("SELECT image_url FROM prompt_images WHERE prompt_id = ? ORDER BY display_order", id)
	if err != nil {
		return nil, err
	}
	defer imgRows.Close()
	for imgRows.Next() {
		var url string
		if err := imgRows.Scan(&url); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, url)
	}

	// Get tags
	tagRows, err := r.db.Query(`
		SELECT t.id, t.name, t.slug, COALESCE(t.description,''), t.prompt_count
		FROM prompt_tags t
		JOIN prompt_tag_relations ptr ON t.id = ptr.tag_id
		WHERE ptr.prompt_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var t models.Tag
		if err := tagRows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.PromptCount); err != nil {
			return nil, err
		}
		p.Tags = append(p.Tags, t)
	}

	// Increment view count (fire and forget - non-critical operation)
	go r.db.Exec("UPDATE prompts SET view_count = view_count + 1 WHERE id = ?", id)

	return &p, nil
}

// Create creates a new prompt
func (r *PromptRepository) Create(input models.PromptInput) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO prompts (case_number, title, prompt_text, source_url, author_id, thumbnail, is_hidden, prompt_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.CaseNumber, input.Title, input.PromptText, input.SourceURL,
		input.AuthorID, input.Thumbnail, input.IsHidden, input.PromptCount)

	if err != nil {
		return 0, err
	}

	promptID, _ := result.LastInsertId()

	// Add images
	for i, img := range input.Images {
		r.db.Exec("INSERT INTO prompt_images (prompt_id, image_url, display_order) VALUES (?, ?, ?)",
			promptID, img, i)
	}

	// Add tags (lookup by slug)
	for _, tagSlug := range input.Tags {
		r.db.Exec(`INSERT INTO prompt_tag_relations (prompt_id, tag_id) 
			SELECT ?, id FROM prompt_tags WHERE slug = ?`, promptID, tagSlug)
	}

	return promptID, nil
}

// Update updates an existing prompt
func (r *PromptRepository) Update(id string, input models.PromptInput) error {
	_, err := r.db.Exec(`
		UPDATE prompts SET title = ?, prompt_text = ?, source_url = ?, 
		author_id = ?, thumbnail = ?, is_hidden = ?, prompt_count = ?
		WHERE id = ?
	`, input.Title, input.PromptText, input.SourceURL, input.AuthorID, input.Thumbnail,
		input.IsHidden, input.PromptCount, id)

	if err != nil {
		return err
	}

	// Update images if provided
	if len(input.Images) > 0 {
		r.db.Exec("DELETE FROM prompt_images WHERE prompt_id = ?", id)
		for i, img := range input.Images {
			r.db.Exec("INSERT INTO prompt_images (prompt_id, image_url, display_order) VALUES (?, ?, ?)",
				id, img, i)
		}
	}

	// Update tags - delete old and add new
	if len(input.Tags) > 0 {
		r.db.Exec("DELETE FROM prompt_tag_relations WHERE prompt_id = ?", id)
		for _, tagSlug := range input.Tags {
			r.db.Exec(`INSERT INTO prompt_tag_relations (prompt_id, tag_id) 
				SELECT ?, id FROM prompt_tags WHERE slug = ?`, id, tagSlug)
		}
	}

	return nil
}

// Delete deletes a prompt
func (r *PromptRepository) Delete(id string) error {
	// Delete related data first
	r.db.Exec("DELETE FROM prompt_images WHERE prompt_id = ?", id)
	r.db.Exec("DELETE FROM prompt_tag_relations WHERE prompt_id = ?", id)

	_, err := r.db.Exec("DELETE FROM prompts WHERE id = ?", id)
	return err
}

// Search searches prompts by query
func (r *PromptRepository) Search(query string, page, perPage int) ([]models.Prompt, int, error) {
	offset := (page - 1) * perPage
	searchQuery := "%" + query + "%"

	// Get total count
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM prompts WHERE title LIKE ? OR prompt_text LIKE ?`,
		searchQuery, searchQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get prompts
	rows, err := r.db.Query(`
		SELECT id, case_number, title, prompt_text, COALESCE(source_url,''), 
		COALESCE(thumbnail,''), is_hidden, view_count, prompt_count, created_at
		FROM prompts 
		WHERE title LIKE ? OR prompt_text LIKE ?
		ORDER BY case_number DESC
		LIMIT ? OFFSET ?
	`, searchQuery, searchQuery, perPage, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		if err := rows.Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
			&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		prompts = append(prompts, p)
	}

	return prompts, total, nil
}

// ListByTag returns prompts filtered by tag slug
func (r *PromptRepository) ListByTag(slug string, page, perPage int) ([]models.Prompt, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	if err := r.db.QueryRow(`
		SELECT COUNT(*) FROM prompts p
		JOIN prompt_tag_relations ptr ON p.id = ptr.prompt_id
		JOIN prompt_tags t ON ptr.tag_id = t.id
		WHERE t.slug = ? AND p.is_hidden = 0
	`, slug).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get prompts
	rows, err := r.db.Query(`
		SELECT p.id, p.case_number, p.title, p.prompt_text, COALESCE(p.source_url,''), 
		COALESCE(p.thumbnail,''), p.is_hidden, p.view_count, p.prompt_count, p.created_at
		FROM prompts p
		JOIN prompt_tag_relations ptr ON p.id = ptr.prompt_id
		JOIN prompt_tags t ON ptr.tag_id = t.id
		WHERE t.slug = ? AND p.is_hidden = 0
		ORDER BY p.case_number DESC
		LIMIT ? OFFSET ?
	`, slug, perPage, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		if err := rows.Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
			&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		prompts = append(prompts, p)
	}

	return prompts, total, nil
}

// BulkDelete deletes multiple prompts by their IDs
func (r *PromptRepository) BulkDelete(ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	// Build placeholder string
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	// Delete related data first
	r.db.Exec("DELETE FROM prompt_images WHERE prompt_id IN ("+placeholders+")", args...)
	r.db.Exec("DELETE FROM prompt_tag_relations WHERE prompt_id IN ("+placeholders+")", args...)

	// Delete prompts
	result, err := r.db.Exec("DELETE FROM prompts WHERE id IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// BulkUpdateVisibility updates is_hidden for multiple prompts
func (r *PromptRepository) BulkUpdateVisibility(ids []int, isHidden bool) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	// Build placeholder string
	placeholders := ""
	args := make([]interface{}, len(ids)+1)
	args[0] = isHidden
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i+1] = id
	}

	result, err := r.db.Exec("UPDATE prompts SET is_hidden = ? WHERE id IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetByIDSimple returns a prompt by ID without incrementing view count (for audit logging)
func (r *PromptRepository) GetByIDSimple(id int) (*models.Prompt, error) {
	var p models.Prompt
	err := r.db.QueryRow(`
		SELECT id, case_number, title, prompt_text, COALESCE(source_url,''), 
		COALESCE(thumbnail,''), is_hidden, view_count, prompt_count, created_at
		FROM prompts WHERE id = ?
	`, id).Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
		&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetMultipleByIDs returns multiple prompts by their IDs
func (r *PromptRepository) GetMultipleByIDs(ids []int) ([]models.Prompt, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	rows, err := r.db.Query(`
		SELECT id, case_number, title, prompt_text, COALESCE(source_url,''), 
		COALESCE(thumbnail,''), is_hidden, view_count, prompt_count, created_at
		FROM prompts WHERE id IN (`+placeholders+`)
	`, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		if err := rows.Scan(&p.ID, &p.CaseNumber, &p.Title, &p.PromptText, &p.SourceURL,
			&p.Thumbnail, &p.IsHidden, &p.ViewCount, &p.PromptCount, &p.CreatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}

	return prompts, nil
}
