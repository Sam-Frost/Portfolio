package drawingboard

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// captureLogs redirects slog.Default at debug level into a buffer for the
// duration of the test. drawingboard's service emits a structured line for
// each mutating operation via reqlog.FromContext, which falls back to
// slog.Default() outside an HTTP request.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return buf
}

func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("parse log line %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

func hasLine(lines []map[string]any, msg string) map[string]any {
	for _, l := range lines {
		if l["msg"] == msg {
			return l
		}
	}
	return nil
}

func TestService_LogsBoardLifecycle(t *testing.T) {
	buf := captureLogs(t)
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	b, err := svc.Create(ctx, CreateInput{Name: ptr("Roadmap")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	created := hasLine(logLines(t, buf), "drawing board created")
	if created == nil {
		t.Fatalf("no 'drawing board created' line in:\n%s", buf.String())
	}
	if created["board_id"] != b.ID || created["name"] != "Roadmap" {
		t.Errorf("created line = %v, want board_id=%s name=Roadmap", created, b.ID)
	}

	name := "Roadmap v2"
	if _, err := svc.Update(ctx, b.ID, UpdateInput{Name: &name}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if hasLine(logLines(t, buf), "drawing board renamed") == nil {
		t.Errorf("no 'drawing board renamed' line in:\n%s", buf.String())
	}

	scene := json.RawMessage(`{"elements":[],"appState":{},"files":{}}`)
	if _, err := svc.Update(ctx, b.ID, UpdateInput{SceneData: scene}); err != nil {
		t.Fatalf("Update scene: %v", err)
	}
	if hasLine(logLines(t, buf), "drawing board scene saved") == nil {
		t.Errorf("no 'drawing board scene saved' line in:\n%s", buf.String())
	}

	if err := svc.Delete(ctx, b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if hasLine(logLines(t, buf), "drawing board deleted") == nil {
		t.Errorf("no 'drawing board deleted' line in:\n%s", buf.String())
	}
}

func TestService_NoLogWhenCreateInputRejected(t *testing.T) {
	buf := captureLogs(t)
	svc := NewService(NewMemoryRepository())

	// A rejected update must not emit a success line.
	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	buf.Reset()

	_, err = svc.Update(context.Background(), b.ID, UpdateInput{SceneData: json.RawMessage(`"not-an-object"`)})
	if err == nil {
		t.Fatal("Update with bad scene = nil error, want rejection")
	}
	if strings.Contains(buf.String(), "scene saved") {
		t.Errorf("rejected update still logged a success line:\n%s", buf.String())
	}
}
