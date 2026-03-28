import { createBrowserRouter, Navigate } from 'react-router-dom';
import Dashboard from '../pages/Dashboard';
import UploadPage from '../pages/Upload';
import Analysis from '../pages/Analysis';
import Prediction from '../pages/Prediction';
import History from '../pages/History';
import Login from '../pages/Login';

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
    path: '/app/upload',
    element: <UploadPage />,
  },
  {
    path: '/app/analysis',
    element: <Analysis />,
  },
  {
    path: '/app/prediction',
    element: <Prediction />,
  },
  {
    path: '/app/history',
    element: <History />,
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);