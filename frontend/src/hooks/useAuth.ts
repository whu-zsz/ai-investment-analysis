import { useState, useCallback } from 'react';
import type { UserResponse } from '../api/types';

export type UserInfo = UserResponse;

function normalizeStoredUser(raw: unknown): UserInfo | null {
  if (!raw || typeof raw !== 'object') {
    return null;
  }

  const candidate = raw as Record<string, unknown>;
  const username = typeof candidate.username === 'string' ? candidate.username : '';
  const email = typeof candidate.email === 'string' ? candidate.email : '';

  if (!username || !email) {
    return null;
  }

  const investmentPreference = candidate.investment_preference;
  const legacyRole = candidate.role;
  const normalizedPreference = investmentPreference === 'conservative' || investmentPreference === 'balanced' || investmentPreference === 'aggressive'
    ? investmentPreference
    : legacyRole === '保守型'
      ? 'conservative'
      : legacyRole === '激进型'
        ? 'aggressive'
        : 'balanced';

  return {
    id: typeof candidate.id === 'number' ? candidate.id : 0,
    username,
    email,
    phone: typeof candidate.phone === 'string' ? candidate.phone : undefined,
    avatar_url: typeof candidate.avatar_url === 'string'
      ? candidate.avatar_url
      : typeof candidate.avatar === 'string'
        ? candidate.avatar
        : undefined,
    investment_preference: normalizedPreference,
    total_profit: typeof candidate.total_profit === 'string' ? candidate.total_profit : '0.00',
    risk_tolerance: typeof candidate.risk_tolerance === 'string' ? candidate.risk_tolerance : 'unknown',
  };
}

const getInitialState = () => {
  const token = localStorage.getItem('token');
  if (!token) {
    return { isLoggedIn: false, userInfo: null as UserInfo | null, isLoading: false };
  }

  try {
    const saved = localStorage.getItem('userInfo');
    return {
      isLoggedIn: true,
      userInfo: normalizeStoredUser(saved ? JSON.parse(saved) : null),
      isLoading: false,
    };
  } catch {
    localStorage.removeItem('userInfo');
    return { isLoggedIn: true, userInfo: null as UserInfo | null, isLoading: false };
  }
};

export function useAuth() {
  const [{ isLoggedIn, userInfo }, setState] = useState(getInitialState);

  const login = useCallback((user: UserResponse, token: string) => {
    localStorage.setItem('token', token);
    localStorage.setItem('userInfo', JSON.stringify(user));
    setState({ isLoggedIn: true, userInfo: user, isLoading: false });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('userInfo');
    setState({ isLoggedIn: false, userInfo: null, isLoading: false });
  }, []);

  const updateUserInfo = useCallback((info: Partial<UserInfo>) => {
    if (!userInfo) {
      return;
    }

    const updated = { ...userInfo, ...info };
    localStorage.setItem('userInfo', JSON.stringify(updated));
    setState(prev => ({ isLoggedIn: prev.isLoggedIn, userInfo: updated, isLoading: false }));
  }, [userInfo]);

  return { isLoggedIn, userInfo, login, logout, updateUserInfo, isLoading: false };
}
