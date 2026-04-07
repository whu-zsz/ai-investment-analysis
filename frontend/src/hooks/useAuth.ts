import { useState, useCallback } from 'react';

export interface UserInfo {
  username: string;
  displayName: string;
  email: string;
  avatar?: string;
  role: string;
  joinDate: string;
}

const DEFAULT_USER: UserInfo = {
  username: 'admin',
  displayName: '投资顾问',
  email: 'admin@guanshi.ai',
  role: '高级分析师',
  joinDate: '2024-01-01',
};

const getInitialState = () => {
  const token = localStorage.getItem('token');
  if (!token) {
    return { isLoggedIn: false, userInfo: null, isLoading: false };
  }
  const saved = localStorage.getItem('userInfo');
  return {
    isLoggedIn: true,
    userInfo: saved ? JSON.parse(saved) : DEFAULT_USER,
    isLoading: false,
  };
};

export function useAuth() {
  const [{ isLoggedIn, userInfo }, setState] = useState(getInitialState);

  const login = useCallback((username: string, email: string, role: string, token: string) => {
    const info: UserInfo = {
      username,
      displayName: username,
      email,
      role,
      joinDate: new Date().toISOString().slice(0, 10),
    };
    localStorage.setItem('token', token);
    localStorage.setItem('userInfo', JSON.stringify(info));
    setState({ isLoggedIn: true, userInfo: info, isLoading: false });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('userInfo');
    setState({ isLoggedIn: false, userInfo: null, isLoading: false });
  }, []);

  const updateUserInfo = useCallback((info: Partial<UserInfo>) => {
    const updated = { ...(userInfo ?? DEFAULT_USER), ...info };
    localStorage.setItem('userInfo', JSON.stringify(updated));
    setState(prev => ({ isLoggedIn: prev.isLoggedIn, userInfo: updated, isLoading: false }));
  }, [userInfo]);

  return { isLoggedIn, userInfo, login, logout, updateUserInfo, isLoading: false };
}
