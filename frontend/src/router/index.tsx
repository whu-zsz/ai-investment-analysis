import { createBrowserRouter, Navigate } from 'react-router-dom';
import Dashboard from '../pages/Dashboard';
import UploadPage from '../pages/Upload';
import Analysis from '../pages/Analysis';
import Prediction from '../pages/Prediction';
import History from '../pages/History';
import Login from '../pages/Login';
import Profile from '../pages/Profile';
import PortfolioPage from '../pages/Portfolio';
import MarketTrendPage from '../pages/MarketTrend';
import BoardDetailPage from '../pages/BoardDetail';
import RecommendationPage from '../pages/Recommendation';
import StockChatPage from '../pages/StockChat';
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
    path: '/profile',
    element: <ProtectedRoute><Profile /></ProtectedRoute>,
  },
  {
    path: '/app/upload',
    element: <ProtectedRoute><UploadPage /></ProtectedRoute>,
  },
  {
    path: '/app/analysis',
    element: <ProtectedRoute><Analysis /></ProtectedRoute>,
  },
  {
    path: '/app/prediction',
    element: <ProtectedRoute><Prediction /></ProtectedRoute>,
  },
  {
    path: '/app/history',
    element: <ProtectedRoute><History /></ProtectedRoute>,
  },
  {
    path: '/app/portfolio',
    element: <ProtectedRoute><PortfolioPage /></ProtectedRoute>,
  },
  {
    path: '/app/market-trend',
    element: <ProtectedRoute><MarketTrendPage /></ProtectedRoute>,
  },
  {
    path: '/app/board',
    element: <ProtectedRoute><BoardDetailPage /></ProtectedRoute>,
  },
  {
    path: '/app/recommendation',
    element: <ProtectedRoute><RecommendationPage /></ProtectedRoute>,
  },
  {
    path: '/app/chat',
    element: <ProtectedRoute><StockChatPage /></ProtectedRoute>,
  },
  {
    path: '/app/stock-chat',
    element: <Navigate to="/app/chat" replace />,
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);
