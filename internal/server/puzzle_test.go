package server

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"das-raetsel/internal/story"
)

func TestEvaluatePuzzleReadingQuestion(t *testing.T) {
	r := httptest.NewRequest("POST", "/puzzle", strings.NewReader(url.Values{"answer": {"  Lesen! "}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()

	p := &story.PuzzleBlock{Type: "reading_question", AcceptedAnswers: []string{"lesen"}}
	solved, unknown := evaluatePuzzle(p, r)
	if !solved || unknown {
		t.Fatalf("expected solved reading_question, got solved=%v unknown=%v", solved, unknown)
	}
}

func TestEvaluatePuzzleFillInBlank(t *testing.T) {
	form := url.Values{"word1": {"Brücke"}, "word2": {"Tal"}}
	r := httptest.NewRequest("POST", "/puzzle", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()

	p := &story.PuzzleBlock{
		Type: "fill_in_blank",
		Fields: []story.PuzzleField{
			{Name: "word1", AcceptedAnswers: []string{"brücke"}},
			{Name: "word2", AcceptedAnswers: []string{"tal"}},
		},
	}
	solved, unknown := evaluatePuzzle(p, r)
	if !solved || unknown {
		t.Fatalf("expected solved fill_in_blank, got solved=%v unknown=%v", solved, unknown)
	}
}

func TestEvaluatePuzzleSentenceOrder(t *testing.T) {
	form := url.Values{
		"order_0": {"unter der Erde"},
		"order_1": {"liegt die Spur"},
	}
	r := httptest.NewRequest("POST", "/puzzle", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()

	p := &story.PuzzleBlock{
		Type:          "sentence_order",
		AcceptedOrder: []string{"unter der Erde", "liegt die Spur"},
	}
	solved, unknown := evaluatePuzzle(p, r)
	if !solved || unknown {
		t.Fatalf("expected solved sentence_order, got solved=%v unknown=%v", solved, unknown)
	}
}

func TestEvaluatePuzzleDialogChoice(t *testing.T) {
	r := httptest.NewRequest("POST", "/puzzle", strings.NewReader(url.Values{"option": {"kind"}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()

	p := &story.PuzzleBlock{Type: "dialog_choice", AcceptedOption: "kind"}
	solved, unknown := evaluatePuzzle(p, r)
	if !solved || unknown {
		t.Fatalf("expected solved dialog_choice, got solved=%v unknown=%v", solved, unknown)
	}
}

func TestEvaluatePuzzleUnknownType(t *testing.T) {
	r := httptest.NewRequest("POST", "/puzzle", nil)
	p := &story.PuzzleBlock{Type: "something_new"}
	solved, unknown := evaluatePuzzle(p, r)
	if solved || !unknown {
		t.Fatalf("expected unknown puzzle type, got solved=%v unknown=%v", solved, unknown)
	}
}
