import { createBrowserRouter, createMemoryRouter, type RouteObject } from "react-router-dom";

import { DomainExpansionPage } from "./pages/DomainExpansionPage";
import { DomainErrorPage } from "./pages/DomainErrorPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { DomainLayout } from "./layouts/DomainLayout";
import { RequireAuth } from "./features/auth/RequireAuth";
import { CredentialManagerPage } from "./features/credentials/CredentialManagerPage";
import { TodosPage } from "./features/todos/TodosPage";
import { WorkProfilePage } from "./features/work-profile/WorkProfilePage";
import { SettingsPage } from "./features/settings/SettingsPage";
import { NotepadPage } from "./features/notepad/NotepadPage";
import { NoteEditorPage } from "./features/notepad/NoteEditorPage";
import { ScratchNotePage } from "./features/notepad/ScratchNotePage";
import { DrawingBoardListPage } from "./features/drawingboard/DrawingBoardListPage";
import { DrawingBoardEditorPage } from "./features/drawingboard/DrawingBoardEditorPage";
import { UpskillTopicsPage } from "./features/upskill/UpskillTopicsPage";
import { UpskillTopicPage } from "./features/upskill/UpskillTopicPage";
import { DiaryCalendarPage } from "./features/diary/DiaryCalendarPage";
import { DiaryEntryPage } from "./features/diary/DiaryEntryPage";
import { HourlyTrackerPage } from "./features/hourly-tracker/HourlyTrackerPage";
import { FitnessCyclesPage } from "./features/fitness/FitnessCyclesPage";
import { FitnessCyclePage } from "./features/fitness/FitnessCyclePage";
import { ExerciseDetailPage } from "./features/fitness/ExerciseDetailPage";
import { DocumentsPage } from "./features/documents/DocumentsPage";
import { CmsPage } from "./features/cms/CmsPage";

const routes: RouteObject[] = [
  {
    path: "/",
    element: <DomainExpansionPage />,
    errorElement: <DomainErrorPage />,
  },
  {
    element: <RequireAuth />,
    errorElement: <DomainErrorPage />,
    children: [
      {
        element: <DomainLayout />,
        children: [
          { path: "credentials", element: <CredentialManagerPage /> },
          {
            path: "todos",
            element: <TodosPage />,
            handle: { title: "Todos", fullWidth: true },
          },
          {
            path: "work",
            element: <WorkProfilePage />,
            handle: { title: "Work Profile", fullWidth: true },
          },
          {
            path: "upskill",
            element: <UpskillTopicsPage />,
            handle: { title: "Upskill", fullWidth: true },
          },
          {
            path: "upskill/:topicId",
            element: <UpskillTopicPage />,
            handle: { title: "Upskill", fullWidth: true },
          },
          {
            path: "hourly-tracker",
            element: <HourlyTrackerPage />,
            handle: { title: "Sessions" },
          },
          { path: "settings", element: <SettingsPage /> },
          {
            path: "cms",
            element: <CmsPage />,
            handle: { title: "CMS", subtitle: "Public site content", fullWidth: true },
          },
          {
            path: "notepad",
            element: <NotepadPage />,
            handle: { title: "Notepad" },
          },
          {
            path: "notepad/scratch",
            element: <ScratchNotePage />,
            handle: { title: "Notepad" },
          },
          {
            path: "notepad/:id",
            element: <NoteEditorPage />,
            handle: { title: "Notepad" },
          },
          {
            path: "drawing-board",
            element: <DrawingBoardListPage />,
            handle: { title: "Drawing Board" },
          },
          {
            path: "drawing-board/:id",
            element: <DrawingBoardEditorPage />,
            handle: { title: "Drawing Board", fullWidth: true },
          },
          {
            path: "diary",
            element: <DiaryCalendarPage />,
            handle: { title: "Personal Diary" },
          },
          {
            path: "diary/:date",
            element: <DiaryEntryPage />,
            handle: { title: "Personal Diary" },
          },
          {
            path: "fitness",
            element: <FitnessCyclesPage />,
            handle: { title: "Fitness" },
          },
          {
            path: "fitness/:cycleId",
            element: <FitnessCyclePage />,
            handle: { title: "Fitness", fullWidth: true },
          },
          {
            path: "fitness/exercises/:exerciseId",
            element: <ExerciseDetailPage />,
            handle: { title: "Fitness" },
          },
          {
            path: "documents",
            element: <DocumentsPage />,
            handle: { title: "Documents", fullWidth: true },
          },
        ],
      },
    ],
  },
  // Unmatched URL — outside RequireAuth so a bad link doesn't force a login
  // redirect before showing the 404.
  { path: "*", element: <NotFoundPage /> },
];

// When the site is launched as an installed app — an iOS home-screen bookmark
// or any PWA in standalone display mode — route entirely in memory so the
// address the OS remembers for the app never changes (iOS treats a
// standalone bookmark whose URL stays put as its own "app"). In a normal
// browser tab we keep real URL routing, so shared deep links like
// /notepad/<uuid> and /diary/:date still work and are bookmarkable.
//
// The choice is made once, at load: the first entry is seeded from the real
// URL so a standalone launch still honours whatever path it opened at, then
// every in-app navigation (including the :id / :date routes) stays silent.
function isStandaloneLaunch(): boolean {
  if (typeof window === "undefined") return false;
  const iosStandalone =
    (window.navigator as Navigator & { standalone?: boolean }).standalone === true;
  const displayModeStandalone =
    window.matchMedia?.("(display-mode: standalone)").matches ?? false;
  return iosStandalone || displayModeStandalone;
}

export const router = isStandaloneLaunch()
  ? createMemoryRouter(routes, {
      initialEntries: [
        window.location.pathname + window.location.search + window.location.hash,
      ],
      initialIndex: 0,
    })
  : createBrowserRouter(routes);
