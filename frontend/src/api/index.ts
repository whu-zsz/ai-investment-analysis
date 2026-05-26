import request, { API_BASE_URL } from './request';
import type {
  LoginRequest, RegisterRequest, LoginResponse,
  UserResponse, UpdateProfileRequest,
  TransactionListResponse, TransactionStats, CreateTransactionRequest, UpdateTransactionRequest, TransactionResponse,
  PortfolioResponse,
  DashboardMarketSnapshotResponse, MarketSnapshotResponse,
  MarketStockDetailResponse, MarketStockKlineResponse,
  DashboardMarketBreadthResponse, StockProfileResponse, StockNewsResponse, BoardNewsResponse, MarketBoardDetailResponse,
  AnalysisReportResponse, AnalysisTaskResponse, AnalysisTaskDetailResponse, AnalysisTaskListResponse, AnalysisReportDetailResponse,
  UploadResponse, UploadHistoryResponse,
  AnalysisCandidatesResponse, AnalysisRecommendationsResponse,
  StockChatResponse, BoardChatResponse, RecommendationChatResponse, ChatContextSnapshotResponse, StockChatMessageRequest, StockChatStreamEvent,
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

  /** GET /dashboard/market-breadth —— 全市场宽度、排行与板块 */
  getDashboardMarketBreadth: (params?: { limit?: number }): Promise<DashboardMarketBreadthResponse> =>
    request.get('/dashboard/market-breadth', { params }),

  /** GET /market/snapshots/latest */
  getLatestSnapshots: (): Promise<MarketSnapshotResponse[]> =>
    request.get('/market/snapshots/latest'),

  /** GET /market/snapshots/history */
  getSnapshotHistory: (params?: {
    symbol?: string;
    limit?: number;
  }): Promise<MarketSnapshotResponse[]> =>
    request.get('/market/snapshots/history', { params }),

  /** GET /market/stocks/search */
  searchStocks: (params: { q: string; limit?: number }): Promise<MarketSnapshotResponse[]> =>
    request.get('/market/stocks/search', { params }),

  /** GET /market/stocks/:symbol/detail */
  getStockDetail: (symbol: string, params?: { refresh?: boolean }): Promise<MarketStockDetailResponse> =>
    request.get(`/market/stocks/${encodeURIComponent(symbol)}/detail`, { params }),

  /** GET /market/stocks/:symbol/kline */
  getStockKlines: (
    symbol: string,
    params?: { period?: string; adjust?: string; limit?: number; refresh?: boolean },
  ): Promise<MarketStockKlineResponse> =>
    request.get(`/market/stocks/${encodeURIComponent(symbol)}/kline`, { params }),

  /** GET /market/stocks/:symbol/profile */
  getStockProfile: (symbol: string): Promise<StockProfileResponse> =>
    request.get(`/market/stocks/${encodeURIComponent(symbol)}/profile`),

  /** GET /market/stocks/:symbol/news */
  getStockNews: (symbol: string, params?: { limit?: number }): Promise<StockNewsResponse> =>
    request.get(`/market/stocks/${encodeURIComponent(symbol)}/news`, { params }),

  /** GET /market/boards/:boardType/:code */
  getBoardDetail: (boardType: string, code: string, params?: { limit?: number }): Promise<MarketBoardDetailResponse> =>
    request.get(`/market/boards/${encodeURIComponent(boardType)}/${encodeURIComponent(code)}`, { params }),

  /** GET /market/boards/:boardType/:code/news */
  getBoardNews: (boardType: string, code: string, params?: { limit?: number }): Promise<BoardNewsResponse> =>
    request.get(`/market/boards/${encodeURIComponent(boardType)}/${encodeURIComponent(code)}/news`, { params }),
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

  /** GET /analysis/reports/:id/pdf */
  downloadReportPDF: (id: number): Promise<Blob> =>
    request.get(`/analysis/reports/${id}/pdf`, { responseType: 'blob' }),

  /** POST /analysis/stock-chat */
  stockChat: (data: {
    symbol: string;
    question: string;
    messages?: StockChatMessageRequest[];
    context_id?: number;
    refresh_market?: boolean;
  }): Promise<StockChatResponse> =>
    request.post('/analysis/stock-chat', data),

  /** POST /analysis/stock-chat/stream */
  stockChatStream: async (
    data: {
      symbol: string;
      question: string;
      messages?: StockChatMessageRequest[];
      context_id?: number;
      refresh_market?: boolean;
    },
    onEvent: (event: StockChatStreamEvent) => void,
  ): Promise<void> => {
    const token = localStorage.getItem('token');
    const resp = await fetch(`${API_BASE_URL}/api/v1/analysis/stock-chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(data),
    });

    if (resp.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      window.location.href = '/login';
      return;
    }
    if (!resp.ok || !resp.body) {
      const text = await resp.text();
      throw new Error(text || `请求失败：${resp.status}`);
    }

    const reader = resp.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    const parseChunk = (chunk: string) => {
        const dataLine = chunk.split('\n').find((line) => line.startsWith('data:'));
        if (!dataLine) return;
        const payload = dataLine.slice(5).trim();
        if (!payload) return;
        onEvent(JSON.parse(payload) as StockChatStreamEvent);
    };

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const chunks = buffer.replace(/\r\n/g, '\n').split('\n\n');
      buffer = chunks.pop() ?? '';
      for (const chunk of chunks) {
        parseChunk(chunk);
      }
    }
    buffer += decoder.decode();
    const tail = buffer.replace(/\r\n/g, '\n').trim();
    if (tail) parseChunk(tail);
  },

  /** POST /analysis/board-chat */
  boardChat: (data: {
    board_type: string;
    code: string;
    question: string;
    messages?: StockChatMessageRequest[];
    context_id?: number;
  }): Promise<BoardChatResponse> =>
    request.post('/analysis/board-chat', data),

  /** POST /analysis/board-chat/stream */
  boardChatStream: async (
    data: {
      board_type: string;
      code: string;
      question: string;
      messages?: StockChatMessageRequest[];
      context_id?: number;
    },
    onEvent: (event: StockChatStreamEvent) => void,
  ): Promise<void> => {
    const token = localStorage.getItem('token');
    const resp = await fetch(`${API_BASE_URL}/api/v1/analysis/board-chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(data),
    });

    if (resp.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      window.location.href = '/login';
      return;
    }
    if (!resp.ok || !resp.body) {
      const text = await resp.text();
      throw new Error(text || `请求失败：${resp.status}`);
    }

    const reader = resp.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    const parseChunk = (chunk: string) => {
      const dataLine = chunk.split('\n').find((line) => line.startsWith('data:'));
      if (!dataLine) return;
      const payload = dataLine.slice(5).trim();
      if (!payload) return;
      onEvent(JSON.parse(payload) as StockChatStreamEvent);
    };

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const chunks = buffer.replace(/\r\n/g, '\n').split('\n\n');
      buffer = chunks.pop() ?? '';
      for (const chunk of chunks) {
        parseChunk(chunk);
      }
    }
    buffer += decoder.decode();
    const tail = buffer.replace(/\r\n/g, '\n').trim();
    if (tail) parseChunk(tail);
  },

  /** GET /analysis/candidates */
  getCandidates: (): Promise<AnalysisCandidatesResponse> =>
    request.get('/analysis/candidates'),

  /** GET /analysis/recommendations */
  getRecommendations: (): Promise<AnalysisRecommendationsResponse> =>
    request.get('/analysis/recommendations'),

  /** GET /analysis/chat-contexts/:id */
  getChatContext: (id: number): Promise<ChatContextSnapshotResponse> =>
    request.get(`/analysis/chat-contexts/${id}`),

  /** GET /analysis/recommendation-chat/contexts/:id */
  getRecommendationChatContext: (id: number): Promise<import('./types').RecommendationChatContextSnapshotResponse> =>
    request.get(`/analysis/recommendation-chat/contexts/${id}`),

  /** GET /analysis/recommendation-chat/reports/:id/context */
  getRecommendationChatContextByReportId: (id: number): Promise<ChatContextSnapshotResponse> =>
    request.get(`/analysis/recommendation-chat/reports/${id}/context`),

  /** POST /analysis/recommendation-chat */
  recommendationChat: (data: {
    question: string;
    messages?: StockChatMessageRequest[];
    report_id?: number;
    context_id?: number;
  }): Promise<RecommendationChatResponse> =>
    request.post('/analysis/recommendation-chat', data),

  /** POST /analysis/recommendation-chat/stream */
  recommendationChatStream: async (
    data: {
      question: string;
      messages?: StockChatMessageRequest[];
      report_id?: number;
      context_id?: number;
    },
    onEvent: (event: StockChatStreamEvent) => void,
  ): Promise<void> => {
    const token = localStorage.getItem('token');
    const resp = await fetch(`${API_BASE_URL}/api/v1/analysis/recommendation-chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(data),
    });

    if (resp.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      window.location.href = '/login';
      return;
    }
    if (!resp.ok || !resp.body) {
      const text = await resp.text();
      throw new Error(text || `请求失败：${resp.status}`);
    }

    const reader = resp.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    const parseChunk = (chunk: string) => {
      const dataLine = chunk.split('\n').find((line) => line.startsWith('data:'));
      if (!dataLine) return;
      const payload = dataLine.slice(5).trim();
      if (!payload) return;
      onEvent(JSON.parse(payload) as StockChatStreamEvent);
    };

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const chunks = buffer.replace(/\r\n/g, '\n').split('\n\n');
      buffer = chunks.pop() ?? '';
      for (const chunk of chunks) {
        parseChunk(chunk);
      }
    }
    buffer += decoder.decode();
    const tail = buffer.replace(/\r\n/g, '\n').trim();
    if (tail) parseChunk(tail);
  },
};
