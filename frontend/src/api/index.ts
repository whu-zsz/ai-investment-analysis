import request from './request';
import type {
  LoginRequest, RegisterRequest, LoginResponse,
  UserResponse, UpdateProfileRequest,
  TransactionListResponse, TransactionStats, CreateTransactionRequest, UpdateTransactionRequest, TransactionResponse,
  PortfolioResponse,
  DashboardMarketSnapshotResponse, MarketSnapshotResponse,
  AnalysisReportResponse, AnalysisTaskResponse, AnalysisTaskDetailResponse, AnalysisTaskListResponse, AnalysisReportDetailResponse,
  UploadResponse, UploadHistoryResponse,
} from './types';

// ══════════════════════════════════════════
//  AUTH  /api/v1/auth
// ══════════════════════════════════════════

export const authApi = {
  /** POST /auth/register */
  register: (data: RegisterRequest): Promise<UserResponse> =>
    request.post('/auth/register', data),

  /** POST /auth/login */
  login: (data: LoginRequest): Promise<LoginResponse> =>
    request.post('/auth/login', data),

  /** POST /auth/logout */
  logout: (): Promise<void> =>
    request.post('/auth/logout'),
};

// ══════════════════════════════════════════
//  USER  /api/v1/user
// ══════════════════════════════════════════

export const userApi = {
  /** GET /user/profile */
  getProfile: (): Promise<UserResponse> =>
    request.get('/user/profile'),

  /** PUT /user/profile */
  updateProfile: (data: UpdateProfileRequest): Promise<UserResponse> =>
    request.put('/user/profile', data),
};

// ══════════════════════════════════════════
//  UPLOAD  /api/v1/upload
// ══════════════════════════════════════════

export const uploadApi = {
  /**
   * POST /upload
   * 上传 CSV / Excel 文件，需用 FormData 发送
   */
  uploadFile: (file: File): Promise<UploadResponse> => {
    const form = new FormData();
    form.append('file', file);
    return request.post('/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  /** GET /upload/history */
  getHistory: (): Promise<UploadHistoryResponse[]> =>
    request.get('/upload/history'),
};

// ══════════════════════════════════════════
//  TRANSACTIONS  /api/v1/transactions
// ══════════════════════════════════════════

export const transactionApi = {
  /**
   * GET /transactions
   * 当前仅支持分页参数
   */
  getList: (params?: {
    page?: number;
    page_size?: number;
  }): Promise<TransactionListResponse> =>
    request.get('/transactions', { params }),

  /** GET /transactions/stats */
  getStats: (): Promise<TransactionStats> =>
    request.get('/transactions/stats'),

  /** GET /transactions/:id */
  getDetail: (id: number): Promise<TransactionResponse> =>
    request.get(`/transactions/${id}`),

  /** POST /transactions */
  create: (data: CreateTransactionRequest): Promise<void> =>
    request.post('/transactions', data),

  /** PUT /transactions/:id */
  update: (id: number, data: UpdateTransactionRequest): Promise<TransactionResponse> =>
    request.put(`/transactions/${id}`, data),

  /** DELETE /transactions/:id */
  delete: (id: number): Promise<void> =>
    request.delete(`/transactions/${id}`),
};

// ══════════════════════════════════════════
//  PORTFOLIOS  /api/v1/portfolios
// ══════════════════════════════════════════

export const portfolioApi = {
  /** GET /portfolios */
  getList: (): Promise<PortfolioResponse[]> =>
    request.get('/portfolios'),
};

// ══════════════════════════════════════════
//  MARKET  /api/v1/market  +  /api/v1/dashboard
// ══════════════════════════════════════════

export const marketApi = {
  /** GET /dashboard/market-snapshot —— Dashboard 专用聚合数据 */
  getDashboardSnapshot: (): Promise<DashboardMarketSnapshotResponse> =>
    request.get('/dashboard/market-snapshot'),

  /** GET /market/snapshots/latest */
  getLatestSnapshots: (): Promise<MarketSnapshotResponse[]> =>
    request.get('/market/snapshots/latest'),

  /** GET /market/snapshots/history */
  getSnapshotHistory: (params?: {
    symbol?: string;
    limit?: number;
  }): Promise<MarketSnapshotResponse[]> =>
    request.get('/market/snapshots/history', { params }),
};

// ══════════════════════════════════════════
//  ANALYSIS  /api/v1/analysis
// ══════════════════════════════════════════

export const analysisApi = {
  /**
   * POST /analysis/tasks
   * 创建异步分析任务
   */
  createTask: (data: {
    start_date: string;
    end_date: string;
    symbols?: string[];
    force_refresh_market?: boolean;
    force_refresh_metrics?: boolean;
  }): Promise<AnalysisTaskResponse> =>
    request.post('/analysis/tasks', data),

  /** GET /analysis/tasks */
  getTasks: (params?: {
    status?: string;
    page?: number;
    page_size?: number;
  }): Promise<AnalysisTaskListResponse> =>
    request.get('/analysis/tasks', { params }),

  /** GET /analysis/tasks/:id */
  getTask: (id: number): Promise<AnalysisTaskDetailResponse> =>
    request.get(`/analysis/tasks/${id}`),

  /**
   * POST /analysis/summary
   * 触发 AI 生成分析报告，可能耗时较长
   */
  generateSummary: (params: { start_date: string; end_date: string }): Promise<AnalysisReportResponse> =>
    request.post('/analysis/summary', null, { params }),

  /** GET /analysis/reports —— 获取历史报告列表 */
  getReports: (params?: { report_type?: string; limit?: number }): Promise<AnalysisReportResponse[]> =>
    request.get('/analysis/reports', { params }),

  /** GET /analysis/reports/:id */
  getReportDetail: (id: number): Promise<AnalysisReportDetailResponse> =>
    request.get(`/analysis/reports/${id}`),
};
