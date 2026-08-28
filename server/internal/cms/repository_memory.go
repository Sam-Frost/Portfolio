package cms

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// MemoryRepository is an in-memory Repository for tests. It enforces the
// same slug-uniqueness and ordering guarantees the Postgres schema does.
type MemoryRepository struct {
	mu           sync.Mutex
	projects     map[string]Project
	experiences  map[string]Experience
	blogs        map[string]Blog
	summary      Summary
	publications []Publication
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		projects:    make(map[string]Project),
		experiences: make(map[string]Experience),
		blogs:       make(map[string]Blog),
		summary:     Summary{UpdatedAt: time.Now().UTC()},
	}
}

func now() time.Time { return time.Now().UTC() }

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (r *MemoryRepository) ListProjects(_ context.Context) ([]Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Project, 0, len(r.projects))
	for _, p := range r.projects {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *MemoryRepository) CreateProject(_ context.Context, p Project) (Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.projects {
		if existing.Slug == p.Slug {
			return Project{}, apperr.InvalidInput("slug already in use")
		}
	}
	p.ID = id.New()
	p.UpdatedAt = now()
	if p.Stack == nil {
		p.Stack = []string{}
	}
	r.projects[p.ID] = p
	return p, nil
}

func (r *MemoryRepository) UpdateProject(_ context.Context, pid string, in UpdateProjectInput) (Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.projects[pid]
	if !ok {
		return Project{}, apperr.NotFound("project not found")
	}
	if in.Slug != nil {
		for otherID, other := range r.projects {
			if otherID != pid && other.Slug == *in.Slug {
				return Project{}, apperr.InvalidInput("slug already in use")
			}
		}
		p.Slug = *in.Slug
	}
	if in.Title != nil {
		p.Title = *in.Title
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.Stack != nil {
		p.Stack = *in.Stack
	}
	if in.Github != nil {
		p.Github = *in.Github
	}
	if in.LiveLink != nil {
		p.LiveLink = *in.LiveLink
	}
	if in.Visible != nil {
		p.Visible = *in.Visible
	}
	if in.Order != nil {
		p.Order = *in.Order
	}
	p.UpdatedAt = now()
	r.projects[pid] = p
	return p, nil
}

func (r *MemoryRepository) DeleteProject(_ context.Context, pid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[pid]; !ok {
		return apperr.NotFound("project not found")
	}
	delete(r.projects, pid)
	return nil
}

// ─────────────────────────────────────────────
// Experiences
// ─────────────────────────────────────────────

func (r *MemoryRepository) ListExperiences(_ context.Context) ([]Experience, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Experience, 0, len(r.experiences))
	for _, e := range r.experiences {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *MemoryRepository) CreateExperience(_ context.Context, e Experience) (Experience, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e.ID = id.New()
	e.UpdatedAt = now()
	if e.Details == nil {
		e.Details = []string{}
	}
	if e.TechStack == nil {
		e.TechStack = []string{}
	}
	r.experiences[e.ID] = e
	return e, nil
}

func (r *MemoryRepository) UpdateExperience(_ context.Context, eid string, in UpdateExperienceInput) (Experience, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.experiences[eid]
	if !ok {
		return Experience{}, apperr.NotFound("experience not found")
	}
	if in.Logo != nil {
		e.Logo = *in.Logo
	}
	if in.Position != nil {
		e.Position = *in.Position
	}
	if in.Company != nil {
		e.Company = *in.Company
	}
	if in.Description != nil {
		e.Description = *in.Description
	}
	if in.Details != nil {
		e.Details = *in.Details
	}
	if in.TechStack != nil {
		e.TechStack = *in.TechStack
	}
	if in.StartDate != nil {
		e.StartDate = *in.StartDate
	}
	if in.EndDate != nil {
		e.EndDate = *in.EndDate
	}
	if in.Visible != nil {
		e.Visible = *in.Visible
	}
	if in.Order != nil {
		e.Order = *in.Order
	}
	e.UpdatedAt = now()
	r.experiences[eid] = e
	return e, nil
}

func (r *MemoryRepository) DeleteExperience(_ context.Context, eid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.experiences[eid]; !ok {
		return apperr.NotFound("experience not found")
	}
	delete(r.experiences, eid)
	return nil
}

// ─────────────────────────────────────────────
// Blogs
// ─────────────────────────────────────────────

func (r *MemoryRepository) ListBlogs(_ context.Context) ([]Blog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Blog, 0, len(r.blogs))
	for _, b := range r.blogs {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *MemoryRepository) CreateBlog(_ context.Context, b Blog) (Blog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.blogs {
		if existing.Slug == b.Slug {
			return Blog{}, apperr.InvalidInput("slug already in use")
		}
	}
	b.ID = id.New()
	b.UpdatedAt = now()
	r.blogs[b.ID] = b
	return b, nil
}

func (r *MemoryRepository) UpdateBlog(_ context.Context, bid string, in UpdateBlogInput) (Blog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.blogs[bid]
	if !ok {
		return Blog{}, apperr.NotFound("blog not found")
	}
	if in.Slug != nil {
		for otherID, other := range r.blogs {
			if otherID != bid && other.Slug == *in.Slug {
				return Blog{}, apperr.InvalidInput("slug already in use")
			}
		}
		b.Slug = *in.Slug
	}
	if in.Title != nil {
		b.Title = *in.Title
	}
	if in.ReadTime != nil {
		b.ReadTime = *in.ReadTime
	}
	if in.Genre != nil {
		b.Genre = *in.Genre
	}
	if in.Date != nil {
		b.Date = *in.Date
	}
	if in.Body != nil {
		b.Body = *in.Body
	}
	if in.Visible != nil {
		b.Visible = *in.Visible
	}
	if in.Order != nil {
		b.Order = *in.Order
	}
	b.UpdatedAt = now()
	r.blogs[bid] = b
	return b, nil
}

func (r *MemoryRepository) DeleteBlog(_ context.Context, bid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.blogs[bid]; !ok {
		return apperr.NotFound("blog not found")
	}
	delete(r.blogs, bid)
	return nil
}

// ─────────────────────────────────────────────
// Summary
// ─────────────────────────────────────────────

func (r *MemoryRepository) GetSummary(_ context.Context) (Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.summary, nil
}

func (r *MemoryRepository) UpdateSummary(_ context.Context, in UpdateSummaryInput) (Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if in.Domain != nil {
		r.summary.Domain = *in.Domain
	}
	if in.ImageSubText != nil {
		r.summary.ImageSubText = *in.ImageSubText
	}
	if in.HeroHighlightText != nil {
		r.summary.HeroHighlightText = *in.HeroHighlightText
	}
	if in.HeroName != nil {
		r.summary.HeroName = *in.HeroName
	}
	if in.HeroSubText != nil {
		r.summary.HeroSubText = *in.HeroSubText
	}
	if in.HeroDetails != nil {
		r.summary.HeroDetails = *in.HeroDetails
	}
	r.summary.UpdatedAt = now()
	return r.summary, nil
}

// ─────────────────────────────────────────────
// Publications
// ─────────────────────────────────────────────

func (r *MemoryRepository) LatestPublication(_ context.Context) (Publication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.publications) == 0 {
		return Publication{}, apperr.NotFound("no publications yet")
	}
	return r.publications[len(r.publications)-1], nil
}

func (r *MemoryRepository) LatestSuccessfulPublication(_ context.Context) (Publication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.publications) - 1; i >= 0; i-- {
		if r.publications[i].Status == StatusSuccess {
			return r.publications[i], nil
		}
	}
	return Publication{}, apperr.NotFound("no successful publications yet")
}

func (r *MemoryRepository) RecordPublication(_ context.Context, p Publication) (Publication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.ID = id.New()
	r.publications = append(r.publications, p)
	return p, nil
}

func (r *MemoryRepository) ListPublications(_ context.Context, limit int) ([]Publication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Publication, 0, limit)
	for i := len(r.publications) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, r.publications[i])
	}
	return out, nil
}

func (r *MemoryRepository) NextMaxOrder(_ context.Context, section string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	max := -1
	switch section {
	case SectionProjects:
		for _, p := range r.projects {
			if p.Order > max {
				max = p.Order
			}
		}
	case SectionExperiences:
		for _, e := range r.experiences {
			if e.Order > max {
				max = e.Order
			}
		}
	case SectionBlogs:
		for _, b := range r.blogs {
			if b.Order > max {
				max = b.Order
			}
		}
	}
	return max + 1, nil
}
