import axios from 'axios';
import { message } from 'antd';

const request = axios.create({
  baseURL: 'http://localhost:8080/api/v1', // 后端 Gin 接口地址
  timeout: 10000,
});

// 请求拦截器：自动注入 Token
request.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// 响应拦截器：统一处理错误提示
request.interceptors.response.use(
  response => response.data,
  error => {
    message.error(error.response?.data?.message || '网络请求失败');
    return Promise.reject(error);
  }
);

export default request;

/** 后端接口预留 **/
export const api = {
  login: (data: any) => request.post('/auth/login', data),
  getDashboard: () => request.get('/analysis/dashboard'),
  getAnalysis: () => request.get('/analysis/report'), // AI 总结与风险
  getPrediction: () => request.get('/analysis/prediction'), // 趋势预测
  getHistory: () => request.get('/history'),
  uploadFile: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return request.post('/files/upload', formData);
  }
};