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
export type AnalysisReportType = 'summary' | 'risk' | 'prediction' | 'pattern' | 'recommendation';

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
  refreshed_at: string;
  is_stale: boolean;
  source: string;
  indices: MarketIndexItemResponse[];
  main_chart: MarketChartResponse;
  stats: DashboardStatResponse[];
}

export interface MarketBreadthItemResponse {
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
  turnover_rate: string;
  total_market_cap: string;
  float_market_cap: string;
}

export interface MarketBoardItemResponse {
  board_type: string;
  code: string;
  name: string;
  last_price: string;
  change_amount: string;
  change_percent: string;
  volume: string;
  turnover: string;
  total_market_cap: string;
  float_market_cap: string;
  stock_count: number;
  rise_count: number;
  fall_count: number;
  flat_count: number;
  constituent_node: string;
}

export interface BoardConstituentResponse {
  symbol: string;
  name: string;
  market: string;
  last_price: string;
  change_amount: string;
  change_percent: string;
  volume: string;
  turnover: string;
  total_market_cap: string;
  float_market_cap: string;
  source: string;
  has_snapshot: boolean;
}

export interface MarketBoardDetailResponse {
  snapshot_time: string;
  refreshed_at: string;
  source: string;
  is_partial: boolean;
  message: string;
  board: MarketBoardItemResponse;
  coverage: DashboardStatResponse[];
  constituents: BoardConstituentResponse[];
  top_gainers: BoardConstituentResponse[];
  top_losers: BoardConstituentResponse[];
  top_turnover: BoardConstituentResponse[];
}

export interface DistributionBucketResponse {
  label: string;
  count: number;
  value: string;
}

