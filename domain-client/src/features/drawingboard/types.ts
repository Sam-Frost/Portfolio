import type { ExcalidrawElement } from "@excalidraw/excalidraw/element/types";
import type { BinaryFiles } from "@excalidraw/excalidraw/types";

export interface DrawingBoardSummary {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

// A curated subset of Excalidraw's AppState — just enough to restore the
// viewport a board was left in. AppState also carries a lot of
// non-persistable, purely-runtime UI state (e.g. `collaborators` is a Map,
// which JSON.stringify silently turns into "{}"), so we save only the
// fields that matter for restoring the canvas rather than round-tripping
// the whole object.
export interface DrawingBoardSceneAppState {
  viewBackgroundColor: string;
  scrollX: number;
  scrollY: number;
  zoom: { value: number };
}

export interface DrawingBoardScene {
  elements: readonly ExcalidrawElement[];
  appState: DrawingBoardSceneAppState;
  files: BinaryFiles;
}

export interface DrawingBoard extends DrawingBoardSummary {
  sceneData: DrawingBoardScene;
}
