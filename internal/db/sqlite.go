package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  current_scene TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS session_puzzles (
  session_id TEXT NOT NULL,
  scene_id TEXT NOT NULL,
  solved INTEGER NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY(session_id, scene_id)
);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func (s *Store) UpsertSession(ctx context.Context, id string, currentScene string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	const q = `
INSERT INTO sessions (id, current_scene, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  current_scene = excluded.current_scene,
  updated_at = excluded.updated_at;
`
	_, err := s.db.ExecContext(ctx, q, id, currentScene, now, now)
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	return nil
}

func (s *Store) GetSessionScene(ctx context.Context, id string) (string, bool, error) {
	const q = `SELECT current_scene FROM sessions WHERE id = ?`
	var scene string
	if err := s.db.QueryRowContext(ctx, q, id).Scan(&scene); err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, fmt.Errorf("get session: %w", err)
	}
	return scene, true, nil
}

func (s *Store) IsPuzzleSolved(ctx context.Context, sessionID string, sceneID string) (bool, error) {
	const q = `SELECT solved FROM session_puzzles WHERE session_id = ? AND scene_id = ?`
	var solved int
	if err := s.db.QueryRowContext(ctx, q, sessionID, sceneID).Scan(&solved); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("get puzzle state: %w", err)
	}
	return solved == 1, nil
}

func (s *Store) SetPuzzleSolved(ctx context.Context, sessionID string, sceneID string, solved bool) error {
	now := time.Now().UTC().Format(time.RFC3339)
	solvedInt := 0
	if solved {
		solvedInt = 1
	}

	const q = `
INSERT INTO session_puzzles (session_id, scene_id, solved, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(session_id, scene_id) DO UPDATE SET
  solved = excluded.solved,
  updated_at = excluded.updated_at;
`
	if _, err := s.db.ExecContext(ctx, q, sessionID, sceneID, solvedInt, now); err != nil {
		return fmt.Errorf("set puzzle state: %w", err)
	}
	return nil
}
