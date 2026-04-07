// ─────────────────────────────────────────────────────────
//  统一 Mock 数据文件
//  结构与后端 response DTO 完全一致
//  后端联调好后：删除此文件，各页面改为真实 API 调用即可
// ─────────────────────────────────────────────────────────
import type {
  DashboardMarketSnapshotResponse,
  TransactionListResponse,
  TransactionStats,
  PortfolioResponse,
  AnalysisReportResponse,
  UserResponse,
  UploadHistoryResponse,
} from './api/types';

// ── Dashboard 市场快照 ──────────────────────────────────
export const mockDashboardSnapshot: DashboardMarketSnapshotResponse = {
  snapshot_time: '2024-03-27 15:00',
  is_stale: false,
  source: 'mock',
  indices: [
    { symbol: '000001.SH', name: '上证指数', last_price: '3128.42', change_amount: '12.36', change_percent: '0.40' },
    { symbol: '399001.SZ', name: '深证成指', last_price: '9856.74', change_amount: '-23.11', change_percent: '-0.23' },
    { symbol: 'HSI',       name: '恒生指数', last_price: '16852.30', change_amount: '105.20', change_percent: '0.63' },
  ],
  main_chart: {
    index_name: '上证指数',
    series: [
      { label: '03-17', value: '3058' },
      { label: '03-18', value: '3072' },
      { label: '03-19', value: '3064' },
      { label: '03-20', value: '3096' },
      { label: '03-21', value: '3108' },
      { label: '03-24', value: '3116' },
      { label: '03-25', value: '3128' },
    ],
  },
  stats: [
    { label: '区间涨跌幅', value: '+6.82%' },
    { label: '跑赢基准',   value: '+1.36%' },
    { label: '最大回撤',   value: '4.90%'  },
    { label: '年化波动率', value: '18.4%'  },
    { label: '月度换手率', value: '32%'    },
    { label: '近30日胜率', value: '61%'    },
  ],
};

// ── 用户信息 ────────────────────────────────────────────
export const mockUser: UserResponse = {
  id: 1,
  username: 'admin',
  email: 'admin@guanshi.ai',
  investment_preference: 'aggressive',
  total_profit: '18400.00',
  risk_tolerance: 'high',
};

// ── 持仓列表 ────────────────────────────────────────────
export const mockPortfolios: PortfolioResponse[] = [
  { id: 1, asset_code: '0700.HK', asset_name: '腾讯控股', asset_type: 'stock', total_quantity: '200', available_quantity: '200', average_cost: '280.00', current_price: '310.50', market_value: '62100.00', profit_loss: '6100.00', profit_loss_percent: '10.89', last_updated: '2024-03-27' },
  { id: 2, asset_code: '600519',  asset_name: '贵州茅台', asset_type: 'stock', total_quantity: '10',  available_quantity: '10',  average_cost: '1650.00', current_price: '1720.00', market_value: '17200.00', profit_loss: '700.00',  profit_loss_percent: '4.24',  last_updated: '2024-03-27' },
  { id: 3, asset_code: '513100',  asset_name: '纳指100ETF', asset_type: 'fund', total_quantity: '5000', available_quantity: '5000', average_cost: '1.18', current_price: '1.25', market_value: '6250.00', profit_loss: '350.00', profit_loss_percent: '5.93', last_updated: '2024-03-27' },
  { id: 4, asset_code: 'NVDA.US', asset_name: '英伟达', asset_type: 'stock', total_quantity: '5', available_quantity: '5', average_cost: '650.00', current_price: '890.20', market_value: '4451.00', profit_loss: '1201.00', profit_loss_percent: '36.95', last_updated: '2024-03-27' },
];

