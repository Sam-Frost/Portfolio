// Package document owns the Document Storage feature: a folder tree plus
// document metadata. The bytes never pass through this server — the browser
// uploads and downloads directly against a blob store (S3 in prod, local
// disk in dev, see blobstore.go) using short-lived pre-signed URLs this
// package hands out.
//
// Feature-sliced like internal/todo: model / repository / repository_memory
// / repository_postgres / service / handler. Errors via internal/apperr,
// responses via internal/httpx, IDs via internal/id.
package document

import (
	"strings"
	"time"
)

// Folder is one node of the folder tree. ParentID is nil for a root-level
// folder.
type Folder struct {
	ID        string    `json:"id"`
	ParentID  *string   `json:"parentId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Document status values.
const (
	StatusPending = "pending" // created, awaiting the browser's upload + complete
	StatusReady   = "ready"   // blob confirmed present in the store
)

// Document is the metadata row. FolderID is nil when the document sits at
// the root; LabelID is nil when unlabeled. S3Key is the blob store key and
// is never exposed to clients.
type Document struct {
	ID          string     `json:"id"`
	FolderID    *string    `json:"folderId"`
	LabelID     *string    `json:"labelId"`
	Name        string     `json:"name"`
	S3Key       string     `json:"-"`
	ContentType string     `json:"contentType"`
	SizeBytes   int64      `json:"sizeBytes"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UploadedAt  *time.Time `json:"uploadedAt"`
}

type CreateFolderInput struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parentId"`
}

// UpdateFolderInput is a partial update: nil fields are left unchanged. An
// empty-string ParentID means "move to the root" (distinct from nil =
// "leave unchanged"), mirroring internal/todo's empty-string-clears rule.
type UpdateFolderInput struct {
	Name     *string `json:"name"`
	ParentID *string `json:"parentId"`
}

type CreateDocumentInput struct {
	Name        string  `json:"name"`
	FolderID    *string `json:"folderId"`
	ContentType string  `json:"contentType"`
	SizeBytes   int64   `json:"sizeBytes"`
}

// UpdateDocumentInput is a partial update. Empty-string FolderID / LabelID
// means "clear it" (move to root / remove label).
type UpdateDocumentInput struct {
	Name     *string `json:"name"`
	FolderID *string `json:"folderId"`
	LabelID  *string `json:"labelId"`
}

// ListFilter narrows a document listing. FolderID scopes the listing to a
// folder's direct contents (nil = root). A Query or a LabelID makes the
// search global instead — FolderID is ignored — matching how the Todos
// label filter works and how a file-manager search behaves.
type ListFilter struct {
	FolderID *string
	LabelID  *string
	Query    string
}

// global reports whether the filter should span all folders rather than a
// single folder's contents.
func (f ListFilter) global() bool {
	return strings.TrimSpace(f.Query) != "" || f.LabelID != nil
}

// CreatedDocument is the create response: the pending row plus the URL the
// browser PUTs the file bytes to.
type CreatedDocument struct {
	Document  Document `json:"document"`
	UploadURL string   `json:"uploadUrl"`
}
