package document

import (
	"context"
	"sync"
	"testing"
)

// fakeBlob is an in-memory BlobStore for service tests: PresignPut/Get just
// echo the key, Stat reports whatever was "uploaded" via markUploaded.
type fakeBlob struct {
	mu       sync.Mutex
	uploaded map[string]int64
	deleted  []string
}

func newFakeBlob() *fakeBlob { return &fakeBlob{uploaded: map[string]int64{}} }

func (f *fakeBlob) PresignPut(_ context.Context, key, _ string) (string, error) {
	return "https://blob.test/put/" + key, nil
}
func (f *fakeBlob) PresignGet(_ context.Context, key, _ string) (string, error) {
	return "https://blob.test/get/" + key, nil
}
func (f *fakeBlob) Stat(_ context.Context, key string) (int64, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	size, ok := f.uploaded[key]
	return size, ok, nil
}
func (f *fakeBlob) Delete(_ context.Context, keys ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, keys...)
	return nil
}
func (f *fakeBlob) markUploaded(key string, size int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.uploaded[key] = size
}

func newTestService() (*Service, *MemoryRepository, *fakeBlob) {
	repo := NewMemoryRepository()
	blob := newFakeBlob()
	return NewService(repo, blob), repo, blob
}

func TestService_CreateFolderRejectsBlankName(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "  "})
	assertInvalidInput(t, err)
}

func TestService_UpdateFolderRejectsMoveIntoOwnSubtree(t *testing.T) {
	svc, _, _ := newTestService()
	parent, err := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	child, err := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "Child", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}

	_, err = svc.UpdateFolder(context.Background(), parent.ID, UpdateFolderInput{ParentID: &child.ID})
	assertInvalidInput(t, err)

	_, err = svc.UpdateFolder(context.Background(), parent.ID, UpdateFolderInput{ParentID: &parent.ID})
	assertInvalidInput(t, err)
}

func TestService_CreateDocumentReturnsUploadURLForPendingRow(t *testing.T) {
	svc, _, _ := newTestService()

	created, err := svc.CreateDocument(context.Background(), CreateDocumentInput{Name: "  Invoice.pdf  ", ContentType: ""})
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	if created.Document.Name != "Invoice.pdf" {
		t.Errorf("Name = %q, want trimmed", created.Document.Name)
	}
	if created.Document.Status != StatusPending {
		t.Errorf("Status = %q, want pending", created.Document.Status)
	}
	if created.Document.ContentType != "application/octet-stream" {
		t.Errorf("ContentType = %q, want default", created.Document.ContentType)
	}
	if created.UploadURL == "" {
		t.Error("UploadURL is empty")
	}
}

func TestService_CompleteUploadRequiresBlobPresent(t *testing.T) {
	svc, _, blob := newTestService()
	created, err := svc.CreateDocument(context.Background(), CreateDocumentInput{Name: "a.pdf"})
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	if _, err := svc.CompleteUpload(context.Background(), created.Document.ID); err == nil {
		t.Fatal("CompleteUpload should fail before the blob is uploaded")
	}

	blob.markUploaded(created.Document.S3Key, 4321)
	done, err := svc.CompleteUpload(context.Background(), created.Document.ID)
	if err != nil {
		t.Fatalf("CompleteUpload: %v", err)
	}
	if done.Status != StatusReady || done.SizeBytes != 4321 {
		t.Fatalf("after complete: status=%q size=%d, want ready/4321", done.Status, done.SizeBytes)
	}
}

func TestService_DeleteDocumentAlsoDeletesBlob(t *testing.T) {
	svc, _, blob := newTestService()
	created, _ := svc.CreateDocument(context.Background(), CreateDocumentInput{Name: "a.pdf"})
	blob.markUploaded(created.Document.S3Key, 1)
	if _, err := svc.CompleteUpload(context.Background(), created.Document.ID); err != nil {
		t.Fatalf("CompleteUpload: %v", err)
	}

	if err := svc.DeleteDocument(context.Background(), created.Document.ID); err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}
	if len(blob.deleted) != 1 || blob.deleted[0] != created.Document.S3Key {
		t.Fatalf("blob deletes = %v, want [%s]", blob.deleted, created.Document.S3Key)
	}
}

func TestService_DeleteFolderPurgesDescendantBlobs(t *testing.T) {
	svc, repo, blob := newTestService()
	parent, _ := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "P"})
	child, _ := svc.CreateFolder(context.Background(), CreateFolderInput{Name: "C", ParentID: &parent.ID})
	d := mustReadyDoc(t, repo, "deep.pdf", &child.ID, nil)

	if err := svc.DeleteFolder(context.Background(), parent.ID); err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}
	if len(blob.deleted) != 1 || blob.deleted[0] != d.S3Key {
		t.Fatalf("blob deletes = %v, want [%s]", blob.deleted, d.S3Key)
	}
}
