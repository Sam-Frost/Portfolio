package cms

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// strSlice maps a Go []string to/from a JSONB column, so the repo doesn't
// need a Postgres array-scanning dependency (pgx/v5 is used via the stdlib
// database/sql driver here, not its native interface).
type strSlice []string

func (s strSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]string(s))
	return string(b), err
}

func (s *strSlice) Scan(src any) error {
	if src == nil {
		*s = []string{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cms: cannot scan %T into strSlice", src)
	}
	if len(b) == 0 {
		*s = []string{}
		return nil
	}
	return json.Unmarshal(b, (*[]string)(s))
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — how a duplicate slug surfaces as a 400.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

const projectCols = `id, title, slug, description, stack, github, live_link, visible, sort_order, updated_at`

func scanProject(sc interface{ Scan(...any) error }) (Project, error) {
	var p Project
	var stack strSlice
	if err := sc.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &stack, &p.Github, &p.LiveLink, &p.Visible, &p.Order, &p.UpdatedAt); err != nil {
		return Project{}, err
	}
	p.Stack = stack
	return p, nil
}

func (r *PostgresRepository) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+projectCols+` FROM cms_projects ORDER BY sort_order, id`)
	if err != nil {
		return nil, apperr.Internal("failed to list projects")
	}
	defer rows.Close()

	out := make([]Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan project")
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list projects")
	}
	return out, nil
}

func (r *PostgresRepository) CreateProject(ctx context.Context, p Project) (Project, error) {
	p.ID = id.New()
	p.UpdatedAt = time.Now().UTC()
	const q = `INSERT INTO cms_projects (` + projectCols + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, q, p.ID, p.Title, p.Slug, p.Description, strSlice(p.Stack), p.Github, p.LiveLink, p.Visible, p.Order, p.UpdatedAt)
	if isUniqueViolation(err) {
		return Project{}, apperr.InvalidInput("slug already in use")
	}
	if err != nil {
		return Project{}, apperr.Internal("failed to create project")
	}
	return p, nil
}

func (r *PostgresRepository) UpdateProject(ctx context.Context, pid string, in UpdateProjectInput) (Project, error) {
	b := newSetBuilder()
	b.add("title", in.Title)
	b.add("slug", in.Slug)
	b.add("description", in.Description)
	if in.Stack != nil {
		b.addRaw("stack", strSlice(*in.Stack))
	}
	b.add("github", in.Github)
	b.add("live_link", in.LiveLink)
	b.addBool("visible", in.Visible)
	b.addInt("sort_order", in.Order)

	if b.empty() {
		return r.getProject(ctx, pid)
	}
	b.addRaw("updated_at", time.Now().UTC())

	q := fmt.Sprintf(`UPDATE cms_projects SET %s WHERE id = %s RETURNING %s`, b.clause(), b.next(), projectCols)
	p, err := scanProject(r.db.QueryRowContext(ctx, q, b.args(pid)...))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, apperr.NotFound("project not found")
	}
	if isUniqueViolation(err) {
		return Project{}, apperr.InvalidInput("slug already in use")
	}
	if err != nil {
		return Project{}, apperr.Internal("failed to update project")
	}
	return p, nil
}

func (r *PostgresRepository) getProject(ctx context.Context, pid string) (Project, error) {
	p, err := scanProject(r.db.QueryRowContext(ctx, `SELECT `+projectCols+` FROM cms_projects WHERE id = $1`, pid))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, apperr.NotFound("project not found")
	}
	if err != nil {
		return Project{}, apperr.Internal("failed to get project")
	}
	return p, nil
}

func (r *PostgresRepository) DeleteProject(ctx context.Context, pid string) error {
	return r.deleteByID(ctx, "cms_projects", pid, "project")
}

// ─────────────────────────────────────────────
// Experiences
// ─────────────────────────────────────────────

const experienceCols = `id, logo, position, company, description, details, tech_stack, start_date, end_date, visible, sort_order, updated_at`

func scanExperience(sc interface{ Scan(...any) error }) (Experience, error) {
	var e Experience
	var details, tech strSlice
	if err := sc.Scan(&e.ID, &e.Logo, &e.Position, &e.Company, &e.Description, &details, &tech, &e.StartDate, &e.EndDate, &e.Visible, &e.Order, &e.UpdatedAt); err != nil {
		return Experience{}, err
	}
	e.Details = details
	e.TechStack = tech
	return e, nil
}

