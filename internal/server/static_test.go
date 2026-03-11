package server

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"das-raetsel/internal/story"
)

func TestRoutes_ServesStaticFiles(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "img"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(tmp, "img", "test.txt")
	if err := os.WriteFile(path, []byte("hello-static"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	srv := New(nil, &story.Story{}, template.Must(template.New("x").Parse(`{{define "scene.html"}}ok{{end}}`)), tmp)
	r := httptest.NewRequest("GET", "/static/img/test.txt", nil)
	w := httptest.NewRecorder()
	srv.Routes().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "hello-static" {
		t.Fatalf("unexpected body: %q", got)
	}
}
