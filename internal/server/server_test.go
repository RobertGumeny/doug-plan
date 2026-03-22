package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteFile verifies the atomic write helper.
func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	data := "# Hello\n"

	if err := writeFile(target, []byte(data)); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != data {
		t.Errorf("got %q, want %q", string(got), data)
	}
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Error(".tmp file still present after rename")
	}
}

func TestNewMux_RootServesHTML(t *testing.T) {
	approved := make(chan approvedPayload, 1)
	handler := newMux("Roadmapping", []byte("# Roadmap\n"), nil, []byte("<html>ok</html>"), approved)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want %q", got, "text/html; charset=utf-8")
	}
}

func TestNewMux_ArtifactReturnsJSON(t *testing.T) {
	approved := make(chan approvedPayload, 1)
	handler := newMux("Roadmapping", []byte("# Roadmap\n"), []byte("secondary"), []byte("<html>ok</html>"), approved)

	req := httptest.NewRequest(http.MethodGet, "/artifact", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var artifact map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&artifact); err != nil {
		t.Fatalf("decode artifact: %v", err)
	}
	if artifact["stage"] != "Roadmapping" {
		t.Errorf("stage = %q, want %q", artifact["stage"], "Roadmapping")
	}
	if artifact["content"] != "# Roadmap\n" {
		t.Errorf("content = %q, want %q", artifact["content"], "# Roadmap\n")
	}
	if artifact["secondaryContent"] != "secondary" {
		t.Errorf("secondaryContent = %q, want %q", artifact["secondaryContent"], "secondary")
	}
}

func TestNewMux_ApproveRejectsBadJSON(t *testing.T) {
	approved := make(chan approvedPayload, 1)
	handler := newMux("Discovery", []byte("content"), nil, []byte("<html>ok</html>"), approved)

	req := httptest.NewRequest(http.MethodPost, "/approve", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	select {
	case payload := <-approved:
		t.Fatalf("unexpected approval payload: %+v", payload)
	default:
	}
}

func TestNewMux_ApproveQueuesPayload(t *testing.T) {
	approved := make(chan approvedPayload, 1)
	handler := newMux("PRD", []byte("old"), []byte("old-secondary"), []byte("<html>ok</html>"), approved)

	req := httptest.NewRequest(http.MethodPost, "/approve", strings.NewReader(`{"content":"new","secondaryContent":"tasks"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	select {
	case payload := <-approved:
		if string(payload.primary) != "new" {
			t.Errorf("primary = %q, want %q", string(payload.primary), "new")
		}
		if string(payload.secondary) != "tasks" {
			t.Errorf("secondary = %q, want %q", string(payload.secondary), "tasks")
		}
	default:
		t.Fatal("expected approval payload")
	}
}
