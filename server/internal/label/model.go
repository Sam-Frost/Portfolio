package label

// AllowedColors are the fixed set of preset swatches the settings UI offers;
// Color must be one of these tokens, kept in sync with the frontend's
// LABEL_COLORS.
var AllowedColors = map[string]bool{
	"red": true, "orange": true, "yellow": true, "green": true,
	"teal": true, "blue": true, "purple": true, "pink": true,
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CreateInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateInput is a partial update: nil fields are left unchanged, matching
// the todo package's UpdateInput convention.
type UpdateInput struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}
