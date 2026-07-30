import { createBrowserRouter } from "react-router-dom";

import { RootLayout } from "./layouts/RootLayout";
import { HomePage } from "./pages/HomePage";
import { BadUrlPage } from "./pages/BadUrlPage";
import { ErrorPage } from "./pages/ErrorPage";
import { AllBlogsPage } from "./pages/AllBlogsPage";
import { BlogPage } from "./pages/BlogPage";
import { AllProjectsPage } from "./pages/AllProjectsPage";
import { ProjectPage } from "./pages/ProjectPage";
import { ExperiencePage } from "./pages/ExperiencePage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    errorElement: <ErrorPage />,
    children: [
      {
        index: true, element: <HomePage />
      },
      {
             path: "blogs",
             children: [
               {
                 index: true,
                 element: <AllBlogsPage />,
               },
               {
                 path: ":blogSlug",
                 element: <BlogPage />,
               },
             ],
           },

           {
             path: "projects",
             children: [
               {
                 index: true,
                 element: <AllProjectsPage />,
               },
               {
                 path: ":projectSlug",
                 element: <ProjectPage />,
               },
             ],
           },

           {
             path: "experience",
             children: [
               {
                 index: true,
                 element: <ExperiencePage />,
               },
             ],
           },
      { path: "*", element: <BadUrlPage /> },
    ],
  },

]);
