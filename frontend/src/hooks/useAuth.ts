import { createContext, createElement, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { api, type BackendUserProfile, type LoginRequest } from '../types';

interface AuthContextValue {
  isLoggedIn: boolean;
  isLoading: boolean;
  userInfo: BackendUserProfile | null;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => void;
  refreshProfile: () => Promise<void>;
  updateUserInfo: (info: Partial<BackendUserProfile>) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const TOKEN_KEY = 'token';
const USER_KEY = 'userInfo';

function readStoredUser(): BackendUserProfile | null {
  const saved = localStorage.getItem(USER_KEY);
  if (!saved) return null;

  try {
    return JSON.parse(saved) as BackendUserProfile;
  } catch {
    localStorage.removeItem(USER_KEY);
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isLoading, setIsLoading] = useState(true);
  const [userInfo, setUserInfo] = useState<BackendUserProfile | null>(() => readStoredUser());
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));

  const clearAuth = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setToken(null);
    setUserInfo(null);
  }, []);

  const refreshProfile = useCallback(async () => {
    const currentToken = localStorage.getItem(TOKEN_KEY);
    if (!currentToken) {
      setUserInfo(null);
      return;
    }

    try {
      const profile = await api.getProfile();
      localStorage.setItem(USER_KEY, JSON.stringify(profile.data));
      setUserInfo(profile.data);
      setToken(currentToken);
    } catch {
      clearAuth();
      throw new Error('refresh profile failed');
    }
  }, [clearAuth]);

  useEffect(() => {
    const currentToken = localStorage.getItem(TOKEN_KEY);
    if (!currentToken) {
      setIsLoading(false);
      return;
    }

    refreshProfile()
      .catch(() => undefined)
      .finally(() => setIsLoading(false));
  }, [refreshProfile]);

  const login = useCallback(async (credentials: LoginRequest) => {
    const loginData = await api.login(credentials);

    localStorage.setItem(TOKEN_KEY, loginData.data.token);
    localStorage.setItem(USER_KEY, JSON.stringify(loginData.data.user));
    setToken(loginData.data.token);
    setUserInfo(loginData.data.user);
  }, []);

  const logout = useCallback(() => {
    clearAuth();
  }, [clearAuth]);

  const updateUserInfo = useCallback((info: Partial<BackendUserProfile>) => {
    setUserInfo(prev => {
      if (!prev) return prev;
      const updated = { ...prev, ...info };
      localStorage.setItem(USER_KEY, JSON.stringify(updated));
      return updated;
    });
  }, []);

  const value = useMemo<AuthContextValue>(() => ({
    isLoggedIn: Boolean(token),
    isLoading,
    userInfo,
    login,
    logout,
    refreshProfile,
    updateUserInfo,
  }), [isLoading, userInfo, login, logout, refreshProfile, token, updateUserInfo]);

  return createElement(AuthContext.Provider, { value }, children);
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
