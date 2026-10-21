package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"confluent/internal/auth"
	"confluent/internal/models"
)

type SpacesHandler struct {
	db *sql.DB
}

func NewSpacesHandler(db *sql.DB) *SpacesHandler {
	return &SpacesHandler{db: db}
}

type spaceListItem struct {
	models.Space
	PageCount int `json:"page_count"`
}

// List returns all spaces with total page count.
func (h *SpacesHandler) List(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT s.id, s.slug, s.name, s.description, s.icon, s.created_by, s.created_at, s.updated_at,
		       COUNT(p.id) as page_count
		FROM spaces s
		LEFT JOIN pages p ON s.id = p.space_id
		GROUP BY s.id
		ORDER BY s.name ASC
	`
	rows, err := h.db.Query(query)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to query spaces")
		return
	}
	defer rows.Close()

	var list []spaceListItem
	for rows.Next() {
		var item spaceListItem
		if err := rows.Scan(
			&item.ID, &item.Slug, &item.Name, &item.Description, &item.Icon,
			&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt, &item.PageCount,
		); err != nil {
			Error(w, http.StatusInternalServerError, "Failed to scan space row")
			return
		}
		list = append(list, item)
	}

	if list == nil {
		list = []spaceListItem{}
	}
	JSON(w, http.StatusOK, list)
}

// Get returns details for a specific space slug.
func (h *SpacesHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var s models.Space
	err := h.db.QueryRow(`
		SELECT id, slug, name, description, icon, created_by, created_at, updated_at
		FROM spaces WHERE slug = ?
	`, slug).Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Icon, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "Space not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	JSON(w, http.StatusOK, s)
}

type createSpaceRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Create creates a new space (Admin only).
func (h *SpacesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	var req createSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		Error(w, http.StatusBadRequest, "Space name is required")
		return
	}

	slug := req.Slug
	if slug == "" {
		slug = Slugify(req.Name)
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	res, err := h.db.Exec(`
		INSERT INTO spaces (slug, name, description, icon, created_by)
		VALUES (?, ?, ?, ?, ?)
	`, slug, req.Name, req.Description, req.Icon, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			Error(w, http.StatusConflict, "A space with this identifier already exists")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to create space")
		return
	}

	id, _ := res.LastInsertId()
	JSON(w, http.StatusCreated, map[string]any{
		"id":          id,
		"slug":        slug,
		"name":        req.Name,
		"description": req.Description,
		"icon":        req.Icon,
	})
}

// Delete removes a space and all its nested pages (Admin only).
func (h *SpacesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	res, err := h.db.Exec("DELETE FROM spaces WHERE slug = ?", slug)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to delete space")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		Error(w, http.StatusNotFound, "Space not found")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Space deleted successfully"})
}
