package notepadlabel

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer wires a real Handler (backed by a fresh MemoryRepository)
// onto its own mux and starts it, so these tests exercise the full
// HTTP -> handler -> service -> repository stack end to end.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	NewHandler(NewService(NewMemoryRepository())).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func doJSON(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func decode[T any](t *testing.T, res *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return v
}

func TestHandler_FullLabelLifecycle(t *testing.T) {
	srv := newTestServer(t)

	// Create.
	createRes := doJSON(t, http.MethodPost, srv.URL+"/api/notepad-labels",
		CreateInput{Name: "Ideas", Color: "blue"})
	if createRes.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d", createRes.StatusCode, http.StatusCreated)
	}
	created := decode[Label](t, createRes)
	if created.Name != "Ideas" || created.Color != "blue" || created.ID == "" {
		t.Fatalf("created = %+v, want name/color set and non-empty ID", created)
	}

	// List includes it.
	list := decode[[]Label](t, doJSON(t, http.MethodGet, srv.URL+"/api/notepad-labels", nil))
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("List = %+v, want just %q", list, created.ID)
	}

	// Update recolors it.
	color := "purple"
	updateRes := doJSON(t, http.MethodPatch, srv.URL+"/api/notepad-labels/"+created.ID,
		UpdateInput{Color: &color})
	if updateRes.StatusCode != http.StatusOK {
		t.Fatalf("PATCH status = %d, want %d", updateRes.StatusCode, http.StatusOK)
	}
	if updated := decode[Label](t, updateRes); updated.Color != "purple" {
		t.Errorf("updated.Color = %q, want %q", updated.Color, "purple")
	}

	// Delete removes it.
	delRes := doJSON(t, http.MethodDelete, srv.URL+"/api/notepad-labels/"+created.ID, nil)
	if delRes.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", delRes.StatusCode, http.StatusNoContent)
	}
	if list := decode[[]Label](t, doJSON(t, http.MethodGet, srv.URL+"/api/notepad-labels", nil)); len(list) != 0 {
		t.Fatalf("List after delete = %+v, want empty", list)
	}
}

func TestHandler_CreateWithUnknownColorReturns400(t *testing.T) {
	srv := newTestServer(t)

	res := doJSON(t, http.MethodPost, srv.URL+"/api/notepad-labels",
		CreateInput{Name: "Ideas", Color: "chartreuse"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandler_CreateWithBlankNameReturns400(t *testing.T) {
	srv := newTestServer(t)

	res := doJSON(t, http.MethodPost, srv.URL+"/api/notepad-labels",
		CreateInput{Name: "   ", Color: "blue"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandler_CreateDuplicateNameReturns400(t *testing.T) {
	srv := newTestServer(t)

	first := doJSON(t, http.MethodPost, srv.URL+"/api/notepad-labels", CreateInput{Name: "Ideas", Color: "blue"})
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first POST status = %d, want %d", first.StatusCode, http.StatusCreated)
	}

	dup := doJSON(t, http.MethodPost, srv.URL+"/api/notepad-labels", CreateInput{Name: "ideas", Color: "red"})
	if dup.StatusCode != http.StatusBadRequest {
		t.Fatalf("duplicate POST status = %d, want %d", dup.StatusCode, http.StatusBadRequest)
	}
}

func TestHandler_UpdateUnknownLabelReturns404(t *testing.T) {
	srv := newTestServer(t)

	name := "x"
	res := doJSON(t, http.MethodPatch, srv.URL+"/api/notepad-labels/does-not-exist", UpdateInput{Name: &name})
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}

func TestHandler_DeleteUnknownLabelReturns404(t *testing.T) {
	srv := newTestServer(t)

	res := doJSON(t, http.MethodDelete, srv.URL+"/api/notepad-labels/does-not-exist", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}
