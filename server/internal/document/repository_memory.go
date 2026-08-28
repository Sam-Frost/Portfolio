package document

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu      sync.Mutex
	folders map[string]Folder
	docs    map[string]Document
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		folders: make(map[string]Folder),
		docs:    make(map[string]Document),
	}
}

// --- folders ---

func (r *MemoryRepository) folderNameTaken(parentID *string, name, excludeID string) bool {
	for otherID, f := range r.folders {
		if otherID == excludeID {
			continue
		}
		if samePtr(f.ParentID, parentID) && strings.EqualFold(f.Name, name) {
			return true
		}
	}
	return false
}

func (r *MemoryRepository) CreateFolder(_ context.Context, f Folder) (Folder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if f.ParentID != nil {
		if _, ok := r.folders[*f.ParentID]; !ok {
			return Folder{}, apperr.InvalidInput("parent folder not found")
		}
	}
	if r.folderNameTaken(f.ParentID, f.Name, "") {
		return Folder{}, apperr.InvalidInput("a folder with this name already exists here")
	}

	f.ID = id.New()
	f.CreatedAt = time.Now().UTC()
	r.folders[f.ID] = f
	return f, nil
}

func (r *MemoryRepository) ListFolders(_ context.Context) ([]Folder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	all := make([]Folder, 0, len(r.folders))
	for _, f := range r.folders {
		all = append(all, f)
	}

	// Parent-before-child, siblings by name — a stable pre-order walk.
	byParent := map[string][]Folder{}
	for _, f := range all {
		key := ""
		if f.ParentID != nil {
			key = *f.ParentID
		}
		byParent[key] = append(byParent[key], f)
	}
	for k := range byParent {
		sort.Slice(byParent[k], func(i, j int) bool { return byParent[k][i].Name < byParent[k][j].Name })
	}

	ordered := make([]Folder, 0, len(all))
	var walk func(parentKey string)
	walk = func(parentKey string) {
		for _, f := range byParent[parentKey] {
			ordered = append(ordered, f)
			walk(f.ID)
		}
	}
	walk("")
	return ordered, nil
}

func (r *MemoryRepository) GetFolder(_ context.Context, folderID string) (Folder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.folders[folderID]
	if !ok {
		return Folder{}, apperr.NotFound("folder not found")
	}
	return f, nil
}

func (r *MemoryRepository) UpdateFolder(_ context.Context, folderID string, input UpdateFolderInput) (Folder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.folders[folderID]
	if !ok {
		return Folder{}, apperr.NotFound("folder not found")
	}

	if input.ParentID != nil {
		if *input.ParentID == "" {
			f.ParentID = nil
		} else {
			if _, ok := r.folders[*input.ParentID]; !ok {
				return Folder{}, apperr.InvalidInput("parent folder not found")
			}
			parent := *input.ParentID
			f.ParentID = &parent
		}
	}
	if input.Name != nil {
		f.Name = strings.TrimSpace(*input.Name)
	}
	if r.folderNameTaken(f.ParentID, f.Name, folderID) {
		return Folder{}, apperr.InvalidInput("a folder with this name already exists here")
	}

	r.folders[folderID] = f
	return f, nil
}

func (r *MemoryRepository) DeleteFolder(_ context.Context, folderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.folders[folderID]; !ok {
		return apperr.NotFound("folder not found")
	}

	victims := r.subtreeIDs(folderID)
	for _, fid := range victims {
		delete(r.folders, fid)
	}
	victimSet := map[string]bool{}
	for _, fid := range victims {
		victimSet[fid] = true
	}
	for did, d := range r.docs {
		if d.FolderID != nil && victimSet[*d.FolderID] {
			delete(r.docs, did)
		}
	}
	return nil
}

