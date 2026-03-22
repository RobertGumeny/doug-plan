//go:build integration

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestServe_Endpoints starts a real HTTP server and drives it through the full
// approval flow: GET /, GET /artifact, POST /approve.
func TestServe_Endpoints(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, "ROADMAP.md")
	original := "# Roadmap\n"
	updated := "# Updated Roadmap\n"
	if err := os.WriteFile(artifactPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	urlCh := make(chan string, 1)
	openBrowserFunc = func(url string) {
		urlCh <- url
	}
	defer func() {
		openBrowserFunc = openBrowser
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(artifactPath, "", "Roadmapping", io.Discard)
	}()

	var serverURL string
	select {
	case serverURL = <-urlCh:
	case err := <-errCh:
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("skipping integration test: listener unavailable in this environment: %v", err)
		}
		t.Fatalf("Serve returned before opening browser: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for browser open")
	}

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
	if err := os.WriteFile(artifactPath, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	urlCh := make(chan string, 1)
	openBrowserFunc = func(url string) {
		urlCh <- url
	}
	defer func() {
		openBrowserFunc = openBrowser
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(artifactPath, "", "Discovery", io.Discard)
	}()

	var serverURL string
	select {
	case serverURL = <-urlCh:
	case err := <-errCh:
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("skipping integration test: listener unavailable in this environment: %v", err)
		}
		t.Fatalf("Serve returned before opening browser: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for browser open")
	}

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
