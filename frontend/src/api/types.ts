// ─────────────────────────────────────────
//  对应 dto/response/auth.go
// ─────────────────────────────────────────
export interface UserResponse {
  id: number;
  username: string;
  email: string;
  phone?: string;
  avatar_url?: string;
  investment_preference: 'conservative' | 'balanced' | 'aggressive';
  total_profit: string;
  risk_tolerance: string;
}

export interface LoginResponse {
  token: string;
  user: UserResponse;
}

// ─────────────────────────────────────────
//  对应 dto/request/auth.go
// ─────────────────────────────────────────
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface UpdateProfileRequest {
  phone?: string;
  avatar_url?: string;
  investment_preference?: 'conservative' | 'balanced' | 'aggressive';
}

// ─────────────────────────────────────────
//  对应 dto/response/transaction.go
// ─────────────────────────────────────────
export interface TransactionResponse {
  id: number;
  transaction_date: string;
  transaction_type: 'buy' | 'sell' | 'dividend';
  asset_type: string;
  asset_code: string;
  asset_name: string;
  quantity: string;
  price_per_unit: string;
  total_amount: string;
  commission: string;
  profit?: string;
  notes?: string;
  created_at: string;
}

export interface TransactionListResponse {
  transactions: TransactionResponse[];
  total: number;
  page: number;
  page_size: number;
}

export interface TransactionStats {
  total_transactions: number;
  buy_count: number;
  sell_count: number;
  total_investment: string;
  total_profit: string;
}

// 对应 dto/request/transaction.go
export interface CreateTransactionRequest {
  transaction_date: string;
  transaction_type: 'buy' | 'sell' | 'dividend';
  asset_type: string;
  asset_code: string;
  asset_name: string;
  quantity: string;
  price_per_unit: string;
  commission?: string;
  notes?: string;
}

export interface UpdateTransactionRequest {
  transaction_date: string;
  transaction_type: 'buy' | 'sell' | 'dividend';
  asset_type: string;
  asset_code: string;
  asset_name: string;
  quantity: string;
  price_per_unit: string;
  commission?: string;
  notes?: string;
}

export type MarketDataStatus = 'complete' | 'fetched_live' | 'partial' | 'unavailable';
export type AnalysisReportType = 'summary' | 'risk' | 'prediction' | 'pattern';

export interface AnalysisTaskResponse {
  id: number;
  status: string;
  progress_stage: string;
  created_at: string;
}

export interface AnalysisTaskDetailResponse {
  id: number;
  task_type: string;
  status: string;
  progress_stage: string;
  analysis_period_start: string;
  analysis_period_end: string;
  result_report_id?: number;
  error_message: string;
  started_at: string;
  finished_at: string;
  created_at: string;
  updated_at: string;
}

export interface AnalysisTaskListResponse {
  items: AnalysisTaskDetailResponse[];
  total: number;
  page: number;
  page_size: number;
}

export interface AnalysisReportItemResponse {
  id: number;
  symbol: string;
  asset_name: string;
  market: string;
  trade_count: number;
  buy_count: number;
  sell_count: number;
  buy_amount: string;
  sell_amount: string;
  net_quantity: string;
  realized_profit: string;
  realized_profit_rate: string;
  ending_position_qty: string;
  ending_avg_cost: string;
  latest_price: string;
  latest_market_value: string;
  unrealized_profit: string;
  total_profit: string;
  change_percent_7d: string;
  period_price_change_pct: string;
  market_data_status: MarketDataStatus;
  risk_level: string;
  investment_style: string;
  analysis_text: string;
  recommendation: string;
  key_points: string[];
  created_at: string;
}

export interface RiskAnalysisResponse {
  risk_level: string;
  risk_score: number;
  risk_factors: string[];
  recommendations: string[];
}

export interface RiskAlertItemResponse {
  level: string;
  type: string;
  title: string;
  description: string;
  symbols: string[];
}

export interface RiskSymbolResponse {
  symbol: string;
  asset_name: string;
  risk_level: string;
  risk_score: number;
  trigger_reasons: string[];
}

export interface PredictionScenarioResponse {
  condition: string;
  outcome: string;
}

export interface PredictionResponse {
  bias: string;
  confidence: string;
  horizon: string;
  drivers: string[];
  scenarios: PredictionScenarioResponse[];
  narrative: string;
}

export interface AnalysisReportDetailResponse {
  id: number;
  task_id?: number;
  report_type: AnalysisReportType;
  report_title: string;
  analysis_period_start: string;
  analysis_period_end: string;
  symbols_count: number;
  winning_trades: number;
  losing_trades: number;
  total_investment: string;
  total_profit: string;
  profit_rate: string;
  risk_level: string;
  market_data_status: MarketDataStatus;
  investment_style: string;
  summary_text: string;
  risk_analysis: string;
  pattern_insights: string;
  prediction_text: string;
  chart_data: string;
  recommendations: string[];
  risk_overview: RiskAnalysisResponse;
  risk_alerts: RiskAlertItemResponse[];
  top_risk_symbols: RiskSymbolResponse[];
  prediction?: PredictionResponse;
  ai_model: string;
  created_at: string;
  items: AnalysisReportItemResponse[];
}


