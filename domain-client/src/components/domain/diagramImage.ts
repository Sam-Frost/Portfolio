import type { DrawingBoardScene } from "../../features/drawingboard/types";

// Marks an embedded <img> as an editable Excalidraw diagram (as opposed to
// a plain uploaded image) and carries the raw scene data needed to reopen
// it for editing — base64-encoded so it round-trips safely as an HTML
// attribute value inside contentHtml.
export const DIAGRAM_SCENE_ATTR = "data-excalidraw-scene";

const CHUNK_SIZE = 0x8000;

export function encodeScene(scene: DrawingBoardScene): string {
  const bytes = new TextEncoder().encode(JSON.stringify(scene));
  let binary = "";
  for (let i = 0; i < bytes.length; i += CHUNK_SIZE) {
    binary += String.fromCharCode(...bytes.subarray(i, i + CHUNK_SIZE));
  }
  return btoa(binary);
}

export function decodeScene(encoded: string): DrawingBoardScene | null {
  try {
    const binary = atob(encoded);
    const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0));
    return JSON.parse(new TextDecoder().decode(bytes));
  } catch {
    return null;
  }
}
