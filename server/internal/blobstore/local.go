package blobstore

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// localStore keeps blob bytes on local disk. It stands in for S3 where no
// bucket is configured (local dev, the no-AWS deployment) and presents the
// same browser-direct upload/download flow: the URLs it hands out point
// back at this server's RoutePrefix endpoint, carrying an HMAC signature
// (over method|key|expiry) that ServeBlob verifies — so the endpoint needs
// no bearer token and the frontend code path is identical to S3's.
//
// Multipart uploads are simulated: each part is PUT to "<key>.parts/<n>";
// CompleteMultipart concatenates the parts in order into "<key>" and drops
// the parts directory.
type localStore struct {
	root      string
	secret    []byte
	publicURL string
}

// NewLocal roots the store at dir, signs URLs with secret, and builds them
// against publicURL (this API's externally reachable base URL, so the
// browser can reach the blob endpoint).
func NewLocal(dir string, secret []byte, publicURL string) (Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("blobstore: create local dir: %w", err)
	}
	return &localStore{
		root:      dir,
		secret:    secret,
		publicURL: strings.TrimRight(publicURL, "/"),
	}, nil
}

// blobRoutePrefix is deliberately not nested under an "/api/<feature>/"
// path that carries a "{id}" wildcard in the ServeMux, which would collide.
const blobRoutePrefix = "/api/blob/"

func (s *localStore) RoutePrefix() string { return blobRoutePrefix }

func (s *localStore) sign(method, key string, exp int64) string {
	mac := hmac.New(sha256.New, s.secret)
	fmt.Fprintf(mac, "%s|%s|%d", method, key, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *localStore) presignURL(method, key string, ttl time.Duration, extra url.Values) string {
	exp := time.Now().Add(ttl).Unix()
	q := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("exp", strconv.FormatInt(exp, 10))
	q.Set("sig", s.sign(method, key, exp))
	return s.publicURL + blobRoutePrefix + encodeKeyPath(key) + "?" + q.Encode()
}

func (s *localStore) PresignPut(_ context.Context, key, _ string) (string, error) {
	return s.presignURL(http.MethodPut, key, putURLTTL, nil), nil
}

func (s *localStore) PresignGet(_ context.Context, key, filename string, inline bool, ttlSeconds int) (string, error) {
	extra := url.Values{"filename": {filename}}
	if inline {
		extra.Set("inline", "1")
	}
	return s.presignURL(http.MethodGet, key, time.Duration(ttlSeconds)*time.Second, extra), nil
}

func (s *localStore) Stat(_ context.Context, key string) (int64, bool, error) {
	info, err := os.Stat(s.pathFor(key))
	if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return info.Size(), true, nil
}

func (s *localStore) Delete(_ context.Context, keys ...string) error {
	for _, k := range keys {
		if err := os.Remove(s.pathFor(k)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		_ = os.RemoveAll(s.pathFor(k) + ".parts")
		_ = os.Remove(filepath.Dir(s.pathFor(k)))
	}
	return nil
}

// CreateMultipart just returns a synthetic upload id; the local store keys
// parts by path, not by upload id.
func (s *localStore) CreateMultipart(_ context.Context, key, _ string) (string, error) {
	if err := os.MkdirAll(s.pathFor(key)+".parts", 0o755); err != nil {
		return "", fmt.Errorf("create multipart dir: %w", err)
	}
	return "local-" + hex.EncodeToString([]byte(key))[:16], nil
}

func (s *localStore) PresignUploadPart(_ context.Context, key, uploadID string, partNumber int) (string, error) {
	extra := url.Values{
		"uploadId": {uploadID},
		"part":     {strconv.Itoa(partNumber)},
	}
	return s.presignURL(http.MethodPut, key, putURLTTL, extra), nil
}

func (s *localStore) CompleteMultipart(_ context.Context, key, _ string, parts []Part) error {
	ordered := append([]Part(nil), parts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })

	dst := s.pathFor(key)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, p := range ordered {
		part, err := os.Open(filepath.Join(dst+".parts", strconv.Itoa(p.Number)))
		if err != nil {
			return fmt.Errorf("open part %d: %w", p.Number, err)
		}
		_, copyErr := io.Copy(out, part)
		part.Close()
		if copyErr != nil {
			return fmt.Errorf("append part %d: %w", p.Number, copyErr)
		}
	}
	return os.RemoveAll(dst + ".parts")
}

func (s *localStore) AbortMultipart(_ context.Context, key, _ string) error {
	return os.RemoveAll(s.pathFor(key) + ".parts")
}

// pathFor maps a store key to an on-disk path, guarding against traversal.
func (s *localStore) pathFor(key string) string {
	clean := filepath.Clean("/" + filepath.FromSlash(key))
	return filepath.Join(s.root, clean)
}

// ServeBlob handles PUT/GET on blobRoutePrefix. It is registered on the mux
// only when the local store is in use and is exempt from the bearer-auth
// middleware — it verifies its own signature instead.
func (s *localStore) ServeBlob(w http.ResponseWriter, r *http.Request) {
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
		if part := r.URL.Query().Get("part"); part != "" {
			s.putPart(w, r, key, part)
			return
		}
		s.put(w, r, key)
	case http.MethodGet:
		s.get(w, r, key)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *localStore) put(w http.ResponseWriter, r *http.Request, key string) {
	dst := s.pathFor(key)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	if err := writeFile(dst, r.Body, w); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *localStore) putPart(w http.ResponseWriter, r *http.Request, key, part string) {
	n, err := strconv.Atoi(part)
	if err != nil || n < 1 {
		http.Error(w, "bad part number", http.StatusBadRequest)
		return
	}
	partsDir := s.pathFor(key) + ".parts"
	if err := os.MkdirAll(partsDir, 0o755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(filepath.Join(partsDir, strconv.Itoa(n)), body, 0o644); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	sum := md5.Sum(body)
	// Quote the ETag like S3 does — the browser stores it verbatim and
	// hands it back to CompleteMultipart.
	w.Header().Set("ETag", `"`+hex.EncodeToString(sum[:])+`"`)
	w.WriteHeader(http.StatusOK)
}

func writeFile(dst string, body io.Reader, w http.ResponseWriter) error {
	f, err := os.Create(dst)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		os.Remove(dst)
		http.Error(w, "write failed", http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *localStore) get(w http.ResponseWriter, r *http.Request, key string) {
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

	if r.URL.Query().Get("inline") == "" {
		if filename := r.URL.Query().Get("filename"); filename != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		}
	}
	if ct := contentTypeForKey(key); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	// http.ServeContent handles Range requests (streaming + seeking); it
	// won't override a Content-Type we've already set.
	http.ServeContent(w, r, filepath.Base(key), info.ModTime(), f)
}

// contentTypeForKey maps the common media extensions the diary video
// recorder produces, since Go's mime table doesn't reliably carry .webm.
// "" means "let http.ServeContent sniff it".
func contentTypeForKey(key string) string {
	switch strings.ToLower(filepath.Ext(key)) {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	default:
		return ""
	}
}

// encodeKeyPath percent-encodes each segment of a slash-separated key.
func encodeKeyPath(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
