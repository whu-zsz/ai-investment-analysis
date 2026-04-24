import axios from 'axios';

// 后端地址，开发时改这里即可
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

const request = axios.create({
  baseURL: `${BASE_URL}/api/v1`,
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
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
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
