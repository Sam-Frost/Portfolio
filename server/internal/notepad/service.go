package notepad

import (
	"context"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Note, error) {
	now := time.Now().UTC()

	title := ""
	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
	}
	if title == "" {
		title = now.Format(DefaultTitleLayout)
	}

	return s.repo.Create(ctx, Note{Title: title, CreatedAt: now, UpdatedAt: now})
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]NoteSummary, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id string) (Note, error) {
	return s.repo.Get(ctx, id)
}

// Update re-defaults the title to the note's original creation timestamp if
// the caller clears it to blank (e.g. autosave firing right after the user
// selects-all-and-deletes the title field), the same rule Create applies —
// a note is never left titleless.
func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Note, error) {
	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		if trimmed == "" {
			existing, err := s.repo.Get(ctx, id)
			if err != nil {
				return Note{}, err
			}
			trimmed = existing.CreatedAt.Format(DefaultTitleLayout)
		}
		input.Title = &trimmed
	}

	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