export interface DashboardMarketBreadthResponse {
  snapshot_time: string;
  refreshed_at: string;
  source: string;
  is_partial: boolean;
  message: string;
  coverage: DashboardStatResponse[];
  top_gainers: MarketBreadthItemResponse[];
  top_losers: MarketBreadthItemResponse[];
  top_turnover: MarketBreadthItemResponse[];
  sectors: MarketBoardItemResponse[];
  concepts: MarketBoardItemResponse[];
  change_distribution: DistributionBucketResponse[];
  turnover_distribution: DistributionBucketResponse[];
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

export interface StockBoardMembershipResponse {
  board_type: string;
  code: string;
  name: string;
  source: string;
}

export interface StockCompanyProfileResponse {
  company_name: string;
  english_name: string;
  market_label: string;
  industry_label: string;
  legal_representative: string;
  registered_capital: string;
  founded_at: string;
  listed_at: string;
  website: string;
  email: string;
  phone: string;
  address: string;
  office_address: string;
  business: string;
  business_scope: string;
  introduction: string;
  source: string;
}

export interface StockProfileResponse {
  symbol: string;
  name: string;
  market: string;
  description: string;
  company_profile?: StockCompanyProfileResponse;
  industry: string;
  region: string;
  concepts: string[];
  boards: StockBoardMembershipResponse[];
  last_price: string;
  change_amount: string;
  change_percent: string;
  volume: string;
  turnover: string;
  volume_ratio: string;
  turnover_rate: string;
  amplitude: string;
  limit_up: string;
  limit_down: string;
  total_market_cap: string;
  float_market_cap: string;
  source: string;
  fetched_at: string;
  is_stale: boolean;
  refresh_triggered: boolean;
}

export interface StockNewsItemResponse {
  title: string;
  summary: string;
  source: string;
  url: string;
  published_at: string;
  provider: string;
  is_recent: boolean;
}

export interface StockNewsResponse {
  symbol: string;
  asset_name: string;
  generated_at: string;
  coverage: string;
  items: StockNewsItemResponse[];
}

export interface BoardNewsResponse {
  board_type: string;
  code: string;
  board_name: string;
  generated_at: string;
  coverage: string;
  items: StockNewsItemResponse[];
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
  source_tags: string[];
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
  report_id: number;
  profile_summary: RecommendationProfileSummary;
  summary_text: string;
  candidates: RecommendationItemResponse[];
}

export interface RecommendationChatContextResponse {
  step_key: string;
  step_label: string;
  profile_summary: RecommendationProfileSummary;
  candidate_count: number;
  discovery_count: number;
  held_count: number;
  focus_summary: string;
}

export interface StockChatMessageRequest {
  role: 'user' | 'assistant';
  content: string;
}

export interface StockChatMessageResponse {
  role: 'user' | 'assistant';
  content: string;
}

export interface StockChatNewsItemResponse {
  title: string;
  summary: string;
  source: string;
  url: string;
  published_at: string;
  provider: string;
}

export interface StockChatSnapshotResponse {
  last_price: string;
  change_percent: string;
  high_price: string;
  low_price: string;
  volume: string;
  turnover: string;
  source: string;
  fetched_at: string;
  period: string;
  trend_summary: string;
}

export interface ChatToolTraceStepResponse {
  stage: string;
  label: string;
  status: string;
  summary?: string;
  tool_name?: string;
  started_at?: string;
  finished_at?: string;
}

export interface ChatStepContextResponse {
  stage: string;
  label: string;
  summary?: string;
  tool_name?: string;
  news_items?: StockChatNewsItemResponse[];
  profile_summary?: RecommendationProfileSummary;
  candidate_count?: number;
  held_count?: number;
  discovery_count?: number;
  focus_summary?: string;
  reference_symbols?: string[];
  reference_boards?: string[];
}

export interface StockChatResponse {
  context_id: number;
  symbol: string;
  asset_name: string;
  market: string;
  reply: string;
  ai_model: string;
  generated_at: string;
  news_status: string;
  news_summary: string;
  news_coverage: string;
  news_items: StockChatNewsItemResponse[];
  snapshot: StockChatSnapshotResponse;
  messages: StockChatMessageResponse[];
  tool_trace?: ChatToolTraceStepResponse[];
}

export interface BoardChatResponse {
  context_id: number;
  board_type: string;
  code: string;
  asset_name: string;
  market: string;
  reply: string;
  ai_model: string;
  generated_at: string;
  news_status: string;
  news_summary: string;
  news_coverage: string;
  news_items: StockChatNewsItemResponse[];
  snapshot: StockChatSnapshotResponse;
  messages: StockChatMessageResponse[];
  tool_trace?: ChatToolTraceStepResponse[];
}

export interface RecommendationChatResponse {
  context_id: number;
  asset_name: string;
  market: string;
  reply: string;
  ai_model: string;
  generated_at: string;
  news_status: string;
  news_summary: string;
  news_coverage: string;
  news_items: StockChatNewsItemResponse[];
  snapshot: StockChatSnapshotResponse;
  messages: StockChatMessageResponse[];
  context: RecommendationChatContextResponse;
  report_id: number;
  report_title: string;
  candidates: RecommendationItemResponse[];
  tool_trace?: ChatToolTraceStepResponse[];
  step_context?: ChatStepContextResponse;
}

export interface RecommendationChatContextSnapshotResponse {
  messages: StockChatMessageResponse[];
  tool_trace?: ChatToolTraceStepResponse[];
  step_context?: ChatStepContextResponse;
  news_items?: StockChatNewsItemResponse[];
  reply: string;
  generated_at: string;
  report_id: number;
  report_title: string;
}

export interface ChatContextSnapshotResponse {
  context_id: number;
  context_type: string;
  target_key: string;
  title: string;
  report_id: number;
  messages: StockChatMessageResponse[];
  tool_trace?: ChatToolTraceStepResponse[];
  step_context?: ChatStepContextResponse;
  news_items?: StockChatNewsItemResponse[];
  reply: string;
  generated_at: string;
  report_title?: string;
}

export type StockChatStreamEventType = 'step' | 'context' | 'token' | 'done' | 'error' | 'heartbeat';

export interface StockChatStreamEvent {
  type: StockChatStreamEventType;
  stage?: string;
  message?: string;
  token?: string;
  data?: StockChatResponse | BoardChatResponse | RecommendationChatResponse | StockChatNewsItemResponse[] | ChatStepContextResponse;
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
