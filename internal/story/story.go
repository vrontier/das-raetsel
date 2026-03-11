package story

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Story is the normalized in-memory story model used by the server.
type Story struct {
	ID         string
	Title      string
	Language   string
	StartScene string
	Scenes     map[string]Scene
}

type Scene struct {
	ID               string       `yaml:"-"`
	Title            string       `yaml:"title"`
	Text             string       `yaml:"text"`
	Image            string       `yaml:"image"`
	Hint             string       `yaml:"hint"`
	SolutionFragment string       `yaml:"solution_fragment"`
	Choices          []Choice     `yaml:"choices"`
	Puzzle           *PuzzleBlock `yaml:"puzzle"`
}

type Choice struct {
	Label string `yaml:"label"`
	Next  string `yaml:"next"`
}

type PuzzleBlock struct {
	Type            string         `yaml:"type"`
	Prompt          string         `yaml:"prompt"`
	AcceptedAnswers []string       `yaml:"accepted_answers"`
	Fields          []PuzzleField  `yaml:"fields"`
	Items           []string       `yaml:"items"`
	AcceptedOrder   []string       `yaml:"accepted_order"`
	Options         []PuzzleOption `yaml:"options"`
	AcceptedOption  string         `yaml:"accepted_option"`
	SuccessText     string         `yaml:"success_text"`
	FailureText     string         `yaml:"failure_text"`
}

type PuzzleField struct {
	Name            string   `yaml:"name"`
	AcceptedAnswers []string `yaml:"accepted_answers"`
}

type PuzzleOption struct {
	ID   string `yaml:"id"`
	Text string `yaml:"text"`
}

type storyFileV1 struct {
	Title      string           `yaml:"title"`
	StartScene string           `yaml:"start_scene"`
	Scenes     map[string]Scene `yaml:"scenes"`
}

type storyFileV2 struct {
	Story struct {
		ID         string `yaml:"id"`
		Title      string `yaml:"title"`
		Language   string `yaml:"language"`
		StartScene string `yaml:"start_scene"`
	} `yaml:"story"`
	Scenes []sceneWithID `yaml:"scenes"`
}

type sceneWithID struct {
	ID    string `yaml:"id"`
	Scene `yaml:",inline"`
}

func Load(path string) (*Story, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read story file: %w", err)
	}

	// Prefer v2 shape if `story:` exists.
	var probe map[string]any
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("parse story yaml: %w", err)
	}

	if _, isV2 := probe["story"]; isV2 {
		s, err := parseV2(data)
		if err != nil {
			return nil, err
		}
		if err := validate(s); err != nil {
			return nil, err
		}
		return s, nil
	}

	s, err := parseV1(data)
	if err != nil {
		return nil, err
	}
	if err := validate(s); err != nil {
		return nil, err
	}
	return s, nil
}

func parseV1(data []byte) (*Story, error) {
	var raw storyFileV1
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse v1 story yaml: %w", err)
	}

	scenes := make(map[string]Scene, len(raw.Scenes))
	for id, scene := range raw.Scenes {
		scene.ID = id
		scenes[id] = scene
	}

	return &Story{
		Title:      raw.Title,
		StartScene: raw.StartScene,
		Scenes:     scenes,
	}, nil
}

func parseV2(data []byte) (*Story, error) {
	var raw storyFileV2
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse v2 story yaml: %w", err)
	}

	scenes := make(map[string]Scene, len(raw.Scenes))
	for _, wrapped := range raw.Scenes {
		if wrapped.ID == "" {
			return nil, errors.New("v2 scene must define id")
		}
		scene := wrapped.Scene
		scene.ID = wrapped.ID
		scenes[wrapped.ID] = scene
	}

	return &Story{
		ID:         raw.Story.ID,
		Title:      raw.Story.Title,
		Language:   raw.Story.Language,
		StartScene: raw.Story.StartScene,
		Scenes:     scenes,
	}, nil
}

func validate(s *Story) error {
	if s.StartScene == "" {
		return errors.New("story must define start_scene")
	}
	if len(s.Scenes) == 0 {
		return errors.New("story must define at least one scene")
	}
	if _, ok := s.Scenes[s.StartScene]; !ok {
		return fmt.Errorf("start_scene %q not found", s.StartScene)
	}

	for sceneID, scene := range s.Scenes {
		for _, choice := range scene.Choices {
			if choice.Next == "" {
				return fmt.Errorf("scene %q has choice with empty next", sceneID)
			}
			if _, ok := s.Scenes[choice.Next]; !ok {
				return fmt.Errorf("scene %q has choice to unknown scene %q", sceneID, choice.Next)
			}
		}
	}

	return nil
}
