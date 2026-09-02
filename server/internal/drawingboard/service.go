package drawingboard

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/reqlog"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Board, error) {
	now := time.Now().UTC()

	name := ""
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	if name == "" {
		name = now.Format(DefaultNameLayout)
	}

	board, err := s.repo.Create(ctx, Board{
		Name:      name,
		SceneData: json.RawMessage(DefaultSceneData),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Board{}, err
	}

	reqlog.FromContext(ctx).InfoContext(ctx, "drawing board created",
		"board_id", board.ID, "name", board.Name)
	return board, nil
}

func (s *Service) List(ctx context.Context) ([]BoardSummary, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (Board, error) {
	return s.repo.Get(ctx, id)
}

// Update re-defaults the name to the board's original creation timestamp if
// the caller clears it to blank, the same rule notepad's Update applies — a
// board is never left nameless. SceneData, when present, must decode to a
// JSON object: it's the frontend's serialized Excalidraw scene
// (elements/appState/files), and a malformed or wrong-shaped autosave
// payload would otherwise corrupt the board silently. Note that malformed
// JSON in the *request body itself* never reaches here — httpx.DecodeJSON
// rejects that first, since json.RawMessage can only ever hold a value the
// outer decode already parsed successfully.
func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Board, error) {
	if input.SceneData != nil {
		var v any
		if err := json.Unmarshal(input.SceneData, &v); err != nil {
			return Board{}, apperr.InvalidInput("sceneData must be valid JSON")
		}
		if _, ok := v.(map[string]any); !ok {
			return Board{}, apperr.InvalidInput("sceneData must be a JSON object")
		}
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			existing, err := s.repo.Get(ctx, id)
			if err != nil {
				return Board{}, err
			}
			trimmed = existing.CreatedAt.Format(DefaultNameLayout)
		}
		input.Name = &trimmed
	}

	board, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return Board{}, err
	}

	// Scene autosave fires on every (debounced) canvas change, so it logs
	// at debug; a rename is a deliberate, rare action worth an info line.
	logger := reqlog.FromContext(ctx)
	if input.Name != nil {
		logger.InfoContext(ctx, "drawing board renamed", "board_id", board.ID, "name", board.Name)
	}
	if input.SceneData != nil {
		logger.DebugContext(ctx, "drawing board scene saved",
			"board_id", board.ID, "scene_bytes", len(input.SceneData))
	}
	return board, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	reqlog.FromContext(ctx).InfoContext(ctx, "drawing board deleted", "board_id", id)
	return nil
}
