import { createBrowserRouter } from "react-router-dom";

import { DomainExpansionPage } from "./pages/DomainExpansionPage";
import { DomainErrorPage } from "./pages/DomainErrorPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { DomainLayout } from "./layouts/DomainLayout";
import { RequireAuth } from "./features/auth/RequireAuth";
import { CredentialManagerPage } from "./features/credentials/CredentialManagerPage";
import { TodosPage } from "./features/todos/TodosPage";
import { SettingsPage } from "./features/settings/SettingsPage";
import { NotepadPage } from "./features/notepad/NotepadPage";
import { NoteEditorPage } from "./features/notepad/NoteEditorPage";
import { ScratchNotePage } from "./features/notepad/ScratchNotePage";
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

export const router = createBrowserRouter([
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
            path: "upskill",
            element: <UpskillTopicsPage />,
            handle: { title: "Upskill" },
          },
          {
            path: "upskill/:topicId",
            element: <UpskillTopicPage />,
            handle: { title: "Upskill" },
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
]);
