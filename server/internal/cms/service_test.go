package cms

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// fakePublisher records what it was asked to publish and can be told to fail.
type fakePublisher struct {
	enabled bool
	failErr error
	calls   int
	lastVer int
}

func (f *fakePublisher) Enabled() bool { return f.enabled }

func (f *fakePublisher) Publish(_ context.Context, version int, _ []byte) error {
	f.calls++
	f.lastVer = version
	return f.failErr
}

func newTestService(pub Publisher) (*Service, *MemoryRepository) {
	repo := NewMemoryRepository()
	return NewService(repo, pub), repo
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperr.Error, got %v", err)
	}
	if appErr.Kind != want {
		t.Fatalf("expected kind %v, got %v (%s)", want, appErr.Kind, appErr.Message)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Brocode Crypto Exchange": "brocode-crypto-exchange",
		"  Trailing  Spaces  ":    "trailing-spaces",
		"Shadow-Link":             "shadow-link",
		"C++ & Rust!!!":           "c-rust",
		"already-a-slug":          "already-a-slug",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveSlugRejectsBadExplicitSlug(t *testing.T) {
	if _, err := resolveSlug("Not A Slug", "title"); err == nil {
		t.Fatal("expected error for invalid explicit slug")
	}
	got, err := resolveSlug("", "My Great Post")
	if err != nil || got != "my-great-post" {
		t.Fatalf("resolveSlug fallback = %q, %v", got, err)
	}
}

func TestCreateProjectAssignsIncrementingOrder(t *testing.T) {
	svc, _ := newTestService(&fakePublisher{})
	ctx := context.Background()

	p1, err := svc.CreateProject(ctx, CreateProjectInput{Title: "One"})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Two"})
	if err != nil {
		t.Fatal(err)
	}
	if p1.Order != 0 || p2.Order != 1 {
		t.Fatalf("orders = %d, %d; want 0, 1", p1.Order, p2.Order)
	}
}

func TestCreateProjectDuplicateSlugRejected(t *testing.T) {
	svc, _ := newTestService(&fakePublisher{})
	ctx := context.Background()

	if _, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Dup", Slug: "dup"}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Other", Slug: "dup"})
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestCreateProjectInvalidURLRejected(t *testing.T) {
	svc, _ := newTestService(&fakePublisher{})
	_, err := svc.CreateProject(context.Background(), CreateProjectInput{Title: "X", Github: "not a url"})
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestPublishDisabledReturnsConflict(t *testing.T) {
	svc, _ := newTestService(&fakePublisher{enabled: false})
	_, err := svc.Publish(context.Background())
	assertKind(t, err, apperr.KindConflict)
}

func TestStatusLifecycle(t *testing.T) {
	pub := &fakePublisher{enabled: true}
	svc, _ := newTestService(pub)
	ctx := context.Background()

	// Fresh: never published → everything counts as unpublished.
	st, err := svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !st.NeverPublished || !st.HasUnpublishedChanges {
		t.Fatalf("fresh status = %+v", st)
	}

	if _, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Alpha"}); err != nil {
		t.Fatal(err)
	}

	// Publish → clean.
	if _, err := svc.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	if pub.calls != 1 || pub.lastVer != 1 {
		t.Fatalf("publisher calls=%d ver=%d", pub.calls, pub.lastVer)
	}
	st, err = svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.NeverPublished || st.HasUnpublishedChanges || len(st.ChangedSections) != 0 {
		t.Fatalf("post-publish status = %+v", st)
	}

	// Edit a project → only "projects" is dirty.
	list, _ := svc.ListProjects(ctx)
	newTitle := "Alpha v2"
	if _, err := svc.UpdateProject(ctx, list[0].ID, UpdateProjectInput{Title: &newTitle}); err != nil {
		t.Fatal(err)
	}
	st, err = svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !st.HasUnpublishedChanges || len(st.ChangedSections) != 1 || st.ChangedSections[0] != SectionProjects {
		t.Fatalf("post-edit status = %+v", st)
	}

	// Re-publish → version bumps, clean again.
	if _, err := svc.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	if pub.lastVer != 2 {
		t.Fatalf("second publish version = %d, want 2", pub.lastVer)
	}
	st, _ = svc.Status(ctx)
	if st.HasUnpublishedChanges {
		t.Fatalf("status still dirty after re-publish: %+v", st)
	}
}

func TestPublishFailureRecordedAndBaselineUnchanged(t *testing.T) {
	pub := &fakePublisher{enabled: true, failErr: errors.New("s3 down")}
	svc, _ := newTestService(pub)
	ctx := context.Background()

	if _, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Alpha"}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Publish(ctx)
	assertKind(t, err, apperr.KindInternal)

	// A failed publish must not become the diff baseline — the site is still
	// stale, so Status should still report unpublished changes.
	st, err := svc.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !st.HasUnpublishedChanges {
		t.Fatalf("expected still-dirty after failed publish, got %+v", st)
	}
	if st.LastPublishStatus != StatusFailed || st.LastPublishError == nil {
		t.Fatalf("expected failed last-publish info, got %+v", st)
	}
}

func TestResaveIdenticalValuesIsNotAChange(t *testing.T) {
	pub := &fakePublisher{enabled: true}
	svc, _ := newTestService(pub)
	ctx := context.Background()

	p, err := svc.CreateProject(ctx, CreateProjectInput{Title: "Alpha", Description: "d"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Publish(ctx); err != nil {
		t.Fatal(err)
	}

	same := "Alpha"
	if _, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{Title: &same}); err != nil {
		t.Fatal(err)
	}
	st, _ := svc.Status(ctx)
	if st.HasUnpublishedChanges {
		t.Fatalf("re-saving identical values reported a change: %+v", st)
	}
}
