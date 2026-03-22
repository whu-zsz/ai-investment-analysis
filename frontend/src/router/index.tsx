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
    path: '/login', 
    element: <Login /> 
  },
  {
    path: '/',
    element: <ProtectedRoute><MainLayout /></ProtectedRoute>, // 只有这里用 Layout
    children: [
      { index: true, element: <Dashboard /> }, // 首页直接渲染内容，不再套 Layout
      { path: 'upload', element: <UploadPage /> },
      { path: 'analysis', element: <Analysis /> },
      { path: 'prediction', element: <Prediction /> },
      { path: 'history', element: <History /> },
    ],
  },
]);