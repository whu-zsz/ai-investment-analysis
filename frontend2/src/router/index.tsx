import { createBrowserRouter, Navigate } from 'react-router-dom';
import ProtectedRoute from '../components/ProtectedRoute';
import Layout from '../components/layout/Layout';
import Analysis from '../pages/Analysis';
import Dashboard from '../pages/Dashboard';
import History from '../pages/History';
import Login from '../pages/Login';
import Prediction from '../pages/Prediction';
import Upload from '../pages/Upload';

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <Layout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'upload', element: <Upload /> },
      { path: 'analysis', element: <Analysis /> },
      { path: 'prediction', element: <Prediction /> },
      { path: 'history', element: <History /> },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;
