import { useState, useEffect } from 'react';

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

export function useAuth() {
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      setIsLoggedIn(true);
      const saved = localStorage.getItem('userInfo');
      setUserInfo(saved ? JSON.parse(saved) : DEFAULT_USER);
    }
  }, []);

  const login = (username: string) => {
    const info: UserInfo = { ...DEFAULT_USER, username, displayName: username };
    localStorage.setItem('token', 'auth_token_' + Date.now());
    localStorage.setItem('userInfo', JSON.stringify(info));
    setIsLoggedIn(true);
    setUserInfo(info);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('userInfo');
    setIsLoggedIn(false);
    setUserInfo(null);
  };

  const updateUserInfo = (info: Partial<UserInfo>) => {
    const updated = { ...(userInfo ?? DEFAULT_USER), ...info };
    localStorage.setItem('userInfo', JSON.stringify(updated));
    setUserInfo(updated);
  };

  return { isLoggedIn, userInfo, login, logout, updateUserInfo };
}