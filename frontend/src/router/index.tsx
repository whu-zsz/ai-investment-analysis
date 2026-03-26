import { createBrowserRouter } from 'react-router-dom';
import MainLayout from '../components/layout/Layout';
import Dashboard from '../pages/Dashboard';
import UploadPage from '../pages/Upload';
import Analysis from '../pages/Analysis';
import Prediction from '../pages/Prediction';
import History from '../pages/History';
import Login from '../pages/Login';
import { ProtectedRoute } from '../components/ProtectedRoute';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Dashboard />,
  },
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/app',
    element: <ProtectedRoute><MainLayout /></ProtectedRoute>,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'upload', element: <UploadPage /> },
      { path: 'analysis', element: <Analysis /> },
      { path: 'prediction', element: <Prediction /> },
      { path: 'history', element: <History /> },
    ],
  },
]);