package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/robertgumeny/doug-plan/internal/ui"
)

// approvedPayload holds content written back to disk after browser approval.
type approvedPayload struct {
	primary   []byte
	secondary []byte
}

// Serve starts an HTTP server on a dynamic port, serves the embedded bundle,
// and blocks until the browser POSTs to /approve. On approval, the updated
// content from the POST body is written back to artifactPath (and
// secondaryPath if non-empty). The URL is printed to out and the default
// browser is opened if possible.
func Serve(artifactPath, secondaryPath, stage string, out io.Writer) error {
	content, err := os.ReadFile(artifactPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading artifact %s: %w", artifactPath, err)
	}

	var secondaryContent []byte
	if secondaryPath != "" {
		secondaryContent, err = os.ReadFile(secondaryPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("reading secondary artifact %s: %w", secondaryPath, err)
		}
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting listener: %w", err)
	}

	url := "http://" + ln.Addr().String()
	approved := make(chan approvedPayload, 1)

	bundleBytes, err := ui.Bundle.ReadFile("bundle.html")
	if err != nil {
		return fmt.Errorf("reading embedded bundle: %w", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(bundleBytes)
	})

	mux.HandleFunc("GET /artifact", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"stage":            stage,
			"content":          string(content),
			"secondaryContent": string(secondaryContent),
		})
	})

	mux.HandleFunc("POST /approve", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Content          string `json:"content"`
			SecondaryContent string `json:"secondaryContent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		approved <- approvedPayload{
			primary:   []byte(body.Content),
			secondary: []byte(body.SecondaryContent),
		}
	})

	srv := &http.Server{Handler: mux}

	writef(out, "Review URL: %s\n", url)
	openBrowser(url)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Serve(ln)
	}()

	select {
	case payload := <-approved:
		_ = srv.Shutdown(context.Background())
		if err := writeFile(artifactPath, payload.primary); err != nil {
			return fmt.Errorf("writing approved artifact: %w", err)
		}
		if secondaryPath != "" && len(payload.secondary) > 0 {
			if err := writeFile(secondaryPath, payload.secondary); err != nil {
				return fmt.Errorf("writing approved secondary artifact: %w", err)
			}
		}
		return nil
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}

func writeFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// openBrowser attempts to open url in the default system browser.
// Failures are silently ignored — the URL is always printed as a fallback.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
