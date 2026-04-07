import axios from 'axios';
import { message } from 'antd';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface BackendUserProfile {
  id: number;
  username: string;
  email: string;
  phone: string | null;
  avatar_url: string | null;
  investment_preference: string;
  total_profit: string;
  risk_tolerance: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginData {
  token: string;
  user: BackendUserProfile;
}

const request = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
});

request.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  response => response.data,
  error => {
    message.error(error.response?.data?.message || '网络请求失败');
    return Promise.reject(error);
  }
);

const apiRequest = request as unknown as {
  post<T>(url: string, data?: unknown, config?: unknown): Promise<T>;
  get<T>(url: string, config?: unknown): Promise<T>;
  put<T>(url: string, data?: unknown, config?: unknown): Promise<T>;
};

export default request;

export const api = {
  login: (data: LoginRequest) => apiRequest.post<ApiResponse<LoginData>>('/auth/login', data),
  getProfile: () => apiRequest.get<ApiResponse<BackendUserProfile>>('/user/profile'),
  updateProfile: (data: Partial<Pick<BackendUserProfile, 'phone' | 'avatar_url' | 'investment_preference'>>) =>
    apiRequest.put<ApiResponse<null>>('/user/profile', data),
  getTransactions: (params?: { page?: number; page_size?: number }) => apiRequest.get<any>('/transactions', { params }),
  getTransactionStats: () => apiRequest.get<any>('/transactions/stats'),
  getPortfolios: () => apiRequest.get<any>('/portfolios'),
  getReports: (params?: { report_type?: string; limit?: number }) => apiRequest.get<any>('/analysis/reports', { params }),
  generateSummary: (params: { start_date: string; end_date: string }) => apiRequest.post<any>('/analysis/summary', null, { params }),
  getUploadHistory: () => apiRequest.get<any>('/upload/history'),
  uploadFile: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiRequest.post<any>('/upload', formData);
  }
};
