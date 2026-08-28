package cms

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo      Repository
	publisher Publisher
}

func NewService(repo Repository, publisher Publisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// slugify turns a title into a URL-safe slug: lowercase, alphanumerics
// kept, every other run collapsed to a single dash.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func resolveSlug(explicit, title string) (string, error) {
	s := strings.TrimSpace(explicit)
	if s == "" {
		s = slugify(title)
	}
	if s == "" {
		return "", apperr.InvalidInput("slug is required (could not derive one from the title)")
	}
	if !slugPattern.MatchString(s) {
		return "", apperr.InvalidInput("slug must be lowercase letters, numbers, and single dashes")
	}
	return s, nil
}

func validateURL(field, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return apperr.InvalidInput(field + " must be a valid http(s) URL")
	}
	return nil
}

func required(field, val string) error {
	if strings.TrimSpace(val) == "" {
		return apperr.InvalidInput(field + " is required")
	}
	return nil
}

func normStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context) ([]Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *Service) CreateProject(ctx context.Context, in CreateProjectInput) (Project, error) {
	if err := required("title", in.Title); err != nil {
		return Project{}, err
	}
	slug, err := resolveSlug(in.Slug, in.Title)
	if err != nil {
		return Project{}, err
	}
	if err := validateURL("github", in.Github); err != nil {
		return Project{}, err
	}
	if err := validateURL("liveLink", in.LiveLink); err != nil {
		return Project{}, err
	}

	order, err := s.repo.NextMaxOrder(ctx, SectionProjects)
	if err != nil {
		return Project{}, err
	}

	return s.repo.CreateProject(ctx, Project{
		Title:       strings.TrimSpace(in.Title),
		Slug:        slug,
		Description: strings.TrimSpace(in.Description),
		Stack:       normStrings(in.Stack),
		Github:      strings.TrimSpace(in.Github),
		LiveLink:    strings.TrimSpace(in.LiveLink),
		Visible:     in.Visible == nil || *in.Visible,
		Order:       order,
	})
}

func (s *Service) UpdateProject(ctx context.Context, id string, in UpdateProjectInput) (Project, error) {
	if in.Title != nil {
		if err := required("title", *in.Title); err != nil {
			return Project{}, err
		}
	}
	if in.Slug != nil {
		slug, err := resolveSlug(*in.Slug, deref(in.Title))
		if err != nil {
			return Project{}, err
		}
		in.Slug = &slug
	}
	if in.Github != nil {
		if err := validateURL("github", *in.Github); err != nil {
			return Project{}, err
		}
	}
	if in.LiveLink != nil {
		if err := validateURL("liveLink", *in.LiveLink); err != nil {
			return Project{}, err
		}
	}
	if in.Stack != nil {
		norm := normStrings(*in.Stack)
		in.Stack = &norm
	}
	return s.repo.UpdateProject(ctx, id, in)
}

func (s *Service) DeleteProject(ctx context.Context, id string) error {
	return s.repo.DeleteProject(ctx, id)
}

// ─────────────────────────────────────────────
// Experiences
// ─────────────────────────────────────────────

func (s *Service) ListExperiences(ctx context.Context) ([]Experience, error) {
	return s.repo.ListExperiences(ctx)
}

func (s *Service) CreateExperience(ctx context.Context, in CreateExperienceInput) (Experience, error) {
	if err := required("position", in.Position); err != nil {
		return Experience{}, err
	}
	if err := required("company", in.Company); err != nil {
		return Experience{}, err
	}

	order, err := s.repo.NextMaxOrder(ctx, SectionExperiences)
	if err != nil {
		return Experience{}, err
	}

	return s.repo.CreateExperience(ctx, Experience{
		Logo:        strings.TrimSpace(in.Logo),
		Position:    strings.TrimSpace(in.Position),
		Company:     strings.TrimSpace(in.Company),
		Description: strings.TrimSpace(in.Description),
		Details:     normStrings(in.Details),
		TechStack:   normStrings(in.TechStack),
		StartDate:   strings.TrimSpace(in.StartDate),
		EndDate:     strings.TrimSpace(in.EndDate),
		Visible:     in.Visible == nil || *in.Visible,
		Order:       order,
	})
}

