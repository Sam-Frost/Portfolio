import { afterEach, beforeEach, describe, expect, mock, test } from "bun:test";
import {
  createDrawingBoard,
  deleteDrawingBoard,
  fetchDrawingBoard,
  fetchDrawingBoards,
  updateDrawingBoard,
} from "./api";
import { setToken, clearToken } from "../auth/token";
import type { DrawingBoardScene } from "./types";

// bun test runs outside a DOM, but token.ts (imported transitively via
// apiClient.ts) reads/writes localStorage — stub a minimal in-memory
// implementation so setToken/clearToken below don't throw.
if (typeof globalThis.localStorage === "undefined") {
  const store = new Map<string, string>();
  globalThis.localStorage = {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => void store.set(key, value),
    removeItem: (key: string) => void store.delete(key),
    clear: () => store.clear(),
  } as Storage;
}
if (typeof globalThis.sessionStorage === "undefined") {
  globalThis.sessionStorage = globalThis.localStorage;
}

const API_BASE_URL = "http://localhost:8080";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(status === 204 ? null : JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("drawingboard api", () => {
  let fetchMock: ReturnType<typeof mock>;

  beforeEach(() => {
    setToken("test-token");
    fetchMock = mock(() => Promise.resolve(jsonResponse({})));
    // @ts-expect-error test override of the global fetch
    globalThis.fetch = fetchMock;
  });

  afterEach(() => {
    clearToken();
  });

  test("fetchDrawingBoards calls GET /api/drawing-boards and returns the parsed list", async () => {
    const summaries = [{ id: "1", name: "Board 1", createdAt: "t", updatedAt: "t" }];
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse(summaries)));

    const result = await fetchDrawingBoards();

    expect(result).toEqual(summaries);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${API_BASE_URL}/api/drawing-boards`);
    expect(init.method ?? "GET").toBe("GET");
    expect((init.headers as Record<string, string>).Authorization).toBe("Bearer test-token");
  });

  test("fetchDrawingBoard calls GET /api/drawing-boards/:id", async () => {
    const board = { id: "1", name: "Board 1", createdAt: "t", updatedAt: "t", sceneData: {} };
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse(board)));

    const result = await fetchDrawingBoard("1");

    expect(result).toEqual(board);
    const [url] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${API_BASE_URL}/api/drawing-boards/1`);
  });

  test("createDrawingBoard POSTs a name (or null when omitted)", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse({ id: "new" }, 201)));

    await createDrawingBoard("My Board");
    let [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${API_BASE_URL}/api/drawing-boards`);
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ name: "My Board" });

    await createDrawingBoard();
    [url, init] = fetchMock.mock.calls[1] as [string, RequestInit];
    expect(JSON.parse(init.body as string)).toEqual({ name: null });
  });

  test("updateDrawingBoard PATCHes only the provided fields", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse({ id: "1" })));
    const scene: DrawingBoardScene = {
      elements: [],
      appState: { viewBackgroundColor: "#fff", scrollX: 0, scrollY: 0, zoom: { value: 1 } },
      files: {},
    };

    await updateDrawingBoard("1", { sceneData: scene });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${API_BASE_URL}/api/drawing-boards/1`);
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toEqual({ sceneData: scene });
  });

  test("deleteDrawingBoard DELETEs and resolves with no value on 204", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse(null, 204)));

    const result = await deleteDrawingBoard("1");

    expect(result).toBeUndefined();
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${API_BASE_URL}/api/drawing-boards/1`);
    expect(init.method).toBe("DELETE");
  });

  test("a non-ok response rejects with the server's error message", async () => {
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse({ error: "board not found" }, 404)));

    await expect(fetchDrawingBoard("missing")).rejects.toThrow("board not found");
  });
});
