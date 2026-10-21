package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Open initializes and migrates the SQLite database.
func Open(dbPath, defaultAdminPass string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	// SQLite connection pool configuration for WAL mode
	db.SetMaxOpenConns(1) // modernc sqlite single-writer pattern ensures strict serialization
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Essential SQLite Pragmas for concurrency and integrity
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	if err := seedDefaults(db, defaultAdminPass); err != nil {
		return nil, fmt.Errorf("seeding defaults: %w", err)
	}

	slog.Info("SQLite database initialized successfully", "path", dbPath)
	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE COLLATE NOCASE,
		password_hash TEXT NOT NULL,
		full_name TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('admin', 'editor', 'viewer')) DEFAULT 'editor',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS spaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE COLLATE NOCASE,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		icon TEXT NOT NULL DEFAULT 'folder',
		created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		space_id INTEGER NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
		parent_id INTEGER REFERENCES pages(id) ON DELETE CASCADE,
		slug TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL DEFAULT '',
		order_index INTEGER NOT NULL DEFAULT 0,
		created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		updated_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(space_id, slug)
	);

	CREATE INDEX IF NOT EXISTS idx_pages_space_parent ON pages(space_id, parent_id, order_index);

	CREATE TABLE IF NOT EXISTS page_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		page_id INTEGER NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
		version_num INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		change_summary TEXT NOT NULL DEFAULT '',
		edited_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(page_id, version_num)
	);

	CREATE INDEX IF NOT EXISTS idx_versions_page ON page_versions(page_id, version_num DESC);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		page_id INTEGER NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_comments_page ON comments(page_id, created_at ASC);

	-- SQLite FTS5 Full Text Search Virtual Table
	CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(
		title,
		content,
		content='pages',
		content_rowid='id'
	);

	-- Triggers to automatically sync FTS5 index on pages changes
	CREATE TRIGGER IF NOT EXISTS pages_ai AFTER INSERT ON pages BEGIN
		INSERT INTO pages_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS pages_ad AFTER DELETE ON pages BEGIN
		INSERT INTO pages_fts(pages_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
	END;

	CREATE TRIGGER IF NOT EXISTS pages_au AFTER UPDATE ON pages BEGIN
		INSERT INTO pages_fts(pages_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
		INSERT INTO pages_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
	END;
	`

	_, err := db.Exec(schema)
	return err
}

func seedDefaults(db *sql.DB, defaultAdminPass string) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}

	res, err := db.Exec(`
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ('admin@confluent.local', ?, 'System Administrator', 'admin')
	`, string(hash))
	if err != nil {
		return fmt.Errorf("inserting default admin: %w", err)
	}

	adminID, _ := res.LastInsertId()

	// Seed Initial Default Spaces
	spaces := []struct {
		slug, name, desc, icon string
	}{
		{"engineering", "Engineering", "Technical architecture, API docs, and runbooks", "code-xml"},
		{"product", "Product & Design", "Roadmaps, specs, personas, and design tokens", "layout"},
		{"handbook", "Company Handbook", "Company culture, benefits, onboarding, and policies", "book-open"},
	}

	for _, s := range spaces {
		sRes, err := db.Exec(`
			INSERT INTO spaces (slug, name, description, icon, created_by)
			VALUES (?, ?, ?, ?, ?)
		`, s.slug, s.name, s.desc, s.icon, adminID)
		if err != nil {
			return fmt.Errorf("seeding space %s: %w", s.slug, err)
		}
		spaceID, _ := sRes.LastInsertId()

		// Seed initial welcome page for handbook
		if s.slug == "handbook" {
			pRes, err := db.Exec(`
				INSERT INTO pages (space_id, parent_id, slug, title, content, order_index, created_by, updated_by)
				VALUES (?, NULL, 'welcome', 'Welcome to Confluent', ?, 0, ?, ?)
			`, spaceID, initialWelcomeDoc(), adminID, adminID)
			if err == nil {
				pageID, _ := pRes.LastInsertId()
				_, _ = db.Exec(`
					INSERT INTO page_versions (page_id, version_num, title, content, change_summary, edited_by)
					VALUES (?, 1, 'Welcome to Confluent', ?, 'Initial documentation creation', ?)
				`, pageID, initialWelcomeDoc(), adminID)
			}
		}
	}

	slog.Info("Default administrator and starter spaces seeded successfully")
	return nil
}

func initialWelcomeDoc() string {
	return `# Welcome to Confluent

Confluent is your team's lightweight internal knowledge base, built for speed, simplicity, and clarity.

## Key Features

- **Spaces**: Organize documentation into clean domain spaces like Engineering, Product, and HR.
- **Hierarchical Trees**: Nest pages infinitely to build structured knowledge trees.
- **Fast Full-Text Search**: Powered by SQLite FTS5 for instantaneous page lookups.
- **Version History**: Review historical diffs and track who changed what.
- **Inline Comments**: Collaborate and discuss directly on documentation pages.

> *“Simplicity is prerequisite for reliability.”* — Edsger W. Dijkstra
`
}
