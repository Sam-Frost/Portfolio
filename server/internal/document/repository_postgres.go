package document

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

const folderCols = "id, parent_id, name, created_at"

func scanFolder(sc interface{ Scan(...any) error }) (Folder, error) {
	var f Folder
	if err := sc.Scan(&f.ID, &f.ParentID, &f.Name, &f.CreatedAt); err != nil {
		return Folder{}, err
	}
	return f, nil
}

const documentCols = "id, folder_id, label_id, name, s3_key, content_type, size_bytes, status, created_at, uploaded_at"

func scanDocument(sc interface{ Scan(...any) error }) (Document, error) {
	var d Document
	if err := sc.Scan(
		&d.ID, &d.FolderID, &d.LabelID, &d.Name, &d.S3Key,
		&d.ContentType, &d.SizeBytes, &d.Status, &d.CreatedAt, &d.UploadedAt,
	); err != nil {
		return Document{}, err
	}
	return d, nil
}

// --- folders ---

func (r *PostgresRepository) CreateFolder(ctx context.Context, f Folder) (Folder, error) {
	f.ID = id.New()
	f.CreatedAt = time.Now().UTC()

	const q = `INSERT INTO document_folders (id, parent_id, name, created_at) VALUES ($1, $2, $3, $4)`
	if _, err := r.db.ExecContext(ctx, q, f.ID, f.ParentID, f.Name, f.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return Folder{}, apperr.InvalidInput("a folder with this name already exists here")
		}
		if isForeignKeyViolation(err) {
			return Folder{}, apperr.InvalidInput("parent folder not found")
		}
		return Folder{}, apperr.Internal("failed to create folder")
	}
	return f, nil
}

func (r *PostgresRepository) ListFolders(ctx context.Context) ([]Folder, error) {
	// Depth + a materialised name path give a stable parent-before-child,
	// siblings-by-name ordering the caller can turn straight into a tree.
	const q = `
		WITH RECURSIVE tree AS (
			SELECT ` + folderCols + `, 1 AS depth, lower(name) AS path
			FROM document_folders WHERE parent_id IS NULL
			UNION ALL
			SELECT f.id, f.parent_id, f.name, f.created_at, t.depth + 1, t.path || '/' || lower(f.name)
			FROM document_folders f JOIN tree t ON f.parent_id = t.id
		)
		SELECT ` + folderCols + ` FROM tree ORDER BY path`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list folders")
	}
	defer rows.Close()

	folders := make([]Folder, 0)
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan folder")
		}
		folders = append(folders, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list folders")
	}
	return folders, nil
}

func (r *PostgresRepository) GetFolder(ctx context.Context, folderID string) (Folder, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+folderCols+` FROM document_folders WHERE id = $1`, folderID)
	f, err := scanFolder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Folder{}, apperr.NotFound("folder not found")
	}
	if err != nil {
		return Folder{}, apperr.Internal("failed to get folder")
	}
	return f, nil
}

func (r *PostgresRepository) UpdateFolder(ctx context.Context, folderID string, input UpdateFolderInput) (Folder, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.ParentID != nil {
		sets = append(sets, fmt.Sprintf("parent_id = $%d", argN))
		if *input.ParentID == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.ParentID)
		}
		argN++
	}

	if len(sets) == 0 {
		return r.GetFolder(ctx, folderID)
	}

	args = append(args, folderID)
	q := fmt.Sprintf(
		"UPDATE document_folders SET %s WHERE id = $%d RETURNING %s",
		strings.Join(sets, ", "), argN, folderCols,
	)

	row := r.db.QueryRowContext(ctx, q, args...)
	f, err := scanFolder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Folder{}, apperr.NotFound("folder not found")
	}
	if isUniqueViolation(err) {
		return Folder{}, apperr.InvalidInput("a folder with this name already exists here")
	}
	if isForeignKeyViolation(err) {
		return Folder{}, apperr.InvalidInput("parent folder not found")
	}
	if err != nil {
		return Folder{}, apperr.Internal("failed to update folder")
	}
	return f, nil
}

