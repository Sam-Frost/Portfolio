package notepad

import (
	"context"
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

// Scratch returns the singleton "Random Notepad" scratch note, created on
// first access. It has no title and never appears in List; the frontend
// edits it through the ordinary Update path once it has the returned ID.
func (s *Service) Scratch(ctx context.Context) (Note, error) {
	return s.repo.Scratch(ctx)
}

// Update re-defaults the title to the note's original creation timestamp if
// the caller clears it to blank (e.g. autosave firing right after the user
// selects-all-and-deletes the title field), the same rule Create applies —
// a note is never left titleless.
func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Note, error) {
	// A locked note is view-only: its title and body can't be edited until
	// it's unlocked. Unlocking in the same request is allowed (that's how
	// the editor leaves view-only mode); pin/archive stay available since
	// they're list bookkeeping, not content.
	editsContent := input.Title != nil || input.ContentHTML != nil
	unlocking := input.Locked != nil && !*input.Locked
	if editsContent && !unlocking {
		existing, err := s.repo.Get(ctx, id)
		if err != nil {
			return Note{}, err
		}
		if existing.Locked {
			return Note{}, apperr.InvalidInput("note is locked")
		}
	}

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

	note, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return Note{}, err
	}

	// Title/body autosave is high-frequency and not worth a line each; a
	// label change is a deliberate one-off worth recording.
	if input.LabelID != nil {
		logger := reqlog.FromContext(ctx)
		if *input.LabelID == "" {
			logger.InfoContext(ctx, "note label cleared", "note_id", note.ID)
		} else {
			logger.InfoContext(ctx, "note label set", "note_id", note.ID, "label_id", *input.LabelID)
		}
	}
	return note, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	reqlog.FromContext(ctx).InfoContext(ctx, "note deleted", "note_id", id)
	return nil
}
