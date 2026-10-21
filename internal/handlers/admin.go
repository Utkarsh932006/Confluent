package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"confluent/internal/models"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// ListUsers returns all registered users with their roles.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, email, full_name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to query users")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			Error(w, http.StatusInternalServerError, "Failed to scan user")
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}
	JSON(w, http.StatusOK, users)
}

type updateRoleRequest struct {
	Role models.Role `json:"role"`
}

// UpdateRole alters a user's role (Admin only).
func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Role != models.RoleAdmin && req.Role != models.RoleEditor && req.Role != models.RoleViewer {
		Error(w, http.StatusBadRequest, "Invalid role. Allowed: admin, editor, viewer")
		return
	}

	_, err = h.db.Exec(`
		UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, req.Role, targetID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "User role updated successfully"})
}

// Stats returns system-wide metrics.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]int)

	queries := map[string]string{
		"users":    "SELECT COUNT(*) FROM users",
		"spaces":   "SELECT COUNT(*) FROM spaces",
		"pages":    "SELECT COUNT(*) FROM pages",
		"versions": "SELECT COUNT(*) FROM page_versions",
		"comments": "SELECT COUNT(*) FROM comments",
	}

	for key, q := range queries {
		var cnt int
		_ = h.db.QueryRow(q).Scan(&cnt)
		stats[key] = cnt
	}

	JSON(w, http.StatusOK, stats)
}
