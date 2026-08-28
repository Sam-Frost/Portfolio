import { createRoot } from 'react-dom/client'
import './index.css'
import { router } from './router.tsx'
import { RouterProvider } from 'react-router-dom'
import { ContentProvider } from './content/ContentContext.tsx'


createRoot(document.getElementById('root')!).render(
  <ContentProvider>
    <RouterProvider router={router} />
  </ContentProvider>,
)
