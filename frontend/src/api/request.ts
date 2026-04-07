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
  (res) => res.data,
  (err) => {
    const status = err.response?.status;
    if (status === 401) {
      // token 过期或无效，清除登录态跳回登录页
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      window.location.href = '/login';
    }
    // 把后端返回的 message 透传给调用方
    return Promise.reject(err.response?.data ?? err);
  }
);

export default request;