package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"das-raetsel/internal/db"
	"das-raetsel/internal/story"
)

const sessionCookieName = "das_raetsel_session"

type Server struct {
	store     *db.Store
	story     *story.Story
	tmpl      *template.Template
	staticDir string
}

type ScenePageData struct {
	StoryTitle        string
	SceneID           string
	Scene             story.Scene
	PuzzleSolved      bool
	PuzzleFeedback    string
	PuzzleFeedbackOK  bool
	ChoicesLocked     bool
	PuzzleUnknownType bool
	PuzzleTypeLabel   string
}

func New(store *db.Store, s *story.Story, tmpl *template.Template, staticDir string) *Server {
	return &Server{store: store, story: s, tmpl: tmpl, staticDir: staticDir}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleScene)
	mux.HandleFunc("POST /choice", s.handleChoice)
	mux.HandleFunc("POST /puzzle", s.handlePuzzle)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	return mux
}

func (s *Server) handleScene(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID, _, sceneID, scene, err := s.loadCurrentScene(ctx, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	puzzleSolved, err := s.isScenePuzzleSolved(ctx, sessionID, scene)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	s.renderScene(w, sceneID, scene, puzzleSolved, "", false)
}

func (s *Server) handleChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID, _, sceneID, scene, err := s.loadCurrentScene(ctx, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	puzzleSolved, err := s.isScenePuzzleSolved(ctx, sessionID, scene)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if scene.Puzzle != nil && !puzzleSolved {
		s.renderScene(w, sceneID, scene, false, "Löse zuerst das Rätsel dieser Szene.", false)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	next := strings.TrimSpace(r.FormValue("next"))
	if next == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !isChoiceAllowed(scene, next) {
		http.Error(w, "unknown target scene", http.StatusBadRequest)
		return
	}

	if err := s.store.UpsertSession(ctx, sessionID, next); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handlePuzzle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID, _, sceneID, scene, err := s.loadCurrentScene(ctx, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if scene.Puzzle == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	solved, _ := evaluatePuzzle(scene.Puzzle, r)
	feedback, okFeedback := puzzleFeedback(scene.Puzzle, solved)

	if solved {
		if err := s.store.SetPuzzleSolved(ctx, sessionID, sceneID, true); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	s.renderScene(w, sceneID, scene, solved, feedback, okFeedback)
}

func (s *Server) loadCurrentScene(ctx context.Context, r *http.Request, w http.ResponseWriter) (string, bool, string, story.Scene, error) {
	sessionID, isNew, err := sessionIDFromRequest(r)
	if err != nil {
		return "", false, "", story.Scene{}, fmt.Errorf("session error")
	}
	if isNew {
		setSessionCookie(w, sessionID)
	}

	sceneID, ok, err := s.store.GetSessionScene(ctx, sessionID)
	if err != nil {
		return "", false, "", story.Scene{}, fmt.Errorf("db error")
	}
	if !ok {
		sceneID = s.story.StartScene
		if err := s.store.UpsertSession(ctx, sessionID, sceneID); err != nil {
			return "", false, "", story.Scene{}, fmt.Errorf("db error")
		}
	}

	scene, ok := s.story.Scenes[sceneID]
	if !ok {
		sceneID = s.story.StartScene
		scene, ok = s.story.Scenes[sceneID]
		if !ok {
			return "", false, "", story.Scene{}, fmt.Errorf("scene not found")
		}
		if err := s.store.UpsertSession(ctx, sessionID, sceneID); err != nil {
			return "", false, "", story.Scene{}, fmt.Errorf("db error")
		}
	}

	return sessionID, isNew, sceneID, scene, nil
}

func (s *Server) isScenePuzzleSolved(ctx context.Context, sessionID string, scene story.Scene) (bool, error) {
	if scene.Puzzle == nil {
		return true, nil
	}
	return s.store.IsPuzzleSolved(ctx, sessionID, scene.ID)
}

func (s *Server) renderScene(w http.ResponseWriter, sceneID string, scene story.Scene, puzzleSolved bool, puzzleFeedback string, puzzleFeedbackOK bool) {
	unknownType := scene.Puzzle != nil && !isSupportedPuzzleType(scene.Puzzle.Type)
	choicesLocked := scene.Puzzle != nil && !puzzleSolved && !unknownType

	data := ScenePageData{
		StoryTitle:        s.story.Title,
		SceneID:           sceneID,
		Scene:             scene,
		PuzzleSolved:      puzzleSolved,
		PuzzleFeedback:    puzzleFeedback,
		PuzzleFeedbackOK:  puzzleFeedbackOK,
		ChoicesLocked:     choicesLocked,
		PuzzleUnknownType: unknownType,
		PuzzleTypeLabel:   puzzleTypeLabel(scene.Puzzle),
	}
	if err := s.tmpl.ExecuteTemplate(w, "scene.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func evaluatePuzzle(p *story.PuzzleBlock, r *http.Request) (bool, bool) {
	switch p.Type {
	case "reading_question", "combine_clues":
		answer := normalizeAnswer(r.FormValue("answer"))
		if answer == "" {
			return false, false
		}
		for _, accepted := range p.AcceptedAnswers {
			if answer == normalizeAnswer(accepted) {
				return true, false
			}
		}
		return false, false
	case "fill_in_blank":
		for _, field := range p.Fields {
			value := normalizeAnswer(r.FormValue(field.Name))
			if value == "" || !isInAccepted(value, field.AcceptedAnswers) {
				return false, false
			}
		}
		return len(p.Fields) > 0, false
	case "sentence_order":
		if len(p.AcceptedOrder) == 0 {
			return false, false
		}
		for i := range p.AcceptedOrder {
			value := normalizeAnswer(r.FormValue(fmt.Sprintf("order_%d", i)))
			if value != normalizeAnswer(p.AcceptedOrder[i]) {
				return false, false
			}
		}
		return true, false
	case "dialog_choice":
		selected := strings.TrimSpace(r.FormValue("option"))
		if selected == "" {
			return false, false
		}
		return selected == p.AcceptedOption, false
	default:
		return false, true
	}
}

func isSupportedPuzzleType(kind string) bool {
	switch kind {
	case "reading_question", "combine_clues", "fill_in_blank", "sentence_order", "dialog_choice":
		return true
	default:
		return false
	}
}

func puzzleFeedback(p *story.PuzzleBlock, solved bool) (string, bool) {
	if solved {
		if strings.TrimSpace(p.SuccessText) != "" {
			return p.SuccessText, true
		}
		return "Richtig gelöst.", true
	}
	if strings.TrimSpace(p.FailureText) != "" {
		return p.FailureText, false
	}
	return "Versuche es noch einmal.", false
}

func isInAccepted(value string, accepted []string) bool {
	normalized := make([]string, 0, len(accepted))
	for _, a := range accepted {
		normalized = append(normalized, normalizeAnswer(a))
	}
	return slices.Contains(normalized, value)
}

func normalizeAnswer(value string) string {
	v := strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(
		".", "",
		",", "",
		"!", "",
		"?", "",
		"'", "",
		"\"", "",
		";", "",
		":", "",
	)
	return replacer.Replace(v)
}

func isChoiceAllowed(scene story.Scene, next string) bool {
	for _, choice := range scene.Choices {
		if choice.Next == next {
			return true
		}
	}
	return false
}

func sessionIDFromRequest(r *http.Request) (string, bool, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value, false, nil
	}
	if err != nil && err != http.ErrNoCookie {
		return "", false, err
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", false, fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(b), true, nil
}

func setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
}

func Shutdown(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}

func puzzleTypeLabel(p *story.PuzzleBlock) string {
	if p == nil {
		return ""
	}
	return p.Type
}