func (r *MemoryRepository) CollectSubtreeDocKeys(_ context.Context, folderID string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.folders[folderID]; !ok {
		return nil, apperr.NotFound("folder not found")
	}

	inSubtree := map[string]bool{}
	for _, fid := range r.subtreeIDs(folderID) {
		inSubtree[fid] = true
	}

	keys := make([]string, 0)
	for _, d := range r.docs {
		if d.FolderID != nil && inSubtree[*d.FolderID] {
			keys = append(keys, d.S3Key)
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func (r *MemoryRepository) IsDescendant(_ context.Context, folderID, maybeDescendant string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, fid := range r.subtreeIDs(folderID) {
		if fid == maybeDescendant {
			return true, nil
		}
	}
	return false, nil
}

// subtreeIDs returns folderID and every folder nested beneath it. Caller
// holds the lock.
func (r *MemoryRepository) subtreeIDs(folderID string) []string {
	out := []string{folderID}
	for i := 0; i < len(out); i++ {
		for cid, f := range r.folders {
			if f.ParentID != nil && *f.ParentID == out[i] {
				out = append(out, cid)
			}
		}
	}
	return out
}

// --- documents ---

func (r *MemoryRepository) CreateDocument(_ context.Context, d Document) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if d.FolderID != nil {
		if _, ok := r.folders[*d.FolderID]; !ok {
			return Document{}, apperr.InvalidInput("folder not found")
		}
	}

	d.ID = id.New()
	d.Status = StatusPending
	d.CreatedAt = time.Now().UTC()
	d.UploadedAt = nil
	r.docs[d.ID] = d
	return d, nil
}

func (r *MemoryRepository) ListDocuments(_ context.Context, filter ListFilter) ([]Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := strings.TrimSpace(strings.ToLower(filter.Query))

	out := make([]Document, 0)
	for _, d := range r.docs {
		if d.Status != StatusReady {
			continue
		}
		if filter.LabelID != nil && (d.LabelID == nil || *d.LabelID != *filter.LabelID) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(d.Name), query) {
			continue
		}
		if !filter.global() && !sameFolderScope(d.FolderID, filter.FolderID) {
			continue
		}
		out = append(out, d)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *MemoryRepository) GetDocument(_ context.Context, docID string) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.docs[docID]
	if !ok {
		return Document{}, apperr.NotFound("document not found")
	}
	return d, nil
}

func (r *MemoryRepository) UpdateDocument(_ context.Context, docID string, input UpdateDocumentInput) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.docs[docID]
	if !ok {
		return Document{}, apperr.NotFound("document not found")
	}

	if input.Name != nil {
		d.Name = strings.TrimSpace(*input.Name)
	}
	if input.FolderID != nil {
		if *input.FolderID == "" {
			d.FolderID = nil
		} else {
			if _, ok := r.folders[*input.FolderID]; !ok {
				return Document{}, apperr.InvalidInput("folder not found")
			}
			folder := *input.FolderID
			d.FolderID = &folder
		}
	}
	if input.LabelID != nil {
		if *input.LabelID == "" {
			d.LabelID = nil
		} else {
			label := *input.LabelID
			d.LabelID = &label
		}
	}

	r.docs[docID] = d
	return d, nil
}

func (r *MemoryRepository) MarkReady(_ context.Context, docID string, size int64) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.docs[docID]
	if !ok {
		return Document{}, apperr.NotFound("document not found")
	}
	now := time.Now().UTC()
	d.Status = StatusReady
	d.SizeBytes = size
	d.UploadedAt = &now
	r.docs[docID] = d
	return d, nil
}

func (r *MemoryRepository) DeleteDocument(_ context.Context, docID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.docs[docID]
	if !ok {
		return "", apperr.NotFound("document not found")
	}
	delete(r.docs, docID)
	return d.S3Key, nil
}

// samePtr reports whether two *string point at the same value (both nil
// counts as equal).
func samePtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// sameFolderScope reports whether a document in folder docFolder belongs in
// a listing scoped to filterFolder (nil filterFolder = root).
func sameFolderScope(docFolder, filterFolder *string) bool {
	return samePtr(docFolder, filterFolder)
}