func (r *PostgresRepository) ListExperiences(ctx context.Context) ([]Experience, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+experienceCols+` FROM cms_experiences ORDER BY sort_order, id`)
	if err != nil {
		return nil, apperr.Internal("failed to list experiences")
	}
	defer rows.Close()

	out := make([]Experience, 0)
	for rows.Next() {
		e, err := scanExperience(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan experience")
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list experiences")
	}
	return out, nil
}

func (r *PostgresRepository) CreateExperience(ctx context.Context, e Experience) (Experience, error) {
	e.ID = id.New()
	e.UpdatedAt = time.Now().UTC()
	const q = `INSERT INTO cms_experiences (` + experienceCols + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.db.ExecContext(ctx, q, e.ID, e.Logo, e.Position, e.Company, e.Description, strSlice(e.Details), strSlice(e.TechStack), e.StartDate, e.EndDate, e.Visible, e.Order, e.UpdatedAt)
	if err != nil {
		return Experience{}, apperr.Internal("failed to create experience")
	}
	return e, nil
}

func (r *PostgresRepository) UpdateExperience(ctx context.Context, eid string, in UpdateExperienceInput) (Experience, error) {
	b := newSetBuilder()
	b.add("logo", in.Logo)
	b.add("position", in.Position)
	b.add("company", in.Company)
	b.add("description", in.Description)
	if in.Details != nil {
		b.addRaw("details", strSlice(*in.Details))
	}
	if in.TechStack != nil {
		b.addRaw("tech_stack", strSlice(*in.TechStack))
	}
	b.add("start_date", in.StartDate)
	b.add("end_date", in.EndDate)
	b.addBool("visible", in.Visible)
	b.addInt("sort_order", in.Order)

	if b.empty() {
		return r.getExperience(ctx, eid)
	}
	b.addRaw("updated_at", time.Now().UTC())

	q := fmt.Sprintf(`UPDATE cms_experiences SET %s WHERE id = %s RETURNING %s`, b.clause(), b.next(), experienceCols)
	e, err := scanExperience(r.db.QueryRowContext(ctx, q, b.args(eid)...))
	if errors.Is(err, sql.ErrNoRows) {
		return Experience{}, apperr.NotFound("experience not found")
	}
	if err != nil {
		return Experience{}, apperr.Internal("failed to update experience")
	}
	return e, nil
}

func (r *PostgresRepository) getExperience(ctx context.Context, eid string) (Experience, error) {
	e, err := scanExperience(r.db.QueryRowContext(ctx, `SELECT `+experienceCols+` FROM cms_experiences WHERE id = $1`, eid))
	if errors.Is(err, sql.ErrNoRows) {
		return Experience{}, apperr.NotFound("experience not found")
	}
	if err != nil {
		return Experience{}, apperr.Internal("failed to get experience")
	}
	return e, nil
}

func (r *PostgresRepository) DeleteExperience(ctx context.Context, eid string) error {
	return r.deleteByID(ctx, "cms_experiences", eid, "experience")
}

// ─────────────────────────────────────────────
// Blogs
// ─────────────────────────────────────────────

const blogCols = `id, title, slug, read_time, genre, date, body, visible, sort_order, updated_at`

func scanBlog(sc interface{ Scan(...any) error }) (Blog, error) {
	var b Blog
	if err := sc.Scan(&b.ID, &b.Title, &b.Slug, &b.ReadTime, &b.Genre, &b.Date, &b.Body, &b.Visible, &b.Order, &b.UpdatedAt); err != nil {
		return Blog{}, err
	}
	return b, nil
}

func (r *PostgresRepository) ListBlogs(ctx context.Context) ([]Blog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+blogCols+` FROM cms_blogs ORDER BY sort_order, id`)
	if err != nil {
		return nil, apperr.Internal("failed to list blogs")
	}
	defer rows.Close()

	out := make([]Blog, 0)
	for rows.Next() {
		b, err := scanBlog(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan blog")
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list blogs")
	}
	return out, nil
}

func (r *PostgresRepository) CreateBlog(ctx context.Context, b Blog) (Blog, error) {
	b.ID = id.New()
	b.UpdatedAt = time.Now().UTC()
	const q = `INSERT INTO cms_blogs (` + blogCols + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, q, b.ID, b.Title, b.Slug, b.ReadTime, b.Genre, b.Date, b.Body, b.Visible, b.Order, b.UpdatedAt)
	if isUniqueViolation(err) {
		return Blog{}, apperr.InvalidInput("slug already in use")
	}
	if err != nil {
		return Blog{}, apperr.Internal("failed to create blog")
	}
	return b, nil
}

