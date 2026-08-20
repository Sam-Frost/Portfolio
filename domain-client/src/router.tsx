import { createBrowserRouter, Navigate } from "react-router-dom";

import { DomainExpansionPage } from "./pages/DomainExpansionPage";
import { DomainLayout } from "./layouts/DomainLayout";
import { CredentialManagerPage } from "./pages/CredentialManagerPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <DomainExpansionPage />,
  },
  {
    path: "/dashboard",
    element: <DomainLayout />,
    children: [
      { index: true, element: <Navigate to="credentials" replace /> },
      { path: "credentials", element: <CredentialManagerPage /> },
    ],
  },
]);
