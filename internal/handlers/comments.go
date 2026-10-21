package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"confluent/internal/auth"
	"confluent/internal/models"
)

type CommentsHandler struct {
	db *sql.DB
}

func NewCommentsHandler(db *sql.DB) *CommentsHandler {
	return &CommentsHandler{db: db}
}

// List returns all comments for a specific page.
func (h *CommentsHandler) List(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	rows, err := h.db.Query(`
		SELECT c.id, c.page_id, c.user_id, u.full_name, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.page_id = ?
		ORDER BY c.created_at ASC
	`, pageID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to query comments")
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PageID, &c.UserID, &c.AuthorName, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
			Error(w, http.StatusInternalServerError, "Failed to scan comment")
			return
		}
		comments = append(comments, c)
	}

	if comments == nil {
		comments = []models.Comment{}
	}
	JSON(w, http.StatusOK, comments)
}

type createCommentRequest struct {
	Content string `json:"content"`
}

// Create posts a new comment to a page.
func (h *CommentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		Error(w, http.StatusBadRequest, "Comment content cannot be empty")
		return
	}

	res, err := h.db.Exec(`
		INSERT INTO comments (page_id, user_id, content)
		VALUES (?, ?, ?)
	`, pageID, user.ID, req.Content)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to save comment")
		return
	}

	id, _ := res.LastInsertId()
	JSON(w, http.StatusCreated, map[string]any{
		"id":          id,
		"page_id":     pageID,
		"user_id":     user.ID,
		"author_name": user.FullName,
		"content":     req.Content,
	})
}

// Delete removes a comment (by author or admin).
func (h *CommentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	commentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var authorID int64
	err = h.db.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "Comment not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	if user.Role != models.RoleAdmin && user.ID != authorID {
		Error(w, http.StatusForbidden, "Cannot delete comments written by others")
		return
	}

	_, err = h.db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Comment deleted"})
}
