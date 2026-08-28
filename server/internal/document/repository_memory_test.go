package document

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func ptr(s string) *string { return &s }

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.NotFound", err)
	}
}

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func mustFolder(t *testing.T, r *MemoryRepository, name string, parent *string) Folder {
	t.Helper()
	f, err := r.CreateFolder(context.Background(), Folder{Name: name, ParentID: parent})
	if err != nil {
		t.Fatalf("CreateFolder(%q): %v", name, err)
	}
	return f
}

func mustReadyDoc(t *testing.T, r *MemoryRepository, name string, folder *string, label *string) Document {
	t.Helper()
	d, err := r.CreateDocument(context.Background(), Document{Name: name, FolderID: folder, LabelID: label, S3Key: name + "-key"})
	if err != nil {
		t.Fatalf("CreateDocument(%q): %v", name, err)
	}
	if _, err := r.MarkReady(context.Background(), d.ID, 10); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}
	d, _ = r.GetDocument(context.Background(), d.ID)
	return d
}

func TestMemory_FolderNameUniqueWithinParent(t *testing.T) {
	r := NewMemoryRepository()
	root := mustFolder(t, r, "Taxes", nil)

	if _, err := r.CreateFolder(context.Background(), Folder{Name: "taxes"}); err == nil {
		t.Fatal("expected duplicate root folder name to be rejected")
	}
	// Same name under a different parent is fine.
	if _, err := r.CreateFolder(context.Background(), Folder{Name: "Taxes", ParentID: &root.ID}); err != nil {
		t.Fatalf("nested folder with same name should be allowed: %v", err)
	}
}

func TestMemory_ListFoldersParentBeforeChild(t *testing.T) {
	r := NewMemoryRepository()
	a := mustFolder(t, r, "A", nil)
	mustFolder(t, r, "A-child", &a.ID)
	mustFolder(t, r, "B", nil)

	got, err := r.ListFolders(context.Background())
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if len(got) != 3 || got[0].Name != "A" || got[1].Name != "A-child" || got[2].Name != "B" {
		t.Fatalf("order = %v, want [A, A-child, B]", names(got))
	}
}

func TestMemory_ListDocumentsFiltersFolderLabelAndQuery(t *testing.T) {
	r := NewMemoryRepository()
	folder := mustFolder(t, r, "Work", nil)
	label := "label-1"

	inFolder := mustReadyDoc(t, r, "report.pdf", &folder.ID, &label)
	mustReadyDoc(t, r, "notes.txt", &folder.ID, nil)
	mustReadyDoc(t, r, "root-report.pdf", nil, nil)
	// pending doc must never show up
	if _, err := r.CreateDocument(context.Background(), Document{Name: "pending.pdf", FolderID: &folder.ID, S3Key: "p"}); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	byFolder, _ := r.ListDocuments(context.Background(), ListFilter{FolderID: &folder.ID})
	if len(byFolder) != 2 {
		t.Fatalf("folder listing = %v, want 2 ready docs", docNames(byFolder))
	}

	byRoot, _ := r.ListDocuments(context.Background(), ListFilter{})
	if len(byRoot) != 1 || byRoot[0].Name != "root-report.pdf" {
		t.Fatalf("root listing = %v, want [root-report.pdf]", docNames(byRoot))
	}

	byLabel, _ := r.ListDocuments(context.Background(), ListFilter{FolderID: &folder.ID, LabelID: &label})
	if len(byLabel) != 1 || byLabel[0].ID != inFolder.ID {
		t.Fatalf("label listing = %v, want [report.pdf]", docNames(byLabel))
	}

	byQuery, _ := r.ListDocuments(context.Background(), ListFilter{Query: "report"})
	if len(byQuery) != 2 {
		t.Fatalf("query listing = %v, want the 2 *report* docs across folders", docNames(byQuery))
	}
}

func TestMemory_DeleteFolderCascadesAndReturnsKeys(t *testing.T) {
	r := NewMemoryRepository()
	parent := mustFolder(t, r, "Parent", nil)
	child := mustFolder(t, r, "Child", &parent.ID)
	mustReadyDoc(t, r, "a.pdf", &parent.ID, nil)
	mustReadyDoc(t, r, "b.pdf", &child.ID, nil)

	keys, err := r.CollectSubtreeDocKeys(context.Background(), parent.ID)
	if err != nil {
		t.Fatalf("CollectSubtreeDocKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("keys = %v, want 2", keys)
	}

	if err := r.DeleteFolder(context.Background(), parent.ID); err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}
	if _, err := r.GetFolder(context.Background(), child.ID); err == nil {
		t.Fatal("child folder should have been cascade-deleted")
	}
	remaining, _ := r.ListDocuments(context.Background(), ListFilter{Query: ""})
	if len(remaining) != 0 {
		t.Fatalf("documents survived folder delete: %v", docNames(remaining))
	}
}

func TestMemory_IsDescendant(t *testing.T) {
	r := NewMemoryRepository()
	a := mustFolder(t, r, "A", nil)
	b := mustFolder(t, r, "B", &a.ID)
	c := mustFolder(t, r, "C", &b.ID)
	other := mustFolder(t, r, "Other", nil)

	if ok, _ := r.IsDescendant(context.Background(), a.ID, c.ID); !ok {
		t.Error("C should be a descendant of A")
	}
	if ok, _ := r.IsDescendant(context.Background(), a.ID, a.ID); !ok {
		t.Error("a folder counts as its own descendant for cycle checks")
	}
	if ok, _ := r.IsDescendant(context.Background(), b.ID, other.ID); ok {
		t.Error("Other is not under B")
	}
}

func TestMemory_DeleteDocumentReturnsKey(t *testing.T) {
	r := NewMemoryRepository()
	d := mustReadyDoc(t, r, "x.pdf", nil, nil)

	key, err := r.DeleteDocument(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}
	if key != d.S3Key {
		t.Fatalf("key = %q, want %q", key, d.S3Key)
	}
	_, err = r.GetDocument(context.Background(), d.ID)
	assertNotFound(t, err)
}

func names(fs []Folder) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Name
	}
	return out
}

func docNames(ds []Document) []string {
	out := make([]string, len(ds))
	for i, d := range ds {
		out[i] = d.Name
	}
	return out
}
