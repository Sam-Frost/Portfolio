package upskill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTopic(ctx context.Context, t Topic) (Topic, error) {
	t.ID = id.New()
	t.DateAdded = time.Now().UTC()

	const q = `INSERT INTO upskill_topics (id, name, target_date, date_added) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.ExecContext(ctx, q, t.ID, t.Name, t.TargetDate, t.DateAdded); err != nil {
		return Topic{}, apperr.Internal("failed to create topic")
	}
	return t, nil
}

func (r *PostgresRepository) ListTopics(ctx context.Context) ([]Topic, error) {
	const q = `
		SELECT t.id, t.name, t.target_date, t.date_added,
		       COUNT(s.id) AS subtopic_count,
		       COUNT(s.id) FILTER (WHERE s.done) AS done_count
		FROM upskill_topics t
		LEFT JOIN upskill_subtopics s ON s.topic_id = t.id
		GROUP BY t.id
		ORDER BY t.date_added DESC
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list topics")
	}
	defer rows.Close()

	topics := make([]Topic, 0)
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.ID, &t.Name, &t.TargetDate, &t.DateAdded, &t.SubtopicCount, &t.DoneCount); err != nil {
			return nil, apperr.Internal("failed to scan topic")
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list topics")
	}

	return topics, nil
}

func (r *PostgresRepository) GetTopic(ctx context.Context, topicID string) (Topic, error) {
	const q = `
		SELECT t.id, t.name, t.target_date, t.date_added,
		       COUNT(s.id) AS subtopic_count,
		       COUNT(s.id) FILTER (WHERE s.done) AS done_count
		FROM upskill_topics t
		LEFT JOIN upskill_subtopics s ON s.topic_id = t.id
		WHERE t.id = $1
		GROUP BY t.id
	`

	var t Topic
	err := r.db.QueryRowContext(ctx, q, topicID).
		Scan(&t.ID, &t.Name, &t.TargetDate, &t.DateAdded, &t.SubtopicCount, &t.DoneCount)
	if errors.Is(err, sql.ErrNoRows) {
		return Topic{}, apperr.NotFound("topic not found")
	}
	if err != nil {
		return Topic{}, apperr.Internal("failed to get topic")
	}
	return t, nil
}

func (r *PostgresRepository) UpdateTopic(ctx context.Context, topicID string, input UpdateTopicInput) (Topic, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.TargetDate != nil {
		sets = append(sets, fmt.Sprintf("target_date = $%d", argN))
		if *input.TargetDate == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.TargetDate)
		}
		argN++
	}

	if len(sets) > 0 {
		args = append(args, topicID)
		q := fmt.Sprintf("UPDATE upskill_topics SET %s WHERE id = $%d", strings.Join(sets, ", "), argN)
		res, err := r.db.ExecContext(ctx, q, args...)
		if err != nil {
			return Topic{}, apperr.Internal("failed to update topic")
		}
		n, err := res.RowsAffected()
		if err != nil {
			return Topic{}, apperr.Internal("failed to update topic")
		}
		if n == 0 {
			return Topic{}, apperr.NotFound("topic not found")
		}
	}

	return r.GetTopic(ctx, topicID)
}

func (r *PostgresRepository) DeleteTopic(ctx context.Context, topicID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM upskill_topics WHERE id = $1`, topicID)
	if err != nil {
		return apperr.Internal("failed to delete topic")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete topic")
	}
	if n == 0 {
		return apperr.NotFound("topic not found")
	}
	return nil
}

func (r *PostgresRepository) CreateSubtopic(ctx context.Context, s Subtopic) (Subtopic, error) {
	s.ID = id.New()
	s.DateAdded = time.Now().UTC()
	s.Done = false

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Subtopic{}, apperr.Internal("failed to create subtopic")
	}
	defer tx.Rollback()

	const insertSubtopic = `INSERT INTO upskill_subtopics (id, topic_id, name, target_date, done, date_added) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.ExecContext(ctx, insertSubtopic, s.ID, s.TopicID, s.Name, s.TargetDate, s.Done, s.DateAdded); err != nil {
		if isForeignKeyViolation(err) {
			return Subtopic{}, apperr.InvalidInput("topic not found")
		}
		return Subtopic{}, apperr.Internal("failed to create subtopic")
	}

	const insertResource = `INSERT INTO upskill_resources (id, subtopic_id, label, url) VALUES ($1, $2, $3, $4)`
	resources := make([]Resource, 0, len(s.Resources))
	for _, res := range s.Resources {
		res.ID = id.New()
		res.SubtopicID = s.ID
		if _, err := tx.ExecContext(ctx, insertResource, res.ID, res.SubtopicID, res.Label, res.URL); err != nil {
			return Subtopic{}, apperr.Internal("failed to create resource")
		}
		resources = append(resources, res)
	}
	s.Resources = resources

	if err := tx.Commit(); err != nil {
		return Subtopic{}, apperr.Internal("failed to create subtopic")
	}

	return s, nil
}

