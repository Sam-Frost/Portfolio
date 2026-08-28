package document

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// localBlobStore keeps document bytes on local disk. It stands in for S3
// where no bucket is configured (local dev, the no-AWS friend deployment).
//
// It still presents the same browser-direct upload/download flow as S3: the
// URLs it hands out point back at this server's /api/documents/_blob/
// endpoint, carrying an HMAC signature (over method|key|expiry) that
// ServeBlob verifies — so the endpoint needs no bearer token and the
// frontend code path is identical to the S3 one.
type localBlobStore struct {
	root      string
	secret    []byte
	publicURL string // this server's externally reachable base URL
}

// NewLocalBlobStore roots the store at dir, signs URLs with secret, and
// builds them against publicURL (the API's own external base URL, so the
// browser can reach the _blob endpoint).
func NewLocalBlobStore(dir string, secret []byte, publicURL string) (BlobStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("document: create local blob dir: %w", err)
	}
	return &localBlobStore{
		root:      dir,
		secret:    secret,
		publicURL: strings.TrimRight(publicURL, "/"),
	}, nil
}

// blobRoutePrefix is deliberately not nested under /api/documents/ — a
// path there would collide with the "/api/documents/{id}/..." patterns in
// the ServeMux (the {id} wildcard matches any first segment).
const blobRoutePrefix = "/api/document-blob/"

func (s *localBlobStore) sign(method, key string, exp int64) string {
	mac := hmac.New(sha256.New, s.secret)
	fmt.Fprintf(mac, "%s|%s|%d", method, key, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *localBlobStore) presign(method, key string, ttl time.Duration, extraQuery url.Values) string {
	exp := time.Now().Add(ttl).Unix()
	q := url.Values{}
	for k, vs := range extraQuery {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("exp", strconv.FormatInt(exp, 10))
	q.Set("sig", s.sign(method, key, exp))
	// key segments are already path-safe (id/sanitised-name) but encode to be safe.
	return s.publicURL + blobRoutePrefix + encodeKeyPath(key) + "?" + q.Encode()
}

func (s *localBlobStore) PresignPut(_ context.Context, key, _ string) (string, error) {
	return s.presign(http.MethodPut, key, putURLTTL, nil), nil
}

func (s *localBlobStore) PresignGet(_ context.Context, key, filename string) (string, error) {
	return s.presign(http.MethodGet, key, getURLTTL, url.Values{"filename": {filename}}), nil
}

func (s *localBlobStore) Stat(_ context.Context, key string) (int64, bool, error) {
	info, err := os.Stat(s.pathFor(key))
	if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return info.Size(), true, nil
}

func (s *localBlobStore) Delete(_ context.Context, keys ...string) error {
	for _, k := range keys {
		if err := os.Remove(s.pathFor(k)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// Best-effort: drop the now-empty per-document directory.
		_ = os.Remove(filepath.Dir(s.pathFor(k)))
	}
	return nil
}

// pathFor maps a store key to an on-disk path, guarding against traversal.
func (s *localBlobStore) pathFor(key string) string {
	clean := filepath.Clean("/" + filepath.FromSlash(key))
	return filepath.Join(s.root, clean)
}

// ServeBlob handles PUT/GET /api/documents/_blob/{key...}. It is registered
// on the mux only when the local store is in use and is exempt from the
// bearer-auth middleware — it verifies its own signature instead.
func (s *localBlobStore) ServeBlob(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, blobRoutePrefix)
	key, err := url.PathUnescape(key)
	if err != nil || key == "" {
		http.Error(w, "bad key", http.StatusBadRequest)
		return
	}

	exp, err := strconv.ParseInt(r.URL.Query().Get("exp"), 10, 64)
	if err != nil || time.Now().Unix() > exp {
		http.Error(w, "link expired", http.StatusForbidden)
		return
	}
	want := s.sign(r.Method, key, exp)
	if !hmac.Equal([]byte(want), []byte(r.URL.Query().Get("sig"))) {
		http.Error(w, "bad signature", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.put(w, r, key)
	case http.MethodGet:
		s.get(w, r, key)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *localBlobStore) put(w http.ResponseWriter, r *http.Request, key string) {
	dst := s.pathFor(key)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	f, err := os.Create(dst)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		os.Remove(dst)
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *localBlobStore) get(w http.ResponseWriter, r *http.Request, key string) {
	f, err := os.Open(s.pathFor(key))
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	if filename := r.URL.Query().Get("filename"); filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, "", info.ModTime(), f)
}

// encodeKeyPath percent-encodes each segment of a slash-separated key.
func encodeKeyPath(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
