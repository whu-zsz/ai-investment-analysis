import axios from 'axios';
import { clearAuthSession } from '../utils/authSession';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.PROD ? '' : 'http://localhost:8080');

const request = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

// ── 请求拦截器：自动附加 token ──
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// ── 响应拦截器：统一错误处理 ──
request.interceptors.response.use(
  (res) => {
    const payload = res.data;
    if (payload && typeof payload === 'object' && 'data' in payload) {
      return payload.data;
    }
    return payload;
  },
  (err) => {
    const status = err.response?.status;
    const url = err.config?.url ?? '';

    if (status === 401 && !url.includes('/auth/login')) {
      clearAuthSession();
      window.location.href = '/login';
    }

    return Promise.reject({
      status,
      message: err.response?.data?.message ?? err.message,
      data: err.response?.data,
    });
  }
);

export default request;