// ─────────────────────────────────────────
//  对应 dto/response/portfolio.go
// ─────────────────────────────────────────
export interface PortfolioResponse {
  id: number;
  asset_code: string;
  asset_name: string;
  asset_type: string;
  total_quantity: string;
  available_quantity: string;
  average_cost: string;
  current_price: string;
  market_value: string;
  profit_loss: string;
  profit_loss_percent: string;
  last_updated: string;
}

// ─────────────────────────────────────────
//  对应 dto/response/market.go
// ─────────────────────────────────────────
export interface MarketIndexItemResponse {
  symbol: string;
  name: string;
  market: string;
  last_price: string;
  change_amount: string;
  change_percent: string;
  open_price: string;
  high_price: string;
  low_price: string;
  prev_close: string;
  volume: string;
  turnover: string;
}

export interface MarketChartPoint {
  label: string;
  value: string;
}

export interface MarketChartResponse {
  index_name: string;
  series: MarketChartPoint[];
}

export interface DashboardStatResponse {
  label: string;
  value: string;
}

export interface DashboardMarketSnapshotResponse {
  snapshot_time: string;
  is_stale: boolean;
  source: string;
  indices: MarketIndexItemResponse[];
  main_chart: MarketChartResponse;
  stats: DashboardStatResponse[];
}

export interface MarketSnapshotResponse {
  symbol: string;
  name: string;
  market: string;
  snapshot_time: string;
  last_price: string;
  change_amount: string;
  change_percent: string;
  open_price: string;
  high_price: string;
  low_price: string;
  prev_close: string;
  volume: string;
  turnover: string;
  source: string;
  batch_no: string;
}

export interface MarketStockDetailResponse {
  symbol: string;
  name: string;
  market: string;
  last_price: string;
  open_price: string;
  high_price: string;
  low_price: string;
  prev_close: string;
  change_amount: string;
  change_percent: string;
  volume: string;
  turnover: string;
  volume_ratio: string;
  turnover_rate: string;
  amplitude: string;
  limit_up: string;
  limit_down: string;
  average_price: string;
  total_shares: string;
  float_shares: string;
  total_market_cap: string;
  float_market_cap: string;
  industry: string;
  region: string;
  concepts: string[];
  source: string;
  fetched_at: string;
  is_stale: boolean;
  refresh_triggered: boolean;
}

export interface MarketKlineBarResponse {
  bar_time: string;
  open_price: string;
  close_price: string;
  high_price: string;
  low_price: string;
  volume: string;
  turnover: string;
  amplitude: string;
  change_percent: string;
  change_amount: string;
  turnover_rate: string;
}

export interface MarketStockKlineResponse {
  symbol: string;
  period: string;
  adjust_type: string;
  source: string;
  fetched_at: string;
  is_stale: boolean;
  refresh_triggered: boolean;
  items: MarketKlineBarResponse[];
}

export interface AnalysisCandidateSource {
  type: 'portfolio' | 'transactions' | string;
}

export interface AnalysisCandidateResponse {
  symbol: string;
  asset_name: string;
  asset_type: string;
  market: string;
  sources: AnalysisCandidateSource[];
  is_held: boolean;
  trade_count: number;
  last_price: string;
  change_percent: string;
}

export interface AnalysisCandidatesResponse {
  default_symbol: string;
  candidates: AnalysisCandidateResponse[];
}

export interface RecommendationProfileSummary {
  investment_preference: 'conservative' | 'balanced' | 'aggressive' | string;
  risk_tolerance: string;
  total_profit: string;
  held_positions: number;
  candidate_count: number;
}

export interface RecommendationItemResponse {
  symbol: string;
  asset_name: string;
  asset_type: string;
  market: string;
  action: 'buy' | 'hold' | 'reduce' | 'sell' | 'observe' | string;
  score: string;
  latest_price: string;
  change_percent: string;
  match_reason: string;
  risk_note: string;
  data_status: MarketDataStatus | string;
  is_held: boolean;
  trade_count: number;
}

export interface AnalysisRecommendationsResponse {
  generated_at: string;
  profile_summary: RecommendationProfileSummary;
  summary_text: string;
  candidates: RecommendationItemResponse[];
}

// ─────────────────────────────────────────
//  对应 dto/response/analysis.go
// ─────────────────────────────────────────
export interface AnalysisReportResponse {
  id: number;
  report_type: AnalysisReportType;
  report_title: string;
  analysis_period_start: string;
  analysis_period_end: string;
  total_investment: string;
  total_profit: string;
  profit_rate: string;
  risk_level: string;
  market_data_status: MarketDataStatus;
  investment_style: string;
  summary_text: string;
  risk_analysis: string;
  pattern_insights: string;
  prediction_text: string;
  chart_data: string;
  recommendations: string[];
  ai_model: string;
  created_at: string;
}

// ─────────────────────────────────────────
//  对应 dto/response/upload.go
// ─────────────────────────────────────────
export interface UploadRowError {
  row_number: number;
  reason: string;
}

export interface UploadResponse {
  file_id: number;
  file_name: string;
  upload_status: 'success' | 'partial_success' | 'failed' | string;
  records_total: number;
  records_imported: number;
  records_failed: number;
  errors: UploadRowError[] | null;
  message: string;
}

export interface UploadHistoryResponse {
  id: number;
  file_name: string;
  file_size: number;
  file_type: string;
  upload_status: string;
  records_imported: number;
  uploaded_at: string;
  processed_at: string;
}
