package models

import "time"

// Role defines user access levels.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// User represents a team member account.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Space represents a top-level documentation category (e.g., Engineering, HR).
type Space struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Page represents a documentation article within a space.
type Page struct {
	ID         int64     `json:"id"`
	SpaceID    int64     `json:"space_id"`
	ParentID   *int64    `json:"parent_id"` // nil indicates top-level page in the space
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	OrderIndex int       `json:"order_index"`
	CreatedBy  int64     `json:"created_by"`
	UpdatedBy  int64     `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PageTreeNode represents a page with its nested children for the sidebar tree.
type PageTreeNode struct {
	ID         int64           `json:"id"`
	SpaceID    int64           `json:"space_id"`
	ParentID   *int64          `json:"parent_id"`
	Slug       string          `json:"slug"`
	Title      string          `json:"title"`
	OrderIndex int             `json:"order_index"`
	Children   []*PageTreeNode `json:"children"`
}

// PageVersion tracks historical revisions of a page.
type PageVersion struct {
	ID            int64     `json:"id"`
	PageID        int64     `json:"page_id"`
	VersionNum    int       `json:"version_num"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ChangeSummary string    `json:"change_summary"`
	EditedBy      int64     `json:"edited_by"`
	EditorName    string    `json:"editor_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Comment allows team discussions on specific documentation pages.
type Comment struct {
	ID         int64     `json:"id"`
	PageID     int64     `json:"page_id"`
	UserID     int64     `json:"user_id"`
	AuthorName string    `json:"author_name,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SearchResult represents a matched page from full-text search.
type SearchResult struct {
	PageID    int64  `json:"page_id"`
	SpaceID   int64  `json:"space_id"`
	SpaceName string `json:"space_name"`
	SpaceSlug string `json:"space_slug"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	Rank      float64 `json:"rank,omitempty"`
}
