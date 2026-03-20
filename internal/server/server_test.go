package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestServe_Endpoints starts a real HTTP server and drives it through the full
// approval flow: GET /, GET /artifact, POST /approve.
func TestServe_Endpoints(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "ROADMAP.md")
	original := "# Roadmap\n"
	updated := "# Updated Roadmap\n"
	if err := os.WriteFile(artifactPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	out := newURLCapture()
	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(artifactPath, "", "Roadmapping", out)
	}()

	serverURL := out.waitForURL(t, 5*time.Second)

	// GET / must return HTML.
	resp, err := http.Get(serverURL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("GET /: closing response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /: status %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("GET /: Content-Type %q, want text/html", ct)
	}

	// GET /artifact must return JSON with stage and content.
	resp, err = http.Get(serverURL + "/artifact")
	if err != nil {
		t.Fatalf("GET /artifact: %v", err)
	}
	var artifact map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&artifact); err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("GET /artifact: closing response body after decode failure: %v", closeErr)
		}
		t.Fatalf("GET /artifact decode: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("GET /artifact: closing response body: %v", err)
	}
	if artifact["stage"] != "Roadmapping" {
		t.Errorf("artifact stage = %q, want Roadmapping", artifact["stage"])
	}
	if artifact["content"] != original {
		t.Errorf("artifact content = %q, want %q", artifact["content"], original)
	}

	// POST /approve must return 204, update disk, and shut down the server.
	body, _ := json.Marshal(map[string]string{"content": updated, "secondaryContent": ""})
	resp, err = http.Post(serverURL+"/approve", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /approve: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("POST /approve: closing response body: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("POST /approve: status %d, want 204", resp.StatusCode)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Serve did not return after approval")
	}

	got, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != updated {
		t.Errorf("artifact after approval = %q, want %q", string(got), updated)
	}
}

// TestServe_BadApproveBody checks that an invalid POST body returns 400.
func TestServe_BadApproveBody(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "VISION.md")
	if err := os.WriteFile(artifactPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	out := newURLCapture()
	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(artifactPath, "", "Discovery", out)
	}()

	serverURL := out.waitForURL(t, 5*time.Second)

	resp, err := http.Post(serverURL+"/approve", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("POST /approve: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("POST /approve bad body: closing response body: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /approve bad body: status %d, want 400", resp.StatusCode)
	}

	// Server must still be running — send a valid approval to clean up.
	body, _ := json.Marshal(map[string]string{"content": "done", "secondaryContent": ""})
	resp, err = http.Post(serverURL+"/approve", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("cleanup POST /approve: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("cleanup POST /approve: closing response body: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned error after cleanup: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Serve did not return after cleanup approval")
	}
}

// urlCapture is an io.Writer that intercepts the "Review URL: ..." line
// emitted by Serve and makes the URL available via waitForURL.
type urlCapture struct {
	mu   sync.Mutex
	buf  strings.Builder
	ch   chan string
	sent bool
}

func newURLCapture() *urlCapture {
	return &urlCapture{ch: make(chan string, 1)}
}

func (u *urlCapture) Write(p []byte) (int, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.buf.Write(p)
	if !u.sent {
		const prefix = "Review URL: "
		s := u.buf.String()
		if idx := strings.Index(s, prefix); idx >= 0 {
			rest := s[idx+len(prefix):]
			if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
				u.ch <- strings.TrimSpace(rest[:nl])
				u.sent = true
			}
		}
	}
	return len(p), nil
}

func (u *urlCapture) waitForURL(t *testing.T, timeout time.Duration) string {
	t.Helper()
	select {
	case url := <-u.ch:
		return url
	case <-time.After(timeout):
		t.Fatal("timed out waiting for server URL")
		return ""
	}
}
