package blobstore

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// putSigned replays a presigned local-store URL against ServeBlob.
func putSigned(t *testing.T, s Store, url string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	local := s.(*localStore)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	local.ServeBlob(rec, req)
	return rec
}

func TestLocalMultipart_ConcatenatesPartsInOrder(t *testing.T) {
	dir := t.TempDir()
	s, err := NewLocal(dir, []byte("secret"), "http://localhost:8080")
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	ctx := context.Background()
	key := "vid1/clip.webm"

	uploadID, err := s.CreateMultipart(ctx, key, "video/webm")
	if err != nil {
		t.Fatalf("CreateMultipart: %v", err)
	}

	// Upload three parts out of order (2, then 1, then 3).
	chunks := map[int][]byte{1: []byte("hello "), 2: []byte("brave "), 3: []byte("world")}
	var parts []Part
	for _, n := range []int{2, 1, 3} {
		url, err := s.PresignUploadPart(ctx, key, uploadID, n)
		if err != nil {
			t.Fatalf("PresignUploadPart(%d): %v", n, err)
		}
		rec := putSigned(t, s, url, chunks[n])
		if rec.Code != http.StatusOK {
			t.Fatalf("part %d PUT: status %d", n, rec.Code)
		}
		etag := strings.Trim(rec.Header().Get("ETag"), `"`)
		if etag == "" {
			t.Fatalf("part %d: no ETag returned", n)
		}
		parts = append(parts, Part{Number: n, ETag: etag})
	}

	if err := s.CompleteMultipart(ctx, key, uploadID, parts); err != nil {
		t.Fatalf("CompleteMultipart: %v", err)
	}

	got, err := os.ReadFile(dir + "/" + key)
	if err != nil {
		t.Fatalf("read assembled object: %v", err)
	}
	if string(got) != "hello brave world" {
		t.Errorf("assembled = %q, want %q", got, "hello brave world")
	}
	if _, err := os.Stat(dir + "/" + key + ".parts"); !os.IsNotExist(err) {
		t.Errorf("parts dir should be gone after complete")
	}

	size, ok, err := s.Stat(ctx, key)
	if err != nil || !ok {
		t.Fatalf("Stat: size=%d ok=%v err=%v", size, ok, err)
	}
	if size != int64(len("hello brave world")) {
		t.Errorf("Stat size = %d, want %d", size, len("hello brave world"))
	}
}

func TestLocalMultipart_AbortRemovesParts(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocal(dir, []byte("secret"), "http://localhost:8080")
	ctx := context.Background()
	key := "vid2/clip.mp4"

	uploadID, _ := s.CreateMultipart(ctx, key, "video/mp4")
	url, _ := s.PresignUploadPart(ctx, key, uploadID, 1)
	putSigned(t, s, url, []byte("partial"))

	if err := s.AbortMultipart(ctx, key, uploadID); err != nil {
		t.Fatalf("AbortMultipart: %v", err)
	}
	if _, err := os.Stat(dir + "/" + key + ".parts"); !os.IsNotExist(err) {
		t.Errorf("parts dir should be gone after abort")
	}
}

func TestLocalServeBlob_RejectsBadSignature(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocal(dir, []byte("secret"), "http://localhost:8080")
	local := s.(*localStore)

	req := httptest.NewRequest(http.MethodGet, blobRoutePrefix+"vid/clip.mp4?exp=9999999999&sig=deadbeef", nil)
	rec := httptest.NewRecorder()
	local.ServeBlob(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestLocalGet_ServesRangeForStreaming(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocal(dir, []byte("secret"), "http://localhost:8080")
	ctx := context.Background()
	key := "vid3/clip.mp4"

	putURL, _ := s.PresignPut(ctx, key, "video/mp4")
	putSigned(t, s, putURL, bytes.Repeat([]byte("x"), 100))

	getURL, _ := s.PresignGet(ctx, key, "clip.mp4", true, 60)
	local := s.(*localStore)
	req := httptest.NewRequest(http.MethodGet, getURL, nil)
	req.Header.Set("Range", "bytes=10-19")
	rec := httptest.NewRecorder()
	local.ServeBlob(rec, req)

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if len(body) != 10 {
		t.Errorf("range body len = %d, want 10", len(body))
	}
	if ct := rec.Header().Get("Content-Type"); ct != "video/mp4" {
		t.Errorf("Content-Type = %q, want video/mp4", ct)
	}
}
