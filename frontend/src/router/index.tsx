import { createBrowserRouter } from 'react-router-dom';
import Layout from '../components/layout/Layout';
import Login from '../pages/Login';
import Dashboard from '../pages/Dashboard';
import Upload from '../pages/Upload';
import Analysis from '../pages/Analysis';
import Prediction from '../pages/Prediction';
import History from '../pages/History';
import ProtectedRoute from '../components/ProtectedRoute';   // ← 新增导入

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
]);

export default router;