package handler

import (
	"encoding/json"
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// --- Plugin Tests ---

func TestListPlugins(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "GET", "/api/v1/admin/plugins", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var plugins []interface{}
	if err := json.NewDecoder(rec.Body).Decode(&plugins); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected empty list, got %d plugins", len(plugins))
	}
}

func TestListPlugins_NonAdmin(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "GET", "/api/v1/admin/plugins", makeToken("user"), nil)
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// --- Research Tools Tests ---

func TestResearchTools_CRUD(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// List — empty
	rec := doRequest(mux, "GET", "/api/v1/admin/research_tools", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("list: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tools []model.ResearchTool
	json.NewDecoder(rec.Body).Decode(&tools)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}

	// Create
	enabled := true
	rec = doRequest(mux, "POST", "/api/v1/admin/research_tools", tok, map[string]interface{}{
		"name": "Google Scholar", "url_template": "https://scholar.google.com/scholar?q={query}", "enabled": enabled,
	})
	if rec.Code != 201 {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created model.ResearchTool
	json.NewDecoder(rec.Body).Decode(&created)
	if created.Name != "Google Scholar" {
		t.Errorf("expected name 'Google Scholar', got %q", created.Name)
	}
	if created.URLTemplate != "https://scholar.google.com/scholar?q={query}" {
		t.Errorf("unexpected url_template: %q", created.URLTemplate)
	}
	if !created.Enabled {
		t.Error("expected enabled=true")
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}

	// List — one tool
	rec = doRequest(mux, "GET", "/api/v1/admin/research_tools", tok, nil)
	json.NewDecoder(rec.Body).Decode(&tools)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	// Update
	rec = doRequest(mux, "PUT", "/api/v1/admin/research_tools/1", tok, map[string]interface{}{
		"name": "Google Scholar v2", "enabled": false,
	})
	if rec.Code != 200 {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updated model.ResearchTool
	json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Name != "Google Scholar v2" {
		t.Errorf("expected updated name, got %q", updated.Name)
	}
	if updated.Enabled {
		t.Error("expected enabled=false after update")
	}

	// Delete
	rec = doRequest(mux, "DELETE", "/api/v1/admin/research_tools/1", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("delete: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List — empty again
	rec = doRequest(mux, "GET", "/api/v1/admin/research_tools", tok, nil)
	json.NewDecoder(rec.Body).Decode(&tools)
	if len(tools) != 0 {
		t.Errorf("expected 0 tools after delete, got %d", len(tools))
	}
}

func TestResearchTools_CreateValidation(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Missing name
	rec := doRequest(mux, "POST", "/api/v1/admin/research_tools", tok, map[string]interface{}{
		"url_template": "https://example.com/{query}",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for missing name, got %d", rec.Code)
	}

	// Missing url_template
	rec = doRequest(mux, "POST", "/api/v1/admin/research_tools", tok, map[string]interface{}{
		"name": "Test",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for missing url_template, got %d", rec.Code)
	}
}

func TestResearchTools_NonAdmin(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("user")

	rec := doRequest(mux, "GET", "/api/v1/admin/research_tools", tok, nil)
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	rec = doRequest(mux, "POST", "/api/v1/admin/research_tools", tok, map[string]interface{}{
		"name": "Test", "url_template": "https://example.com/{query}",
	})
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestResearchTools_DeleteNotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "DELETE", "/api/v1/admin/research_tools/999", tok, nil)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResearchTools_UpdateNotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "PUT", "/api/v1/admin/research_tools/999", tok, map[string]interface{}{
		"name": "Nope",
	})
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