func (s *Service) UpdateExperience(ctx context.Context, id string, in UpdateExperienceInput) (Experience, error) {
	if in.Position != nil {
		if err := required("position", *in.Position); err != nil {
			return Experience{}, err
		}
	}
	if in.Company != nil {
		if err := required("company", *in.Company); err != nil {
			return Experience{}, err
		}
	}
	if in.Details != nil {
		norm := normStrings(*in.Details)
		in.Details = &norm
	}
	if in.TechStack != nil {
		norm := normStrings(*in.TechStack)
		in.TechStack = &norm
	}
	return s.repo.UpdateExperience(ctx, id, in)
}

func (s *Service) DeleteExperience(ctx context.Context, id string) error {
	return s.repo.DeleteExperience(ctx, id)
}

// ─────────────────────────────────────────────
// Blogs
// ─────────────────────────────────────────────

func (s *Service) ListBlogs(ctx context.Context) ([]Blog, error) {
	return s.repo.ListBlogs(ctx)
}

func (s *Service) CreateBlog(ctx context.Context, in CreateBlogInput) (Blog, error) {
	if err := required("title", in.Title); err != nil {
		return Blog{}, err
	}
	if err := required("body", in.Body); err != nil {
		return Blog{}, err
	}
	slug, err := resolveSlug(in.Slug, in.Title)
	if err != nil {
		return Blog{}, err
	}

	order, err := s.repo.NextMaxOrder(ctx, SectionBlogs)
	if err != nil {
		return Blog{}, err
	}

	return s.repo.CreateBlog(ctx, Blog{
		Title:    strings.TrimSpace(in.Title),
		Slug:     slug,
		ReadTime: strings.TrimSpace(in.ReadTime),
		Genre:    strings.TrimSpace(in.Genre),
		Date:     strings.TrimSpace(in.Date),
		Body:     in.Body,
		Visible:  in.Visible == nil || *in.Visible,
		Order:    order,
	})
}

func (s *Service) UpdateBlog(ctx context.Context, id string, in UpdateBlogInput) (Blog, error) {
	if in.Title != nil {
		if err := required("title", *in.Title); err != nil {
			return Blog{}, err
		}
	}
	if in.Body != nil {
		if err := required("body", *in.Body); err != nil {
			return Blog{}, err
		}
	}
	if in.Slug != nil {
		slug, err := resolveSlug(*in.Slug, deref(in.Title))
		if err != nil {
			return Blog{}, err
		}
		in.Slug = &slug
	}
	return s.repo.UpdateBlog(ctx, id, in)
}

func (s *Service) DeleteBlog(ctx context.Context, id string) error {
	return s.repo.DeleteBlog(ctx, id)
}

// ─────────────────────────────────────────────
// Summary
// ─────────────────────────────────────────────

func (s *Service) GetSummary(ctx context.Context) (Summary, error) {
	return s.repo.GetSummary(ctx)
}

func (s *Service) UpdateSummary(ctx context.Context, in UpdateSummaryInput) (Summary, error) {
	return s.repo.UpdateSummary(ctx, in)
}

// ─────────────────────────────────────────────
// Assemble / publish
// ─────────────────────────────────────────────

// Content assembles the current draft: the summary plus every project,
// experience and blog, each list ordered by its Order field.
func (s *Service) Content(ctx context.Context) (Content, error) {
	return s.assemble(ctx)
}

func (s *Service) assemble(ctx context.Context) (Content, error) {
	summary, err := s.repo.GetSummary(ctx)
	if err != nil {
		return Content{}, err
	}
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return Content{}, err
	}
	experiences, err := s.repo.ListExperiences(ctx)
	if err != nil {
		return Content{}, err
	}
	blogs, err := s.repo.ListBlogs(ctx)
	if err != nil {
		return Content{}, err
	}
	return Content{
		Summary:     summary,
		Projects:    projects,
		Experiences: experiences,
		Blogs:       blogs,
	}, nil
}

func (s *Service) Publications(ctx context.Context, limit int) ([]Publication, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.ListPublications(ctx, limit)
}

