package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"confluent/internal/auth"
	"confluent/internal/models"
)

type PagesHandler struct {
	db *sql.DB
}

func NewPagesHandler(db *sql.DB) *PagesHandler {
	return &PagesHandler{db: db}
}

// Tree builds and returns the hierarchical page tree for a given space.
func (h *PagesHandler) Tree(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var spaceID int64
	err := h.db.QueryRow("SELECT id FROM spaces WHERE slug = ?", slug).Scan(&spaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "Space not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	rows, err := h.db.Query(`
		SELECT id, space_id, parent_id, slug, title, order_index
		FROM pages
		WHERE space_id = ?
		ORDER BY order_index ASC, title ASC
	`, spaceID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to query pages")
		return
	}
	defer rows.Close()

	nodes := make(map[int64]*models.PageTreeNode)
	var roots []*models.PageTreeNode
	type rawNode struct {
		id, spaceID int64
		parentID    *int64
		slug, title string
		orderIndex  int
	}
	var rawList []rawNode

	for rows.Next() {
		var n rawNode
		if err := rows.Scan(&n.id, &n.spaceID, &n.parentID, &n.slug, &n.title, &n.orderIndex); err != nil {
			Error(w, http.StatusInternalServerError, "Failed to scan page")
			return
		}
		treeNode := &models.PageTreeNode{
			ID:         n.id,
			SpaceID:    n.spaceID,
			ParentID:   n.parentID,
			Slug:       n.slug,
			Title:      n.title,
			OrderIndex: n.orderIndex,
			Children:   []*models.PageTreeNode{},
		}
		nodes[n.id] = treeNode
		rawList = append(rawList, n)
	}

	// Link children to parents
	for _, raw := range rawList {
		node := nodes[raw.id]
		if raw.parentID == nil {
			roots = append(roots, node)
		} else if parent, exists := nodes[*raw.parentID]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node) // fallback if parent missing
		}
	}

	if roots == nil {
		roots = []*models.PageTreeNode{}
	}
	JSON(w, http.StatusOK, roots)
}

// PageDetailResponse includes page content, breadcrumb path, and author names.
type PageDetailResponse struct {
	models.Page
	SpaceSlug    string            `json:"space_slug"`
	SpaceName    string            `json:"space_name"`
	AuthorName   string            `json:"author_name"`
	UpdaterName  string            `json:"updater_name"`
	Breadcrumbs  []BreadcrumbItem  `json:"breadcrumbs"`
	VersionCount int               `json:"version_count"`
}

type BreadcrumbItem struct {
	ID    int64  `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// Get returns full details for a page, including its ancestral hierarchy.
func (h *PagesHandler) Get(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	var resp PageDetailResponse
	var parentID sql.NullInt64

	query := `
		SELECT p.id, p.space_id, p.parent_id, p.slug, p.title, p.content, p.order_index,
		       p.created_by, p.updated_by, p.created_at, p.updated_at,
		       s.slug, s.name,
		       u1.full_name, u2.full_name,
		       (SELECT COUNT(*) FROM page_versions WHERE page_id = p.id) as version_count
		FROM pages p
		JOIN spaces s ON p.space_id = s.id
		JOIN users u1 ON p.created_by = u1.id
		JOIN users u2 ON p.updated_by = u2.id
		WHERE p.id = ?
	`
	err = h.db.QueryRow(query, pageID).Scan(
		&resp.ID, &resp.SpaceID, &parentID, &resp.Slug, &resp.Title, &resp.Content, &resp.OrderIndex,
		&resp.CreatedBy, &resp.UpdatedBy, &resp.CreatedAt, &resp.UpdatedAt,
		&resp.SpaceSlug, &resp.SpaceName,
		&resp.AuthorName, &resp.UpdaterName,
		&resp.VersionCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "Page not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	if parentID.Valid {
		resp.ParentID = &parentID.Int64
	}

	// Fetch breadcrumb hierarchy recursively
	resp.Breadcrumbs = h.fetchBreadcrumbs(resp.ParentID)
	JSON(w, http.StatusOK, resp)
}

func (h *PagesHandler) fetchBreadcrumbs(parentID *int64) []BreadcrumbItem {
	var crumbs []BreadcrumbItem
	curr := parentID
	for curr != nil {
		var item BreadcrumbItem
		var nextParent sql.NullInt64
		err := h.db.QueryRow(`
			SELECT id, slug, title, parent_id
			FROM pages WHERE id = ?
		`, *curr).Scan(&item.ID, &item.Slug, &item.Title, &nextParent)
		if err != nil {
			break
		}
		crumbs = append([]BreadcrumbItem{item}, crumbs...) // prepend
		if nextParent.Valid {
			curr = &nextParent.Int64
		} else {
			curr = nil
		}
	}
	return crumbs
}

type createPageRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	ParentID   *int64 `json:"parent_id"`
	OrderIndex int    `json:"order_index"`
}

// Create creates a new documentation page and logs version 1.
func (h *PagesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	slug := r.PathValue("slug")

	var spaceID int64
	if err := h.db.QueryRow("SELECT id FROM spaces WHERE slug = ?", slug).Scan(&spaceID); err != nil {
		Error(w, http.StatusNotFound, "Space not found")
		return
	}

	var req createPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		Error(w, http.StatusBadRequest, "Page title is required")
		return
	}

	pageSlug := Slugify(req.Title)
	if pageSlug == "" {
		pageSlug = fmt.Sprintf("page-%d", time.Now().Unix())
	}

	// Ensure unique slug within space
	var existsCount int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM pages WHERE space_id = ? AND slug = ?", spaceID, pageSlug).Scan(&existsCount)
	if existsCount > 0 {
		pageSlug = fmt.Sprintf("%s-%d", pageSlug, time.Now().Unix()%10000)
	}

	res, err := h.db.Exec(`
		INSERT INTO pages (space_id, parent_id, slug, title, content, order_index, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, spaceID, req.ParentID, pageSlug, req.Title, req.Content, req.OrderIndex, user.ID, user.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create page")
		return
	}

	pageID, _ := res.LastInsertId()

	// Create initial version record
	_, _ = h.db.Exec(`
		INSERT INTO page_versions (page_id, version_num, title, content, change_summary, edited_by)
		VALUES (?, 1, ?, ?, 'Initial page creation', ?)
	`, pageID, req.Title, req.Content, user.ID)

	JSON(w, http.StatusCreated, map[string]any{
		"id":        pageID,
		"space_id":  spaceID,
		"parent_id": req.ParentID,
		"slug":      pageSlug,
		"title":     req.Title,
	})
}

type updatePageRequest struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	ParentID      *int64 `json:"parent_id"`
	OrderIndex    *int   `json:"order_index"`
	ChangeSummary string `json:"change_summary"`
}

// Update saves edits and snapshots a new revision version.
func (h *PagesHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	var req updatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		Error(w, http.StatusBadRequest, "Page title cannot be empty")
		return
	}

	// Fetch current max version number
	var currentMaxVersion int
	err = h.db.QueryRow(`
		SELECT COALESCE(MAX(version_num), 0) FROM page_versions WHERE page_id = ?
	`, pageID).Scan(&currentMaxVersion)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch version count")
		return
	}

	nextVersion := currentMaxVersion + 1
	changeSummary := strings.TrimSpace(req.ChangeSummary)
	if changeSummary == "" {
		changeSummary = fmt.Sprintf("Updated to revision %d", nextVersion)
	}

	// Update page row
	_, err = h.db.Exec(`
		UPDATE pages
		SET title = ?, content = ?, parent_id = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Title, req.Content, req.ParentID, user.ID, pageID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update page")
		return
	}

	// Insert into version history
	_, err = h.db.Exec(`
		INSERT INTO page_versions (page_id, version_num, title, content, change_summary, edited_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, pageID, nextVersion, req.Title, req.Content, changeSummary, user.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to record page version")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"message":     "Page updated successfully",
		"version_num": nextVersion,
	})
}

// Delete removes a page and cascades its children and versions.
func (h *PagesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	res, err := h.db.Exec("DELETE FROM pages WHERE id = ?", pageID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to delete page")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		Error(w, http.StatusNotFound, "Page not found")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Page deleted"})
}

// Versions lists all revision history records for a page.
func (h *PagesHandler) Versions(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	rows, err := h.db.Query(`
		SELECT v.id, v.page_id, v.version_num, v.title, v.change_summary, v.edited_by, u.full_name, v.created_at
		FROM page_versions v
		JOIN users u ON v.edited_by = u.id
		WHERE v.page_id = ?
		ORDER BY v.version_num DESC
	`, pageID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to query versions")
		return
	}
	defer rows.Close()

	var versions []models.PageVersion
	for rows.Next() {
		var v models.PageVersion
		if err := rows.Scan(
			&v.ID, &v.PageID, &v.VersionNum, &v.Title, &v.ChangeSummary, &v.EditedBy, &v.EditorName, &v.CreatedAt,
		); err != nil {
			Error(w, http.StatusInternalServerError, "Failed to scan version")
			return
		}
		versions = append(versions, v)
	}

	if versions == nil {
		versions = []models.PageVersion{}
	}
	JSON(w, http.StatusOK, versions)
}

// GetVersion returns content for a specific historical revision.
func (h *PagesHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid page ID")
		return
	}
	versionNum, err := strconv.Atoi(r.PathValue("num"))
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	var v models.PageVersion
	err = h.db.QueryRow(`
		SELECT v.id, v.page_id, v.version_num, v.title, v.content, v.change_summary, v.edited_by, u.full_name, v.created_at
		FROM page_versions v
		JOIN users u ON v.edited_by = u.id
		WHERE v.page_id = ? AND v.version_num = ?
	`, pageID, versionNum).Scan(
		&v.ID, &v.PageID, &v.VersionNum, &v.Title, &v.Content, &v.ChangeSummary, &v.EditedBy, &v.EditorName, &v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "Version not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	JSON(w, http.StatusOK, v)
}
