import { Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';   // ← 这里改成 type 导入（关键！）

interface ProtectedRouteProps {
  children: ReactNode;
}

const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const token = localStorage.getItem('token');
  return token ? <>{children}</> : <Navigate to="/login" replace />;
};

export default ProtectedRoute;