package server

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"das-raetsel/internal/db"
	"das-raetsel/internal/story"
	_ "modernc.org/sqlite"
)

func TestHandleScene_FallbackToStartWhenStoredSceneMissing(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	s := &story.Story{
		Title:      "Test",
		StartScene: "start",
		Scenes: map[string]story.Scene{
			"start": {ID: "start", Title: "Start", Text: "Hallo"},
		},
	}
	tmpl := template.Must(template.New("scene.html").Parse(`{{define "scene.html"}}{{.SceneID}}{{end}}`))
	srv := New(store, s, tmpl, tmp)

	if err := seedSession(dbPath, "sid1", "legacy_scene"); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "sid1"})
	w := httptest.NewRecorder()

	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "start" {
		t.Fatalf("expected rendered start scene, got %q", got)
	}

	sceneID, ok, err := store.GetSessionScene(req.Context(), "sid1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if !ok || sceneID != "start" {
		t.Fatalf("expected session scene reset to start, got ok=%v scene=%q", ok, sceneID)
	}
}

func seedSession(dbPath string, sessionID string, sceneID string) error {
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	_, err = dbConn.Exec(`INSERT INTO sessions (id, current_scene, created_at, updated_at) VALUES (?, ?, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`, sessionID, sceneID)
	return err
}
