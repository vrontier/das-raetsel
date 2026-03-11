package story

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadV1(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "v1.yaml")
	content := `title: "Test"
start_scene: "start"
scenes:
  start:
    title: "Start"
    text: "Hallo"
    choices:
      - label: "Weiter"
        next: "end"
  end:
    title: "Ende"
    text: "Fertig"
    choices: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test yaml: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load(v1) failed: %v", err)
	}
	if s.Title != "Test" {
		t.Fatalf("unexpected title: %q", s.Title)
	}
	if s.StartScene != "start" {
		t.Fatalf("unexpected start_scene: %q", s.StartScene)
	}
	if _, ok := s.Scenes["end"]; !ok {
		t.Fatalf("missing scene end")
	}
}

func TestLoadV2(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "v2.yaml")
	content := `story:
  id: "s1"
  title: "Test V2"
  language: "de"
  start_scene: "start"
scenes:
  - id: "start"
    title: "Start"
    text: "Hallo"
    image: "/static/img/start.png"
    hint: "Lies genau"
    choices:
      - label: "Weiter"
        next: "end"
    puzzle:
      type: "reading_question"
      prompt: "Was tust du zuerst?"
      accepted_answers: ["lesen"]
  - id: "end"
    title: "Ende"
    text: "Fertig"
    solution_fragment: "X"
    choices: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test yaml: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load(v2) failed: %v", err)
	}
	if s.ID != "s1" {
		t.Fatalf("unexpected id: %q", s.ID)
	}
	if s.Language != "de" {
		t.Fatalf("unexpected language: %q", s.Language)
	}
	if s.Scenes["start"].Puzzle == nil {
		t.Fatalf("expected puzzle on scene start")
	}
	if s.Scenes["end"].SolutionFragment != "X" {
		t.Fatalf("unexpected solution fragment")
	}
}
