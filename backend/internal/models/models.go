package models

import "time"

// Author represents a prompt creator
type Author struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Platform    string    `json:"platform"`
	ProfileURL  string    `json:"profile_url"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	PromptCount int       `json:"prompt_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuthorInput represents input for creating/updating an author
type AuthorInput struct {
	Name       string `json:"name" binding:"required"`
	Username   string `json:"username"`
	Platform   string `json:"platform"`
	ProfileURL string `json:"profile_url"`
	AvatarURL  string `json:"avatar_url"`
	Bio        string `json:"bio"`
}

// Prompt represents an AI prompt
type Prompt struct {
	ID          int       `json:"id"`
	CaseNumber  int       `json:"case_number"`
	Title       string    `json:"title"`
	PromptText  string    `json:"prompt_text"`
	SourceURL   string    `json:"source_url"`
	AuthorID    *int      `json:"author_id,omitempty"`
	Author      *Author   `json:"author,omitempty"`
	CategoryID  *int      `json:"category_id,omitempty"`
	Category    *Category `json:"category,omitempty"`
	Thumbnail   string    `json:"thumbnail"`
	IsHidden    bool      `json:"is_hidden"`
	ViewCount   int       `json:"view_count"`
	PromptCount int       `json:"prompt_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []Tag     `json:"tags,omitempty"`
	Images      []string  `json:"images,omitempty"`
}

// PromptInput represents input for creating/updating a prompt
type PromptInput struct {
	CaseNumber  int      `json:"case_number"`
	Title       string   `json:"title" binding:"required"`
	PromptText  string   `json:"prompt_text" binding:"required"`
	SourceURL   string   `json:"source_url"`
	AuthorID    *int     `json:"author_id"`
	CategoryID  *int     `json:"category_id"`
	Thumbnail   string   `json:"thumbnail"`
	IsHidden    bool     `json:"is_hidden"`
	PromptCount int      `json:"prompt_count"`
	Tags        []string `json:"tags"`
	Images      []string `json:"images"`
}

// Tag represents a prompt category/tag
type Tag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	PromptCount int    `json:"prompt_count"`
}

// TagInput represents input for creating/updating a tag
type TagInput struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// Category represents a prompt category for organization
type Category struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	Color        string    `json:"color"`
	ParentID     *int      `json:"parent_id,omitempty"`
	DisplayOrder int       `json:"display_order"`
	PromptCount  int       `json:"prompt_count"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CategoryInput represents input for creating/updating a category
type CategoryInput struct {
	Name         string `json:"name" binding:"required"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	ParentID     *int   `json:"parent_id"`
	DisplayOrder int    `json:"display_order"`
	IsActive     *bool  `json:"is_active"`
}

// AuditLog represents a record of changes made in the system
type AuditLog struct {
	ID          int       `json:"id"`
	Action      string    `json:"action"`      // CREATE, UPDATE, DELETE
	EntityType  string    `json:"entity_type"` // prompt, tag, author, category
	EntityID    int       `json:"entity_id"`
	EntityTitle string    `json:"entity_title"` // Title/name for reference
	OldValues   string    `json:"old_values"`   // JSON string of old values
	NewValues   string    `json:"new_values"`   // JSON string of new values
	UserID      *int      `json:"user_id"`
	Username    string    `json:"username"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
}

// BulkActionInput represents input for bulk operations
type BulkActionInput struct {
	IDs      []int `json:"ids" binding:"required"`
	IsHidden *bool `json:"is_hidden,omitempty"`
}

// PromptImage represents an image associated with a prompt
type PromptImage struct {
	ID           int    `json:"id"`
	PromptID     int    `json:"prompt_id"`
	ImageURL     string `json:"image_url"`
	DisplayOrder int    `json:"display_order"`
}

// Stats represents database statistics
type Stats struct {
	TotalPrompts    int `json:"total_prompts"`
	TotalTags       int `json:"total_tags"`
	TotalImages     int `json:"total_images"`
	TotalCategories int `json:"total_categories"`
}

// User represents an admin user
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // Never expose password hash
	CreatedAt    string `json:"created_at"`
}

// LoginInput represents login request data
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterInput represents registration request data
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	User      struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}
