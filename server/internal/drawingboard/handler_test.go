package drawingboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer wires a real Handler (backed by a fresh MemoryRepository)
// onto its own mux and starts it, so these tests exercise the full
// HTTP -> handler -> service -> repository stack end to end, the same path
// a real client hits.
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

func TestHandler_FullBoardLifecycle(t *testing.T) {
	srv := newTestServer(t)

	// Create.
	createRes := doJSON(t, http.MethodPost, srv.URL+"/api/drawing-boards", CreateInput{Name: ptr("Roadmap")})
	if createRes.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d", createRes.StatusCode, http.StatusCreated)
	}
	created := decode[Board](t, createRes)
	if created.Name != "Roadmap" {
		t.Errorf("created.Name = %q, want %q", created.Name, "Roadmap")
	}
	if created.ID == "" {
		t.Fatal("created.ID is empty")
	}

	// List includes it.
	listRes := doJSON(t, http.MethodGet, srv.URL+"/api/drawing-boards", nil)
	summaries := decode[[]BoardSummary](t, listRes)
	if len(summaries) != 1 || summaries[0].ID != created.ID {
		t.Fatalf("List = %+v, want just %q", summaries, created.ID)
	}

	// Get round-trips the full board.
	getRes := doJSON(t, http.MethodGet, srv.URL+"/api/drawing-boards/"+created.ID, nil)
	got := decode[Board](t, getRes)
	if got.ID != created.ID || got.Name != created.Name {
		t.Fatalf("Get = %+v, want id/name matching create", got)
	}

	// Update renames and pushes a new scene.
	scene := json.RawMessage(`{"elements":[{"id":"r1","type":"rectangle"}],"appState":{},"files":{}}`)
	updateRes := doJSON(t, http.MethodPatch, srv.URL+"/api/drawing-boards/"+created.ID,
		UpdateInput{Name: ptr("Roadmap v2"), SceneData: scene})
	if updateRes.StatusCode != http.StatusOK {
		t.Fatalf("PATCH status = %d, want %d", updateRes.StatusCode, http.StatusOK)
	}
	updated := decode[Board](t, updateRes)
	if updated.Name != "Roadmap v2" {
		t.Errorf("updated.Name = %q, want %q", updated.Name, "Roadmap v2")
	}
	if string(updated.SceneData) != string(scene) {
		t.Errorf("updated.SceneData = %s, want %s", updated.SceneData, scene)
	}

	// Delete removes it.
	delRes := doJSON(t, http.MethodDelete, srv.URL+"/api/drawing-boards/"+created.ID, nil)
	if delRes.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", delRes.StatusCode, http.StatusNoContent)
	}

	// Now gone.
	afterDeleteRes := doJSON(t, http.MethodGet, srv.URL+"/api/drawing-boards/"+created.ID, nil)
	if afterDeleteRes.StatusCode != http.StatusNotFound {
		t.Fatalf("GET after delete status = %d, want %d", afterDeleteRes.StatusCode, http.StatusNotFound)
	}
}

func TestHandler_CreateWithNoBodyDefaultsName(t *testing.T) {
	srv := newTestServer(t)

	res := doJSON(t, http.MethodPost, srv.URL+"/api/drawing-boards", CreateInput{})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusCreated)
	}
	b := decode[Board](t, res)
	if b.Name == "" {
		t.Error("Name is empty, want a default timestamp-based name")
	}
}

func TestHandler_UpdateMalformedSceneDataReturns400(t *testing.T) {
	srv := newTestServer(t)

	created := decode[Board](t, doJSON(t, http.MethodPost, srv.URL+"/api/drawing-boards", CreateInput{}))

	// The request body is well-formed JSON, but sceneData's value is a
	// string rather than the expected scene object — the service-level
	// shape check should reject it, not just DecodeJSON's outer parse.
	res := doJSON(t, http.MethodPatch, srv.URL+"/api/drawing-boards/"+created.ID,
		map[string]any{"sceneData": "not-a-scene-object"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
}

func TestHandler_GetUnknownBoardReturns404(t *testing.T) {
	srv := newTestServer(t)

	res := doJSON(t, http.MethodGet, srv.URL+"/api/drawing-boards/does-not-exist", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}
