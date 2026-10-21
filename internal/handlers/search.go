package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"confluent/internal/models"
)

type SearchHandler struct {
	db *sql.DB
}

func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// Search executes a full-text search across documentation page titles and content.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		JSON(w, http.StatusOK, []models.SearchResult{})
		return
	}

	spaceSlug := strings.TrimSpace(r.URL.Query().Get("space"))

	// Format query for SQLite FTS5 prefix matching
	ftsQuery := formatFTSQuery(query)

	var (
		rows *sql.Rows
		err  error
	)

	if spaceSlug != "" {
		sqlQuery := `
			SELECT p.id, p.space_id, s.name, s.slug, p.title,
			       snippet(pages_fts, 1, '<mark class="bg-amber-200 dark:bg-amber-900/60 font-semibold px-0.5 rounded">', '</mark>', '...', 25) AS snippet,
			       bm25(pages_fts) as score
			FROM pages_fts
			JOIN pages p ON pages_fts.rowid = p.id
			JOIN spaces s ON p.space_id = s.id
			WHERE pages_fts MATCH ? AND s.slug = ?
			ORDER BY score ASC
			LIMIT 30
		`
		rows, err = h.db.Query(sqlQuery, ftsQuery, spaceSlug)
	} else {
		sqlQuery := `
			SELECT p.id, p.space_id, s.name, s.slug, p.title,
			       snippet(pages_fts, 1, '<mark class="bg-amber-200 dark:bg-amber-900/60 font-semibold px-0.5 rounded">', '</mark>', '...', 25) AS snippet,
			       bm25(pages_fts) as score
			FROM pages_fts
			JOIN pages p ON pages_fts.rowid = p.id
			JOIN spaces s ON p.space_id = s.id
			WHERE pages_fts MATCH ?
			ORDER BY score ASC
			LIMIT 30
		`
		rows, err = h.db.Query(sqlQuery, ftsQuery)
	}

	// Fallback to LIKE query if FTS returns syntax issue or no rows
	if err != nil {
		likeTerm := "%" + query + "%"
		fallbackQuery := `
			SELECT p.id, p.space_id, s.name, s.slug, p.title,
			       SUBSTR(p.content, 1, 120) || '...' AS snippet,
			       0.0 as score
			FROM pages p
			JOIN spaces s ON p.space_id = s.id
			WHERE p.title LIKE ? OR p.content LIKE ?
			LIMIT 30
		`
		rows, err = h.db.Query(fallbackQuery, likeTerm, likeTerm)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Search query failed")
			return
		}
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var item models.SearchResult
		if err := rows.Scan(
			&item.PageID, &item.SpaceID, &item.SpaceName, &item.SpaceSlug, &item.Title, &item.Snippet, &item.Rank,
		); err != nil {
			continue
		}
		results = append(results, item)
	}

	if results == nil {
		results = []models.SearchResult{}
	}
	JSON(w, http.StatusOK, results)
}

func formatFTSQuery(raw string) string {
	words := strings.Fields(raw)
	if len(words) == 0 {
		return `""`
	}
	var clean []string
	for _, w := range words {
		// Strip special characters that would break FTS syntax
		w = strings.ReplaceAll(w, `"`, "")
		w = strings.ReplaceAll(w, `'`, "")
		w = strings.ReplaceAll(w, `*`, "")
		if len(w) > 0 {
			clean = append(clean, `"`+w+`"*`)
		}
	}
	return strings.Join(clean, " ")
}
