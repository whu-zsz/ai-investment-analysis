import { Navigate } from 'react-router-dom';
import type { ReactElement } from 'react';

export const ProtectedRoute = ({ children }: { children: ReactElement }) => {
  const token = localStorage.getItem('token');
  // 如果没有 token，直接弹回登录页
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return children;
};
