package document

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type Service struct {
	repo Repository
	blob BlobStore
}

func NewService(repo Repository, blob BlobStore) *Service {
	return &Service{repo: repo, blob: blob}
}

// --- folders ---

func (s *Service) ListFolders(ctx context.Context) ([]Folder, error) {
	return s.repo.ListFolders(ctx)
}

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (Folder, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Folder{}, apperr.InvalidInput("name is required")
	}
	return s.repo.CreateFolder(ctx, Folder{Name: name, ParentID: input.ParentID})
}

func (s *Service) UpdateFolder(ctx context.Context, folderID string, input UpdateFolderInput) (Folder, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Folder{}, apperr.InvalidInput("name is required")
	}
	if input.ParentID != nil && *input.ParentID != "" {
		if *input.ParentID == folderID {
			return Folder{}, apperr.InvalidInput("a folder cannot be its own parent")
		}
		nested, err := s.repo.IsDescendant(ctx, folderID, *input.ParentID)
		if err != nil {
			return Folder{}, err
		}
		if nested {
			return Folder{}, apperr.InvalidInput("cannot move a folder into one of its own subfolders")
		}
	}
	return s.repo.UpdateFolder(ctx, folderID, input)
}

// DeleteFolder removes the folder and everything under it: the blobs of
// every descendant document first (so nothing is orphaned in the store),
// then the folder row, whose ON DELETE CASCADE clears the sub-rows.
func (s *Service) DeleteFolder(ctx context.Context, folderID string) error {
	keys, err := s.repo.CollectSubtreeDocKeys(ctx, folderID)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := s.blob.Delete(ctx, keys...); err != nil {
			return apperr.Internal("failed to delete stored files")
		}
	}
	return s.repo.DeleteFolder(ctx, folderID)
}

// --- documents ---

func (s *Service) ListDocuments(ctx context.Context, filter ListFilter) ([]Document, error) {
	return s.repo.ListDocuments(ctx, filter)
}

func (s *Service) CreateDocument(ctx context.Context, input CreateDocumentInput) (CreatedDocument, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CreatedDocument{}, apperr.InvalidInput("name is required")
	}
	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if input.SizeBytes < 0 {
		return CreatedDocument{}, apperr.InvalidInput("sizeBytes must not be negative")
	}

	key := id.New() + "/" + sanitizeFilename(name)

	created, err := s.repo.CreateDocument(ctx, Document{
		Name:        name,
		FolderID:    input.FolderID,
		ContentType: contentType,
		SizeBytes:   input.SizeBytes,
		S3Key:       key,
	})
	if err != nil {
		return CreatedDocument{}, err
	}

	uploadURL, err := s.blob.PresignPut(ctx, key, contentType)
	if err != nil {
		// Roll back the dangling pending row so a retry starts clean.
		if _, delErr := s.repo.DeleteDocument(ctx, created.ID); delErr != nil {
			log.Printf("document: failed to roll back pending row %s: %v", created.ID, delErr)
		}
		return CreatedDocument{}, apperr.Internal("failed to prepare upload")
	}

	return CreatedDocument{Document: created, UploadURL: uploadURL}, nil
}

// CompleteUpload confirms the browser's direct upload actually landed in
// the store and flips the row to ready.
func (s *Service) CompleteUpload(ctx context.Context, docID string) (Document, error) {
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return Document{}, err
	}
	if doc.Status == StatusReady {
		return doc, nil
	}

	size, ok, err := s.blob.Stat(ctx, doc.S3Key)
	if err != nil {
		return Document{}, apperr.Internal("failed to verify upload")
	}
	if !ok {
		return Document{}, apperr.InvalidInput("upload not found — the file was not received")
	}

	return s.repo.MarkReady(ctx, docID, size)
}

func (s *Service) UpdateDocument(ctx context.Context, docID string, input UpdateDocumentInput) (Document, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Document{}, apperr.InvalidInput("name is required")
	}
	return s.repo.UpdateDocument(ctx, docID, input)
}

// DownloadURL returns a short-lived URL the browser can GET the file from.
func (s *Service) DownloadURL(ctx context.Context, docID string) (string, error) {
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return "", err
	}
	if doc.Status != StatusReady {
		return "", apperr.InvalidInput("document is still uploading")
	}
	url, err := s.blob.PresignGet(ctx, doc.S3Key, doc.Name)
	if err != nil {
		return "", apperr.Internal("failed to prepare download")
	}
	return url, nil
}

func (s *Service) DeleteDocument(ctx context.Context, docID string) error {
	key, err := s.repo.DeleteDocument(ctx, docID)
	if err != nil {
		return err
	}
	// The row is gone; a stray blob is harmless, so don't fail the request.
	if err := s.blob.Delete(ctx, key); err != nil {
		log.Printf("document: failed to delete blob %s: %v", key, err)
	}
	return nil
}

var unsafeFilenameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// sanitizeFilename makes name safe as the last segment of a storage key.
func sanitizeFilename(name string) string {
	cleaned := unsafeFilenameChars.ReplaceAllString(name, "_")
	cleaned = strings.Trim(cleaned, "._-")
	if cleaned == "" {
		cleaned = "file"
	}
	if len(cleaned) > 120 {
		cleaned = cleaned[:120]
	}
	return cleaned
}