func (r *PostgresRepository) DeleteFolder(ctx context.Context, folderID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM document_folders WHERE id = $1`, folderID)
	if err != nil {
		return apperr.Internal("failed to delete folder")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperr.Internal("failed to delete folder")
	}
	if n == 0 {
		return apperr.NotFound("folder not found")
	}
	return nil
}

func (r *PostgresRepository) CollectSubtreeDocKeys(ctx context.Context, folderID string) ([]string, error) {
	if _, err := r.GetFolder(ctx, folderID); err != nil {
		return nil, err
	}

	const q = `
		WITH RECURSIVE subtree AS (
			SELECT id FROM document_folders WHERE id = $1
			UNION ALL
			SELECT f.id FROM document_folders f JOIN subtree s ON f.parent_id = s.id
		)
		SELECT d.s3_key FROM documents d JOIN subtree s ON d.folder_id = s.id`

	rows, err := r.db.QueryContext(ctx, q, folderID)
	if err != nil {
		return nil, apperr.Internal("failed to collect documents")
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, apperr.Internal("failed to scan document key")
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to collect documents")
	}
	return keys, nil
}

func (r *PostgresRepository) IsDescendant(ctx context.Context, folderID, maybeDescendant string) (bool, error) {
	const q = `
		WITH RECURSIVE subtree AS (
			SELECT id FROM document_folders WHERE id = $1
			UNION ALL
			SELECT f.id FROM document_folders f JOIN subtree s ON f.parent_id = s.id
		)
		SELECT EXISTS (SELECT 1 FROM subtree WHERE id = $2)`

	var found bool
	if err := r.db.QueryRowContext(ctx, q, folderID, maybeDescendant).Scan(&found); err != nil {
		return false, apperr.Internal("failed to check folder tree")
	}
	return found, nil
}

// --- documents ---

func (r *PostgresRepository) CreateDocument(ctx context.Context, d Document) (Document, error) {
	d.ID = id.New()
	d.Status = StatusPending
	d.CreatedAt = time.Now().UTC()
	d.UploadedAt = nil

	const q = `INSERT INTO documents (id, folder_id, label_id, name, s3_key, content_type, size_bytes, status, created_at)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, q,
		d.ID, d.FolderID, d.LabelID, d.Name, d.S3Key, d.ContentType, d.SizeBytes, d.Status, d.CreatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return Document{}, apperr.InvalidInput("folder not found")
		}
		return Document{}, apperr.Internal("failed to create document")
	}
	return d, nil
}

func (r *PostgresRepository) ListDocuments(ctx context.Context, filter ListFilter) ([]Document, error) {
	where := []string{"status = $1"}
	args := []any{StatusReady}
	argN := 2

	query := strings.TrimSpace(filter.Query)
	if query != "" {
		where = append(where, fmt.Sprintf("name ILIKE $%d", argN))
		args = append(args, "%"+escapeLike(query)+"%")
		argN++
	}

	if filter.LabelID != nil {
		where = append(where, fmt.Sprintf("label_id = $%d", argN))
		args = append(args, *filter.LabelID)
		argN++
	}

	// A query or a label filter searches every folder; otherwise the
	// listing is one folder's direct contents (nil FolderID = root).
	if !filter.global() {
		if filter.FolderID != nil {
			where = append(where, fmt.Sprintf("folder_id = $%d", argN))
			args = append(args, *filter.FolderID)
			argN++
		} else {
			where = append(where, "folder_id IS NULL")
		}
	}

	q := `SELECT ` + documentCols + ` FROM documents WHERE ` +
		strings.Join(where, " AND ") + ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, apperr.Internal("failed to list documents")
	}
	defer rows.Close()

	docs := make([]Document, 0)
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan document")
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list documents")
	}
	return docs, nil
}

func (r *PostgresRepository) GetDocument(ctx context.Context, docID string) (Document, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+documentCols+` FROM documents WHERE id = $1`, docID)
	d, err := scanDocument(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, apperr.NotFound("document not found")
	}
	if err != nil {
		return Document{}, apperr.Internal("failed to get document")
	}
	return d, nil
}

func (r *PostgresRepository) UpdateDocument(ctx context.Context, docID string, input UpdateDocumentInput) (Document, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argN := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, strings.TrimSpace(*input.Name))
		argN++
	}
	if input.FolderID != nil {
		sets = append(sets, fmt.Sprintf("folder_id = $%d", argN))
		if *input.FolderID == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.FolderID)
		}
		argN++
	}
	if input.LabelID != nil {
		sets = append(sets, fmt.Sprintf("label_id = $%d", argN))
		if *input.LabelID == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.LabelID)
		}
		argN++
	}

	if len(sets) == 0 {
		return r.GetDocument(ctx, docID)
	}

	args = append(args, docID)
	q := fmt.Sprintf(
		"UPDATE documents SET %s WHERE id = $%d RETURNING %s",
		strings.Join(sets, ", "), argN, documentCols,
	)

	row := r.db.QueryRowContext(ctx, q, args...)
	d, err := scanDocument(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, apperr.NotFound("document not found")
	}
	if isForeignKeyViolation(err) {
		return Document{}, apperr.InvalidInput("folder or label not found")
	}
	if err != nil {
		return Document{}, apperr.Internal("failed to update document")
	}
	return d, nil
}

func (r *PostgresRepository) MarkReady(ctx context.Context, docID string, size int64) (Document, error) {
	const q = `UPDATE documents SET status = $1, size_bytes = $2, uploaded_at = now()
	           WHERE id = $3 RETURNING ` + documentCols
	row := r.db.QueryRowContext(ctx, q, StatusReady, size, docID)
	d, err := scanDocument(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, apperr.NotFound("document not found")
	}
	if err != nil {
		return Document{}, apperr.Internal("failed to finalize document")
	}
	return d, nil
}

func (r *PostgresRepository) DeleteDocument(ctx context.Context, docID string) (string, error) {
	var key string
	err := r.db.QueryRowContext(ctx, `DELETE FROM documents WHERE id = $1 RETURNING s3_key`, docID).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.NotFound("document not found")
	}
	if err != nil {
		return "", apperr.Internal("failed to delete document")
	}
	return key, nil
}

// escapeLike neutralises LIKE wildcards in user search input (the query
// runs without an ESCAPE clause, so backslash is Postgres's default).
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
