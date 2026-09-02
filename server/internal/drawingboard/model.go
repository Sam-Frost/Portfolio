package drawingboard

import (
	"encoding/json"
	"time"
)

// DefaultSceneData is the scene an empty new board is created with — the
// shape @excalidraw/excalidraw's <Excalidraw> component expects for its
// initialData prop (elements/appState/files).
const DefaultSceneData = `{"elements":[],"appState":{},"files":{}}`

// Board is the full board, scene data included — returned by Create, Get,
// and Update. List returns BoardSummary instead so the board list view
// isn't forced to pull every board's (potentially large) scene over the
// wire.
type Board struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	SceneData json.RawMessage `json:"sceneData"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type BoardSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DefaultNameLayout is used to name a board when it's created without a
// name: the board's created-at timestamp, e.g. "Jan 2, 2006 3:04 PM".
const DefaultNameLayout = "Jan 2, 2006 3:04 PM"

type CreateInput struct {
	Name *string `json:"name"`
}

// UpdateInput is a partial update: nil fields are left unchanged, matching
// the notepad/todo packages' UpdateInput convention. Autosave sends
// SceneData on (debounced) every canvas change; Name only when the board is
// renamed.
type UpdateInput struct {
	Name      *string         `json:"name"`
	SceneData json.RawMessage `json:"sceneData"`
}

func updatedAtNow() time.Time { return time.Now().UTC() }
