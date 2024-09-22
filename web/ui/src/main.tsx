import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import '@/index.css';

import App from '@/pages/App';
import ErrorPage from '@/pages/Error';

import DashboardPage from '@/pages/dashboard/Dashboard';
import GalleryPage from '@/pages/gallery/Gallery';
import TablePage from '@/pages/table/Table';
import ScreenshotDetailPage from '@/pages/detail/Detail';
import SearchResultsPage from '@/pages/search/Search';
import JobSubmissionPage from '@/pages/submit/Submit';

import { searchAction } from '@/pages/search/action';
import { searchLoader } from '@/pages/search/loader';
import { deleteAction } from '@/pages/detail/actions';
import { submitImmediateAction, submitJobAction } from '@/pages/submit/action';

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    errorElement: <ErrorPage />,
    children: [
      {
        path: '/',
        element: <DashboardPage />
      },
      {
        path: 'gallery',
        element: <GalleryPage />
      },
      {
        path: 'overview',
        element: <TablePage />
      },
      {
        path: 'screenshot/:id',
        element: <ScreenshotDetailPage />,
        action: deleteAction
      },
      {
        path: 'search',
        element: <SearchResultsPage />,
        action: searchAction,
        loader: searchLoader,
      },
      {
        path: 'submit',
        element: <JobSubmissionPage />,
        action: async ({ request }) => {
          const formData = await request.formData();
          const action = formData.get('action');

          switch (action) {
            case 'job':
              return submitJobAction({ formData });
            case 'immediate':
              return submitImmediateAction({ formData });

            default:
              throw new Error('unknown action for job submit route');
          }
        },
      },
    ]
  }
]);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
