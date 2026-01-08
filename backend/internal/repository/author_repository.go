package repository

import (
	"database/sql"

	"mimirprompt/internal/models"
)

type AuthorRepository struct {
	db *sql.DB
}

func NewAuthorRepository(db *sql.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// List returns all authors with pagination
func (r *AuthorRepository) List(page, perPage int) ([]models.Author, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM authors").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get authors
	query := `
		SELECT id, name, COALESCE(username, ''), COALESCE(platform, 'twitter'), 
		       COALESCE(profile_url, ''), COALESCE(avatar_url, ''), COALESCE(bio, ''),
		       prompt_count, created_at, updated_at
		FROM authors
		ORDER BY prompt_count DESC, name ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var authors []models.Author
	for rows.Next() {
		var author models.Author
		err := rows.Scan(
			&author.ID, &author.Name, &author.Username, &author.Platform,
			&author.ProfileURL, &author.AvatarURL, &author.Bio,
			&author.PromptCount, &author.CreatedAt, &author.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		authors = append(authors, author)
	}

	return authors, total, nil
}

// GetByID returns an author by ID
func (r *AuthorRepository) GetByID(id int) (*models.Author, error) {
	query := `
		SELECT id, name, COALESCE(username, ''), COALESCE(platform, 'twitter'), 
		       COALESCE(profile_url, ''), COALESCE(avatar_url, ''), COALESCE(bio, ''),
		       prompt_count, created_at, updated_at
		FROM authors
		WHERE id = ?
	`

	var author models.Author
	err := r.db.QueryRow(query, id).Scan(
		&author.ID, &author.Name, &author.Username, &author.Platform,
		&author.ProfileURL, &author.AvatarURL, &author.Bio,
		&author.PromptCount, &author.CreatedAt, &author.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &author, nil
}

// Create creates a new author
func (r *AuthorRepository) Create(input models.AuthorInput) (int, error) {
	result, err := r.db.Exec(`
		INSERT INTO authors (name, username, platform, profile_url, avatar_url, bio)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.Name, input.Username, input.Platform, input.ProfileURL, input.AvatarURL, input.Bio)

	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return int(id), nil
}

// Update updates an existing author
func (r *AuthorRepository) Update(id int, input models.AuthorInput) error {
	_, err := r.db.Exec(`
		UPDATE authors 
		SET name = ?, username = ?, platform = ?, profile_url = ?, avatar_url = ?, bio = ?
		WHERE id = ?
	`, input.Name, input.Username, input.Platform, input.ProfileURL, input.AvatarURL, input.Bio, id)

	return err
}

// Delete deletes an author
func (r *AuthorRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM authors WHERE id = ?", id)
	return err
}

// Search searches authors by name or username
func (r *AuthorRepository) Search(query string) ([]models.Author, error) {
	searchQuery := `
		SELECT id, name, COALESCE(username, ''), COALESCE(platform, 'twitter'), 
		       COALESCE(profile_url, ''), COALESCE(avatar_url, ''), COALESCE(bio, ''),
		       prompt_count, created_at, updated_at
		FROM authors
		WHERE name LIKE ? OR username LIKE ?
		ORDER BY prompt_count DESC
		LIMIT 50
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(searchQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []models.Author
	for rows.Next() {
		var author models.Author
		err := rows.Scan(
			&author.ID, &author.Name, &author.Username, &author.Platform,
			&author.ProfileURL, &author.AvatarURL, &author.Bio,
			&author.PromptCount, &author.CreatedAt, &author.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// UpdatePromptCount recalculates prompt_count for an author
func (r *AuthorRepository) UpdatePromptCount(authorID int) error {
	_, err := r.db.Exec(`
		UPDATE authors 
		SET prompt_count = (SELECT COUNT(*) FROM prompts WHERE author_id = ?)
		WHERE id = ?
	`, authorID, authorID)
	return err
}
