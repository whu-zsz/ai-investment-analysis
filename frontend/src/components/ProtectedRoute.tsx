import { Navigate, useLocation } from 'react-router-dom';
import type { ReactElement } from 'react';

export const ProtectedRoute = ({ children }: { children: ReactElement }) => {
  const location = useLocation();
  const token = localStorage.getItem('token');

  if (!token) {
    // 把想去的页面存起来，登录后自动跳回
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }
  return children;
};