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

import { searchAction } from '@/pages/search/action';
import { searchLoader } from '@/pages/search/loader';

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
        element: <ScreenshotDetailPage />
      },
      {
        path: 'search',
        element: <SearchResultsPage />,
        action: searchAction,
        loader: searchLoader,
      },
    ]
  }
]);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
