package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"confluent/internal/auth"
	"confluent/internal/database"
	"confluent/internal/handlers"
	"confluent/internal/models"
)

func setupTestDB(t *testing.T) (*handlers.AuthHandler, *handlers.SpacesHandler, *handlers.PagesHandler, *handlers.SearchHandler, string) {
	t.Helper()
	tempDB := t.TempDir() + "/test.db"
	db, err := database.Open(tempDB, "admin123")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tempDB)
	})

	jwtSecret := "test-secret-12345678"
	return handlers.NewAuthHandler(db, jwtSecret),
		handlers.NewSpacesHandler(db),
		handlers.NewPagesHandler(db),
		handlers.NewSearchHandler(db),
		jwtSecret
}

func TestAuthAndPagesWorkflow(t *testing.T) {
	authH, spacesH, pagesH, searchH, _ := setupTestDB(t)

	// 1. Test Login with seeded default admin
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "admin@confluent.local",
		"password": "admin123",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	w := httptest.NewRecorder()
	authH.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK login, got %d: %s", w.Code, w.Body.String())
	}

	var authResp struct {
		Token string       `json:"token"`
		User  *models.User `json:"user"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &authResp)
	if authResp.Token == "" || authResp.User.Role != models.RoleAdmin {
		t.Fatalf("invalid login response: %+v", authResp)
	}

	// 2. Test Listing Spaces
	req = httptest.NewRequest("GET", "/api/v1/spaces", nil)
	w = httptest.NewRecorder()
	spacesH.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK spaces list, got %d", w.Code)
	}

	// 3. Test Creating a Page
	pageBody, _ := json.Marshal(map[string]any{
		"title":   "API Architecture Overview",
		"content": "This document covers our REST API guidelines, rate limits, and authentication patterns.",
	})
	req = httptest.NewRequest("POST", "/api/v1/spaces/engineering/pages", bytes.NewReader(pageBody))
	req.SetPathValue("slug", "engineering")
	req = req.WithContext(auth.ContextWithUser(req.Context(), authResp.User))

	w = httptest.NewRecorder()
	pagesH.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created page, got %d: %s", w.Code, w.Body.String())
	}

	var createResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	pageID := int64(createResp["id"].(float64))

	// 4. Test Search with SQLite FTS5
	req = httptest.NewRequest("GET", "/api/v1/search?q=guidelines", nil)
	w = httptest.NewRecorder()
	searchH.Search(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 search, got %d: %s", w.Code, w.Body.String())
	}

	var searchResults []models.SearchResult
	_ = json.Unmarshal(w.Body.Bytes(), &searchResults)
	if len(searchResults) == 0 {
		t.Fatalf("expected at least 1 search result for 'guidelines', got 0")
	}
	if searchResults[0].PageID != pageID {
		t.Fatalf("expected matched page ID %d, got %d", pageID, searchResults[0].PageID)
	}
}
