package document

import "context"

// Repository is the persistence boundary for folders and document
// metadata, shaped for the Postgres implementation
// (repository_postgres.go); repository_memory.go is the in-memory stand-in
// used by tests.
//
// Update* take a full Update*Input (not a mutation closure) so a SQL
// implementation can build a real "SET" clause, matching internal/todo.
type Repository interface {
	// --- folders ---

	CreateFolder(ctx context.Context, f Folder) (Folder, error)
	// ListFolders returns the whole tree, each parent before its children,
	// siblings by name. The caller assembles the nesting.
	ListFolders(ctx context.Context) ([]Folder, error)
	GetFolder(ctx context.Context, id string) (Folder, error)
	UpdateFolder(ctx context.Context, id string, input UpdateFolderInput) (Folder, error)
	// DeleteFolder removes the folder; the schema's ON DELETE CASCADE
	// clears descendant folders and their documents. Callers delete the
	// blobs first via CollectSubtreeDocKeys.
	DeleteFolder(ctx context.Context, id string) error
	// CollectSubtreeDocKeys returns the S3 keys of every document in the
	// folder and all its descendants, so the service can purge the blobs
	// before the DB cascade drops the rows.
	CollectSubtreeDocKeys(ctx context.Context, folderID string) ([]string, error)
	// IsDescendant reports whether maybeDescendant is folderID itself or
	// nested anywhere beneath it — used to reject a move that would create
	// a cycle.
	IsDescendant(ctx context.Context, folderID, maybeDescendant string) (bool, error)

	// --- documents ---

	CreateDocument(ctx context.Context, d Document) (Document, error)
	// ListDocuments returns ready documents matching filter, newest first.
	ListDocuments(ctx context.Context, filter ListFilter) ([]Document, error)
	GetDocument(ctx context.Context, id string) (Document, error)
	UpdateDocument(ctx context.Context, id string, input UpdateDocumentInput) (Document, error)
	// MarkReady flips a pending document to ready with its confirmed size.
	MarkReady(ctx context.Context, id string, size int64) (Document, error)
	// DeleteDocument removes the row and returns its S3 key so the caller
	// can delete the blob.
	DeleteDocument(ctx context.Context, id string) (s3Key string, err error)
}