func (r *PostgresRepository) UpdateBlog(ctx context.Context, bid string, in UpdateBlogInput) (Blog, error) {
	b := newSetBuilder()
	b.add("title", in.Title)
	b.add("slug", in.Slug)
	b.add("read_time", in.ReadTime)
	b.add("genre", in.Genre)
	b.add("date", in.Date)
	b.add("body", in.Body)
	b.addBool("visible", in.Visible)
	b.addInt("sort_order", in.Order)

	if b.empty() {
		return r.getBlog(ctx, bid)
	}
	b.addRaw("updated_at", time.Now().UTC())

	q := fmt.Sprintf(`UPDATE cms_blogs SET %s WHERE id = %s RETURNING %s`, b.clause(), b.next(), blogCols)
	blog, err := scanBlog(r.db.QueryRowContext(ctx, q, b.args(bid)...))
	if errors.Is(err, sql.ErrNoRows) {
		return Blog{}, apperr.NotFound("blog not found")
	}
	if isUniqueViolation(err) {
		return Blog{}, apperr.InvalidInput("slug already in use")
	}
	if err != nil {
		return Blog{}, apperr.Internal("failed to update blog")
	}
	return blog, nil
}

func (r *PostgresRepository) getBlog(ctx context.Context, bid string) (Blog, error) {
	blog, err := scanBlog(r.db.QueryRowContext(ctx, `SELECT `+blogCols+` FROM cms_blogs WHERE id = $1`, bid))
	if errors.Is(err, sql.ErrNoRows) {
		return Blog{}, apperr.NotFound("blog not found")
	}
	if err != nil {
		return Blog{}, apperr.Internal("failed to get blog")
	}
	return blog, nil
}

func (r *PostgresRepository) DeleteBlog(ctx context.Context, bid string) error {
	return r.deleteByID(ctx, "cms_blogs", bid, "blog")
}

// ─────────────────────────────────────────────
// Summary (singleton, seeded by the migration like internal/settings)
// ─────────────────────────────────────────────

const summaryID = "singleton"

func (r *PostgresRepository) GetSummary(ctx context.Context) (Summary, error) {
	const q = `SELECT domain, image_sub_text, hero_highlight_text, hero_name, hero_sub_text, hero_details, updated_at FROM cms_summary WHERE id = $1`
	var s Summary
	err := r.db.QueryRowContext(ctx, q, summaryID).Scan(&s.Domain, &s.ImageSubText, &s.HeroHighlightText, &s.HeroName, &s.HeroSubText, &s.HeroDetails, &s.UpdatedAt)
	if err != nil {
		return Summary{}, apperr.Internal("failed to get summary")
	}
	return s, nil
}

func (r *PostgresRepository) UpdateSummary(ctx context.Context, in UpdateSummaryInput) (Summary, error) {
	b := newSetBuilder()
	b.add("domain", in.Domain)
	b.add("image_sub_text", in.ImageSubText)
	b.add("hero_highlight_text", in.HeroHighlightText)
	b.add("hero_name", in.HeroName)
	b.add("hero_sub_text", in.HeroSubText)
	b.add("hero_details", in.HeroDetails)

	if b.empty() {
		return r.GetSummary(ctx)
	}
	b.addRaw("updated_at", time.Now().UTC())

	q := fmt.Sprintf(`UPDATE cms_summary SET %s WHERE id = %s`, b.clause(), b.next())
	if _, err := r.db.ExecContext(ctx, q, b.args(summaryID)...); err != nil {
		return Summary{}, apperr.Internal("failed to update summary")
	}
	return r.GetSummary(ctx)
}

// ─────────────────────────────────────────────
// Publications
// ─────────────────────────────────────────────

const publicationCols = `id, version, published_at, status, error, snapshot`

func scanPublication(sc interface{ Scan(...any) error }, withSnapshot bool) (Publication, error) {
	var p Publication
	var errMsg sql.NullString
	var snapshot []byte
	if err := sc.Scan(&p.ID, &p.Version, &p.PublishedAt, &p.Status, &errMsg, &snapshot); err != nil {
		return Publication{}, err
	}
	if errMsg.Valid {
		p.Error = &errMsg.String
	}
	if withSnapshot {
		p.Snapshot = snapshot
	}
	return p, nil
}

