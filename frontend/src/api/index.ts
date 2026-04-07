import request from './request';
import type {
  LoginRequest, RegisterRequest, LoginResponse,
  UserResponse, UpdateProfileRequest,
  TransactionListResponse, TransactionStats, CreateTransactionRequest,
  PortfolioResponse,
  DashboardMarketSnapshotResponse, MarketSnapshotResponse,
  AnalysisReportResponse,
  UploadResponse, UploadHistoryResponse,
} from './types';

// ══════════════════════════════════════════
//  AUTH  /api/v1/auth
// ══════════════════════════════════════════

export const authApi = {
  /** POST /auth/register */
  register: (data: RegisterRequest): Promise<LoginResponse> =>
    request.post('/auth/register', data),

  /** POST /auth/login */
  login: (data: LoginRequest): Promise<LoginResponse> =>
    request.post('/auth/login', data),
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
   * 支持分页和筛选参数
   */
  getList: (params?: {
    page?: number;
    page_size?: number;
    start_date?: string;
    end_date?: string;
    asset_code?: string;
    transaction_type?: string;
  }): Promise<TransactionListResponse> =>
    request.get('/transactions', { params }),

  /** GET /transactions/stats */
  getStats: (): Promise<TransactionStats> =>
    request.get('/transactions/stats'),

  /** POST /transactions */
  create: (data: CreateTransactionRequest): Promise<void> =>
    request.post('/transactions', data),

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
   * POST /analysis/summary
   * 触发 AI 生成分析报告，可能耗时较长
   */
  generateSummary: (params: { start_date: string; end_date: string }): Promise<AnalysisReportResponse> =>
    request.post('/analysis/summary', null, { params }),

  /** GET /analysis/reports —— 获取历史报告列表 */
  getReports: (): Promise<AnalysisReportResponse[]> =>
    request.get('/analysis/reports'),
};