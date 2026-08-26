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
import { UpskillTopicsPage } from "./features/upskill/UpskillTopicsPage";
import { UpskillTopicPage } from "./features/upskill/UpskillTopicPage";
import { HourlyTrackerPage } from "./features/hourly-tracker/HourlyTrackerPage";

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
            handle: { title: "Hourly Tracker" },
          },
          { path: "settings", element: <SettingsPage /> },
          {
            path: "notepad",
            element: <NotepadPage />,
            handle: { title: "Notepad" },
          },
          {
            path: "notepad/:id",
            element: <NoteEditorPage />,
            handle: { title: "Notepad" },
          },
        ],
      },
    ],
  },
  // Unmatched URL — outside RequireAuth so a bad link doesn't force a login
  // redirect before showing the 404.
  { path: "*", element: <NotFoundPage /> },
]);
