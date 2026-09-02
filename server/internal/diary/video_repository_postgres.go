package diary

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresVideoRepository struct {
	db *sql.DB
}

func NewPostgresVideoRepository(db *sql.DB) *PostgresVideoRepository {
	return &PostgresVideoRepository{db: db}
}

const videoColumns = `id, entry_date, title, s3_key, upload_id, content_type, size_bytes, duration_seconds, status, created_at, uploaded_at`

func scanVideo(s interface {
	Scan(dest ...any) error
}) (Video, error) {
	var (
		v        Video
		d        time.Time
		uploadID sql.NullString
	)
	if err := s.Scan(&v.ID, &d, &v.Title, &v.S3Key, &uploadID, &v.ContentType, &v.SizeBytes, &v.DurationSeconds, &v.Status, &v.CreatedAt, &v.UploadedAt); err != nil {
		return Video{}, err
	}
	v.EntryDate = d.Format(EntryDateLayout)
	v.UploadID = uploadID.String
	return v, nil
}

func (r *PostgresVideoRepository) Create(ctx context.Context, v Video) (Video, error) {
	const q = `
		INSERT INTO diary_videos (id, entry_date, title, s3_key, upload_id, content_type, size_bytes, status, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, 0, $7, $8)
		RETURNING ` + videoColumns

	now := time.Now().UTC()
	row := r.db.QueryRowContext(ctx, q, id.New(), v.EntryDate, v.Title, v.S3Key, v.UploadID, v.ContentType, v.Status, now)
	out, err := scanVideo(row)
	if err != nil {
		return Video{}, apperr.Internal("failed to create diary video")
	}
	return out, nil
}

func (r *PostgresVideoRepository) GetByID(ctx context.Context, videoID string) (Video, error) {
	const q = `SELECT ` + videoColumns + ` FROM diary_videos WHERE id = $1`

	v, err := scanVideo(r.db.QueryRowContext(ctx, q, videoID))
	if errors.Is(err, sql.ErrNoRows) {
		return Video{}, apperr.NotFound("diary video not found")
	}
	if err != nil {
		return Video{}, apperr.Internal("failed to get diary video")
	}
	return v, nil
}

func (r *PostgresVideoRepository) ListByDate(ctx context.Context, date string) ([]Video, error) {
	const q = `SELECT ` + videoColumns + ` FROM diary_videos WHERE entry_date = $1 ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, q, date)
	if err != nil {
		return nil, apperr.Internal("failed to list diary videos")
	}
	defer rows.Close()

	out := make([]Video, 0)
	for rows.Next() {
		v, err := scanVideo(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan diary video")
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list diary videos")
	}
	return out, nil
}

func (r *PostgresVideoRepository) CountsByDateRange(ctx context.Context, from, to string) (map[string]int, error) {
	const q = `
		SELECT entry_date, COUNT(*) FROM diary_videos
		WHERE status = $1 AND entry_date BETWEEN $2 AND $3
		GROUP BY entry_date`

	rows, err := r.db.QueryContext(ctx, q, VideoStatusReady, from, to)
	if err != nil {
		return nil, apperr.Internal("failed to count diary videos")
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var d time.Time
		var n int
		if err := rows.Scan(&d, &n); err != nil {
			return nil, apperr.Internal("failed to scan diary video count")
		}
		counts[d.Format(EntryDateLayout)] = n
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to count diary videos")
	}
	return counts, nil
}

func (r *PostgresVideoRepository) SetUploadID(ctx context.Context, videoID, uploadID string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE diary_videos SET upload_id = $1 WHERE id = $2`, uploadID, videoID)
	if err != nil {
		return apperr.Internal("failed to record upload id")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("diary video not found")
	}
	return nil
}

func (r *PostgresVideoRepository) MarkReady(ctx context.Context, videoID string, size int64, durationSeconds *int) (Video, error) {
	const q = `
		UPDATE diary_videos
		SET status = $1, size_bytes = $2, duration_seconds = $3, upload_id = NULL, uploaded_at = $4
		WHERE id = $5
		RETURNING ` + videoColumns

	v, err := scanVideo(r.db.QueryRowContext(ctx, q, VideoStatusReady, size, durationSeconds, time.Now().UTC(), videoID))
	if errors.Is(err, sql.ErrNoRows) {
		return Video{}, apperr.NotFound("diary video not found")
	}
	if err != nil {
		return Video{}, apperr.Internal("failed to finalize diary video")
	}
	return v, nil
}

func (r *PostgresVideoRepository) Delete(ctx context.Context, videoID string) (s3Key, uploadID, status string, err error) {
	const q = `DELETE FROM diary_videos WHERE id = $1 RETURNING s3_key, COALESCE(upload_id, ''), status`

	var key, uid, st string
	scanErr := r.db.QueryRowContext(ctx, q, videoID).Scan(&key, &uid, &st)
	if errors.Is(scanErr, sql.ErrNoRows) {
		return "", "", "", apperr.NotFound("diary video not found")
	}
	if scanErr != nil {
		return "", "", "", apperr.Internal("failed to delete diary video")
	}
	return key, uid, st, nil
}