// ── 交易列表 ────────────────────────────────────────────
export const mockTransactionList: TransactionListResponse = {
  total: 5,
  page: 1,
  page_size: 10,
  transactions: [
    { id: 1, transaction_date: '2024-03-22', transaction_type: 'buy',  asset_type: 'stock', asset_code: '0700.HK', asset_name: '腾讯控股',    quantity: '100',  price_per_unit: '290.50', total_amount: '29050.00', commission: '29.05', profit: null,     notes: null, created_at: '2024-03-22T14:30:00Z' },
    { id: 2, transaction_date: '2024-03-21', transaction_type: 'sell', asset_type: 'stock', asset_code: '600519',  asset_name: '贵州茅台',    quantity: '10',   price_per_unit: '1720.00', total_amount: '17200.00', commission: '17.20', profit: '-380.00', notes: null, created_at: '2024-03-21T10:15:00Z' },
    { id: 3, transaction_date: '2024-03-20', transaction_type: 'buy',  asset_type: 'fund',  asset_code: '513100',  asset_name: '纳指100ETF', quantity: '5000', price_per_unit: '1.25',    total_amount: '6250.00',  commission: '6.25',  profit: null,     notes: null, created_at: '2024-03-20T09:45:00Z' },
    { id: 4, transaction_date: '2024-03-19', transaction_type: 'buy',  asset_type: 'stock', asset_code: 'NVDA.US', asset_name: '英伟达',      quantity: '5',    price_per_unit: '890.20',  total_amount: '4451.00',  commission: '4.45',  profit: null,     notes: null, created_at: '2024-03-19T15:00:00Z' },
    { id: 5, transaction_date: '2024-03-18', transaction_type: 'sell', asset_type: 'stock', asset_code: '600036',  asset_name: '招商银行',    quantity: '1000', price_per_unit: '32.10',   total_amount: '32100.00', commission: '32.10', profit: '-120.00', notes: null, created_at: '2024-03-18T11:20:00Z' },
  ],
};

// ── 交易统计 ────────────────────────────────────────────
export const mockTransactionStats: TransactionStats = {
  total_transactions: 26,
  buy_count: 16,
  sell_count: 10,
  total_investment: '89051.00',
  total_profit: '3700.00',
};

// ── AI 分析报告 ─────────────────────────────────────────
export const mockAnalysisReport: AnalysisReportResponse = {
  id: 1,
  report_type: 'comprehensive',
  report_title: 'AI 深度风险诊断报告',
  analysis_period_start: '2024-01-01',
  analysis_period_end: '2024-03-27',
  total_investment: '89051.00',
  total_profit: '3700.00',
  profit_rate: '4.15',
  risk_level: 'high',
  investment_style: 'aggressive',
  summary_text: '当前组合仍处于偏进攻状态，收益动能尚可，但需尽快压低集中度与高频换手。',
  risk_analysis: '持仓集中度过高。您的前两大持仓占总资产 65%，极易受单一行业波动影响。',
  pattern_insights: '检测到轻微的"处置效应"（倾向于过早卖出盈利股，而长期持有亏损股）。',
  prediction_text: '基于当前持仓结构，预计未来 3 个月收益区间为 -5.1% ~ +34.2%。',
  chart_data: JSON.stringify({
    radar: [82, 45, 30, 65, 90],
    labels: ['收益爆发力', '回撤控制', '资产分散度', '交易纪律', '风格稳定性'],
  }),
  recommendations: '建议将科技板块仓位下调 15%，增配防御性资产如红利低波 ETF。',
  ai_model: 'DeepSeek-V3',
  created_at: '2024-03-27T15:00:00Z',
};

// ── 上传历史 ────────────────────────────────────────────
export const mockUploadHistory: UploadHistoryResponse[] = [
  { id: 1, file_name: '招商银行对账单_202403.csv', file_size: 12400, file_type: 'csv',  upload_status: 'success', records_imported: 18, uploaded_at: '2024-03-22T10:00:00Z', processed_at: '2024-03-22T10:01:00Z' },
  { id: 2, file_name: '华泰证券导出_202403.xlsx',  file_size: 28600, file_type: 'xlsx', upload_status: 'success', records_imported: 32, uploaded_at: '2024-03-15T09:30:00Z', processed_at: '2024-03-15T09:31:00Z' },
];