func (r *PostgresRepository) ListSubtopics(ctx context.Context, topicID string) ([]Subtopic, error) {
	const q = `SELECT id, topic_id, name, target_date, done, date_added FROM upskill_subtopics WHERE topic_id = $1 ORDER BY date_added DESC`

	rows, err := r.db.QueryContext(ctx, q, topicID)
	if err != nil {
		return nil, apperr.Internal("failed to list subtopics")
	}
	defer rows.Close()

	subtopics := make([]Subtopic, 0)
	ids := make([]string, 0)
	byID := make(map[string]*Subtopic)
	for rows.Next() {
		var s Subtopic
		if err := rows.Scan(&s.ID, &s.TopicID, &s.Name, &s.TargetDate, &s.Done, &s.DateAdded); err != nil {
			return nil, apperr.Internal("failed to scan subtopic")
		}
		s.Resources = make([]Resource, 0)
		subtopics = append(subtopics, s)
		ids = append(ids, s.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list subtopics")
	}
	for i := range subtopics {
		byID[subtopics[i].ID] = &subtopics[i]
	}

	if len(ids) == 0 {
		return subtopics, nil
	}

	resources, err := r.resourcesFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, res := range resources {
		if s, ok := byID[res.SubtopicID]; ok {
			s.Resources = append(s.Resources, res)
		}
	}

	return subtopics, nil
}

func (r *PostgresRepository) resourcesFor(ctx context.Context, subtopicIDs []string) ([]Resource, error) {
	placeholders := make([]string, len(subtopicIDs))
	args := make([]any, len(subtopicIDs))
	for i, sid := range subtopicIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = sid
	}

	q := fmt.Sprintf(
		`SELECT id, subtopic_id, label, url FROM upskill_resources WHERE subtopic_id IN (%s) ORDER BY id`,
		strings.Join(placeholders, ", "),
	)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, apperr.Internal("failed to list resources")
	}
	defer rows.Close()

	resources := make([]Resource, 0)
	for rows.Next() {
		var res Resource
		if err := rows.Scan(&res.ID, &res.SubtopicID, &res.Label, &res.URL); err != nil {
			return nil, apperr.Internal("failed to scan resource")
		}
		resources = append(resources, res)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list resources")
	}

	return resources, nil
}

func (r *PostgresRepository) UpdateSubtopic(ctx context.Context, subtopicID string, input UpdateSubtopicInput) (Subtopic, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.TargetDate != nil {
		sets = append(sets, fmt.Sprintf("target_date = $%d", argN))
		if *input.TargetDate == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.TargetDate)
		}
		argN++
	}
	if input.Done != nil {
		sets = append(sets, fmt.Sprintf("done = $%d", argN))
		args = append(args, *input.Done)
		argN++
	}

	if len(sets) > 0 {
		args = append(args, subtopicID)
		q := fmt.Sprintf("UPDATE upskill_subtopics SET %s WHERE id = $%d", strings.Join(sets, ", "), argN)
		res, err := r.db.ExecContext(ctx, q, args...)
		if err != nil {
			return Subtopic{}, apperr.Internal("failed to update subtopic")
		}
		n, err := res.RowsAffected()
		if err != nil {
			return Subtopic{}, apperr.Internal("failed to update subtopic")
		}
		if n == 0 {
			return Subtopic{}, apperr.NotFound("subtopic not found")
		}
	}

	return r.getSubtopic(ctx, subtopicID)
}

func (r *PostgresRepository) getSubtopic(ctx context.Context, subtopicID string) (Subtopic, error) {
	const q = `SELECT id, topic_id, name, target_date, done, date_added FROM upskill_subtopics WHERE id = $1`

	var s Subtopic
	err := r.db.QueryRowContext(ctx, q, subtopicID).Scan(&s.ID, &s.TopicID, &s.Name, &s.TargetDate, &s.Done, &s.DateAdded)
	if errors.Is(err, sql.ErrNoRows) {
		return Subtopic{}, apperr.NotFound("subtopic not found")
	}
	if err != nil {
		return Subtopic{}, apperr.Internal("failed to get subtopic")
	}

	resources, err := r.resourcesFor(ctx, []string{s.ID})
	if err != nil {
		return Subtopic{}, err
	}
	s.Resources = resources

	return s, nil
}

func (r *PostgresRepository) DeleteSubtopic(ctx context.Context, subtopicID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM upskill_subtopics WHERE id = $1`, subtopicID)
	if err != nil {
		return apperr.Internal("failed to delete subtopic")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete subtopic")
	}
	if n == 0 {
		return apperr.NotFound("subtopic not found")
	}
	return nil
}

func (r *PostgresRepository) AddResource(ctx context.Context, subtopicID string, resource Resource) (Resource, error) {
	resource.ID = id.New()
	resource.SubtopicID = subtopicID

	const q = `INSERT INTO upskill_resources (id, subtopic_id, label, url) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.ExecContext(ctx, q, resource.ID, resource.SubtopicID, resource.Label, resource.URL); err != nil {
		if isForeignKeyViolation(err) {
			return Resource{}, apperr.InvalidInput("subtopic not found")
		}
		return Resource{}, apperr.Internal("failed to create resource")
	}
	return resource, nil
}

func (r *PostgresRepository) DeleteResource(ctx context.Context, resourceID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM upskill_resources WHERE id = $1`, resourceID)
	if err != nil {
		return apperr.Internal("failed to delete resource")
	}

	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete resource")
	}
	if n == 0 {
		return apperr.NotFound("resource not found")
	}
	return nil
}

// isForeignKeyViolation reports whether err is a Postgres foreign-key
// violation (SQLSTATE 23503), matching internal/todo's helper.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
