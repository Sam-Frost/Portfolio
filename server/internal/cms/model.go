// Package cms owns the editable content of the public portfolio site
// (projects, experience, blog posts, and the home-page summary). Every edit
// lands in this package's tables as a *draft*; the site only changes when
// Publish serializes the current draft to a content.json and ships it to
// the site's origin (see publisher.go). "Unpublished changes" is the diff
// between the current draft and the last published snapshot.
package cms

import (
	"encoding/json"
	"time"
)

// Project mirrors portfolio-client/src/data/project.ts, plus the CMS-only
// Visible/Order/UpdatedAt bookkeeping fields.
type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Stack       []string  `json:"stack"`
	Github      string    `json:"github"`
	LiveLink    string    `json:"liveLink"`
	Visible     bool      `json:"visible"`
	Order       int       `json:"order"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Experience mirrors portfolio-client/src/data/experience.ts.
type Experience struct {
	ID          string    `json:"id"`
	Logo        string    `json:"logo"`
	Position    string    `json:"position"`
	Company     string    `json:"company"`
	Description string    `json:"description"`
	Details     []string  `json:"details"`
	TechStack   []string  `json:"techStack"`
	StartDate   string    `json:"startDate"`
	EndDate     string    `json:"endDate"`
	Visible     bool      `json:"visible"`
	Order       int       `json:"order"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Blog mirrors portfolio-client/src/data/blog.ts. Body is the post's
// markdown source (what used to live in src/assets/blogs/<slug>.md).
type Blog struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	ReadTime  string    `json:"readTime"`
	Genre     string    `json:"genre"`
	Date      string    `json:"date"`
	Body      string    `json:"body"`
	Visible   bool      `json:"visible"`
	Order     int       `json:"order"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Summary is the single home-page blurb, mirroring
// portfolio-client/src/data/home.ts. Stored as a singleton row.
type Summary struct {
	Domain            string    `json:"domain"`
	ImageSubText      string    `json:"imageSubText"`
	HeroHighlightText string    `json:"heroHighlightText"`
	HeroName          string    `json:"heroName"`
	HeroSubText       string    `json:"heroSubText"`
	HeroDetails       string    `json:"heroDetails"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Content is the assembled draft — and, once published, the exact shape of
// the content.json served at the site origin. The public client filters on
// Visible and sorts on Order itself, so every item is included here.
type Content struct {
	Version     int          `json:"version"`
	PublishedAt time.Time    `json:"publishedAt"`
	Summary     Summary      `json:"summary"`
	Projects    []Project    `json:"projects"`
	Experiences []Experience `json:"experiences"`
	Blogs       []Blog       `json:"blogs"`
}

// PublishStatus values for a Publication row.
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// Publication records one Publish attempt. Snapshot is the content.json
// bytes that were shipped (or would have been, on a failure); it's the
// baseline ChangeSummary diffs the current draft against, and the rollback
// artifact. It's omitted from JSON responses — callers that need it read a
// history object from the origin instead.
type Publication struct {
	ID          string          `json:"id"`
	Version     int             `json:"version"`
	PublishedAt time.Time       `json:"publishedAt"`
	Status      string          `json:"status"`
	Error       *string         `json:"error"`
	Snapshot    json.RawMessage `json:"-"`
}

// Section names reported by ChangeSummary.ChangedSections.
const (
	SectionSummary     = "summary"
	SectionProjects    = "projects"
	SectionExperiences = "experiences"
	SectionBlogs       = "blogs"
)

// ChangeSummary tells the CMS whether the draft differs from what's live,
// and in which sections, so it can render the "N unpublished changes" bar.
type ChangeSummary struct {
	HasUnpublishedChanges bool       `json:"hasUnpublishedChanges"`
	ChangedSections       []string   `json:"changedSections"`
	NeverPublished        bool       `json:"neverPublished"`
	LastPublishedAt       *time.Time `json:"lastPublishedAt"`
	LastPublishVersion    int        `json:"lastPublishVersion"`
	LastPublishStatus     string     `json:"lastPublishStatus"`
	LastPublishError      *string    `json:"lastPublishError"`
}

// ─────────────────────────────────────────────
// Input types — partial-update convention shared with internal/todo,
// internal/notepad, etc.: a nil pointer means "leave unchanged". Slice
// fields are replaced wholesale when non-nil.
// ─────────────────────────────────────────────

type CreateProjectInput struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Stack       []string `json:"stack"`
	Github      string   `json:"github"`
	LiveLink    string   `json:"liveLink"`
	Visible     *bool    `json:"visible"`
}

type UpdateProjectInput struct {
	Title       *string   `json:"title"`
	Slug        *string   `json:"slug"`
	Description *string   `json:"description"`
	Stack       *[]string `json:"stack"`
	Github      *string   `json:"github"`
	LiveLink    *string   `json:"liveLink"`
	Visible     *bool     `json:"visible"`
	Order       *int      `json:"order"`
}

type CreateExperienceInput struct {
	Logo        string   `json:"logo"`
	Position    string   `json:"position"`
	Company     string   `json:"company"`
	Description string   `json:"description"`
	Details     []string `json:"details"`
	TechStack   []string `json:"techStack"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Visible     *bool    `json:"visible"`
}

type UpdateExperienceInput struct {
	Logo        *string   `json:"logo"`
	Position    *string   `json:"position"`
	Company     *string   `json:"company"`
	Description *string   `json:"description"`
	Details     *[]string `json:"details"`
	TechStack   *[]string `json:"techStack"`
	StartDate   *string   `json:"startDate"`
	EndDate     *string   `json:"endDate"`
	Visible     *bool     `json:"visible"`
	Order       *int      `json:"order"`
}

type CreateBlogInput struct {
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	ReadTime string `json:"readTime"`
	Genre    string `json:"genre"`
	Date     string `json:"date"`
	Body     string `json:"body"`
	Visible  *bool  `json:"visible"`
}

type UpdateBlogInput struct {
	Title    *string `json:"title"`
	Slug     *string `json:"slug"`
	ReadTime *string `json:"readTime"`
	Genre    *string `json:"genre"`
	Date     *string `json:"date"`
	Body     *string `json:"body"`
	Visible  *bool   `json:"visible"`
	Order    *int    `json:"order"`
}

// UpdateSummaryInput mirrors settings.UpdateInput: the CMS form always
// submits every field, so these are plain strings, not tri-state pointers.
type UpdateSummaryInput struct {
	Domain            *string `json:"domain"`
	ImageSubText      *string `json:"imageSubText"`
	HeroHighlightText *string `json:"heroHighlightText"`
	HeroName          *string `json:"heroName"`
	HeroSubText       *string `json:"heroSubText"`
	HeroDetails       *string `json:"heroDetails"`
}