func (r *PostgresRepository) LatestPublication(ctx context.Context) (Publication, error) {
	p, err := scanPublication(r.db.QueryRowContext(ctx,
		`SELECT `+publicationCols+` FROM cms_publications ORDER BY version DESC LIMIT 1`), true)
	if errors.Is(err, sql.ErrNoRows) {
		return Publication{}, apperr.NotFound("no publications yet")
	}
	if err != nil {
		return Publication{}, apperr.Internal("failed to get latest publication")
	}
	return p, nil
}

func (r *PostgresRepository) LatestSuccessfulPublication(ctx context.Context) (Publication, error) {
	p, err := scanPublication(r.db.QueryRowContext(ctx,
		`SELECT `+publicationCols+` FROM cms_publications WHERE status = $1 ORDER BY version DESC LIMIT 1`, StatusSuccess), true)
	if errors.Is(err, sql.ErrNoRows) {
		return Publication{}, apperr.NotFound("no successful publications yet")
	}
	if err != nil {
		return Publication{}, apperr.Internal("failed to get latest publication")
	}
	return p, nil
}

func (r *PostgresRepository) RecordPublication(ctx context.Context, p Publication) (Publication, error) {
	p.ID = id.New()
	snapshot := []byte(p.Snapshot)
	if snapshot == nil {
		snapshot = []byte("null")
	}
	const q = `INSERT INTO cms_publications (id, version, published_at, status, error, snapshot) VALUES ($1,$2,$3,$4,$5,$6)`
	if _, err := r.db.ExecContext(ctx, q, p.ID, p.Version, p.PublishedAt, p.Status, nullString(p.Error), snapshot); err != nil {
		return Publication{}, apperr.Internal("failed to record publication")
	}
	return p, nil
}

func (r *PostgresRepository) ListPublications(ctx context.Context, limit int) ([]Publication, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+publicationCols+` FROM cms_publications ORDER BY version DESC LIMIT $1`, limit)
	if err != nil {
		return nil, apperr.Internal("failed to list publications")
	}
	defer rows.Close()

	out := make([]Publication, 0)
	for rows.Next() {
		p, err := scanPublication(rows, false)
		if err != nil {
			return nil, apperr.Internal("failed to scan publication")
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list publications")
	}
	return out, nil
}

func (r *PostgresRepository) NextMaxOrder(ctx context.Context, section string) (int, error) {
	table := map[string]string{
		SectionProjects:    "cms_projects",
		SectionExperiences: "cms_experiences",
		SectionBlogs:       "cms_blogs",
	}[section]
	if table == "" {
		return 0, apperr.Internal("unknown section")
	}
	var max sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT MAX(sort_order) FROM `+table).Scan(&max); err != nil {
		return 0, apperr.Internal("failed to compute order")
	}
	if !max.Valid {
		return 0, nil
	}
	return int(max.Int64) + 1, nil
}

// ─────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────

func (r *PostgresRepository) deleteByID(ctx context.Context, table, rowID, label string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM `+table+` WHERE id = $1`, rowID)
	if err != nil {
		return apperr.Internal("failed to delete " + label)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete " + label)
	}
	if n == 0 {
		return apperr.NotFound(label + " not found")
	}
	return nil
}

func nullString(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// setBuilder assembles a dynamic "col = $n, ..." UPDATE clause, matching the
// approach in internal/todo's repository_postgres.go but shared across the
// four cms tables.
type setBuilder struct {
	sets []string
	vals []any
	n    int
}

func newSetBuilder() *setBuilder { return &setBuilder{n: 1} }

func (b *setBuilder) add(col string, v *string) {
	if v == nil {
		return
	}
	b.addRaw(col, *v)
}

func (b *setBuilder) addBool(col string, v *bool) {
	if v == nil {
		return
	}
	b.addRaw(col, *v)
}

func (b *setBuilder) addInt(col string, v *int) {
	if v == nil {
		return
	}
	b.addRaw(col, *v)
}

func (b *setBuilder) addRaw(col string, v any) {
	b.sets = append(b.sets, fmt.Sprintf("%s = $%d", col, b.n))
	b.vals = append(b.vals, v)
	b.n++
}

func (b *setBuilder) empty() bool          { return len(b.sets) == 0 }
func (b *setBuilder) clause() string       { return strings.Join(b.sets, ", ") }
func (b *setBuilder) next() string         { return fmt.Sprintf("$%d", b.n) }
func (b *setBuilder) args(id string) []any { return append(b.vals, id) }