// Status reports whether the draft differs from the last successful
// publish, and in which sections.
func (s *Service) Status(ctx context.Context) (ChangeSummary, error) {
	draft, err := s.assemble(ctx)
	if err != nil {
		return ChangeSummary{}, err
	}

	cs := ChangeSummary{}

	if last, err := s.repo.LatestPublication(ctx); err == nil {
		cs.LastPublishedAt = &last.PublishedAt
		cs.LastPublishVersion = last.Version
		cs.LastPublishStatus = last.Status
		cs.LastPublishError = last.Error
	} else if !isNotFound(err) {
		return ChangeSummary{}, err
	}

	success, err := s.repo.LatestSuccessfulPublication(ctx)
	if isNotFound(err) {
		cs.NeverPublished = true
		cs.HasUnpublishedChanges = true
		cs.ChangedSections = []string{SectionSummary, SectionProjects, SectionExperiences, SectionBlogs}
		return cs, nil
	}
	if err != nil {
		return ChangeSummary{}, err
	}

	var live Content
	if err := json.Unmarshal(success.Snapshot, &live); err != nil {
		return ChangeSummary{}, apperr.Internal("failed to read last published snapshot")
	}

	d := stripForDiff(draft)
	l := stripForDiff(live)
	changed := make([]string, 0, 4)
	if !jsonEqual(d.Summary, l.Summary) {
		changed = append(changed, SectionSummary)
	}
	if !jsonEqual(d.Projects, l.Projects) {
		changed = append(changed, SectionProjects)
	}
	if !jsonEqual(d.Experiences, l.Experiences) {
		changed = append(changed, SectionExperiences)
	}
	if !jsonEqual(d.Blogs, l.Blogs) {
		changed = append(changed, SectionBlogs)
	}

	cs.ChangedSections = changed
	cs.HasUnpublishedChanges = len(changed) > 0
	return cs, nil
}

// Publish serializes the current draft to content.json and ships it via the
// configured Publisher. Every attempt is recorded as a Publication row
// (StatusFailed on a publisher error), so a failure is visible in the CMS
// rather than silent.
func (s *Service) Publish(ctx context.Context) (Publication, error) {
	if !s.publisher.Enabled() {
		return Publication{}, apperr.Conflict("publishing is not configured on this server")
	}

	draft, err := s.assemble(ctx)
	if err != nil {
		return Publication{}, err
	}

	version := 1
	if last, err := s.repo.LatestPublication(ctx); err == nil {
		version = last.Version + 1
	} else if !isNotFound(err) {
		return Publication{}, err
	}

	draft.Version = version
	draft.PublishedAt = time.Now().UTC()

	body, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return Publication{}, apperr.Internal("failed to serialize content")
	}

	rec := Publication{
		Version:     version,
		PublishedAt: draft.PublishedAt,
		Status:      StatusSuccess,
		Snapshot:    body,
	}

	pubErr := s.publisher.Publish(ctx, version, body)
	if pubErr != nil {
		msg := pubErr.Error()
		rec.Status = StatusFailed
		rec.Error = &msg
	}

	saved, err := s.repo.RecordPublication(ctx, rec)
	if err != nil {
		return Publication{}, err
	}
	if pubErr != nil {
		return saved, apperr.Internal("publish failed: " + pubErr.Error())
	}
	return saved, nil
}

// stripForDiff zeroes the bookkeeping fields that shouldn't count as
// content changes (per-item UpdatedAt, and the envelope's Version /
// PublishedAt), so re-saving an item with identical values doesn't show up
// as an unpublished change.
func stripForDiff(c Content) Content {
	c.Version = 0
	c.PublishedAt = time.Time{}
	c.Summary.UpdatedAt = time.Time{}
	projects := make([]Project, len(c.Projects))
	for i, p := range c.Projects {
		p.UpdatedAt = time.Time{}
		projects[i] = p
	}
	c.Projects = projects
	experiences := make([]Experience, len(c.Experiences))
	for i, e := range c.Experiences {
		e.UpdatedAt = time.Time{}
		experiences[i] = e
	}
	c.Experiences = experiences
	blogs := make([]Blog, len(c.Blogs))
	for i, b := range c.Blogs {
		b.UpdatedAt = time.Time{}
		blogs[i] = b
	}
	c.Blogs = blogs
	return c
}

func jsonEqual(a, b any) bool {
	ab, err1 := json.Marshal(a)
	bb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(ab) == string(bb)
}

func isNotFound(err error) bool {
	var appErr *apperr.Error
	return errors.As(err, &appErr) && appErr.Kind == apperr.KindNotFound
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
