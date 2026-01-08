package repository

import (
	"database/sql"
	"mimirprompt/internal/models"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create inserts a new audit log entry
func (r *AuditRepository) Create(log models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (action, entity_type, entity_id, entity_title, old_values, new_values, user_id, username, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.EntityTitle,
		log.OldValues,
		log.NewValues,
		log.UserID,
		log.Username,
		log.IPAddress,
		log.UserAgent,
	)
	return err
}

// List returns a paginated list of audit logs with optional filters
func (r *AuditRepository) List(page, perPage int, entityType, action string) ([]models.AuditLog, int, error) {
	offset := (page - 1) * perPage

	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if entityType != "" {
		whereClause += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if action != "" {
		whereClause += " AND action = ?"
		args = append(args, action)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM audit_logs " + whereClause
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, action, entity_type, entity_id, entity_title, 
		       COALESCE(old_values, ''), COALESCE(new_values, ''),
		       user_id, COALESCE(username, ''), COALESCE(ip_address, ''), 
		       COALESCE(user_agent, ''), created_at
		FROM audit_logs 
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.EntityTitle,
			&log.OldValues,
			&log.NewValues,
			&log.UserID,
			&log.Username,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetByEntityID returns all audit logs for a specific entity
func (r *AuditRepository) GetByEntityID(entityType string, entityID int) ([]models.AuditLog, error) {
	query := `
		SELECT id, action, entity_type, entity_id, entity_title, 
		       COALESCE(old_values, ''), COALESCE(new_values, ''),
		       user_id, COALESCE(username, ''), COALESCE(ip_address, ''), 
		       COALESCE(user_agent, ''), created_at
		FROM audit_logs 
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.EntityTitle,
			&log.OldValues,
			&log.NewValues,
			&log.UserID,
			&log.Username,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
