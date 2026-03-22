import type { StatItem } from '../types';

export const summaryStats: StatItem[] = [
  { label: '组合总资产', value: '¥ 428,560', hint: '较上周 +3.8%', tone: 'positive' },
  { label: '近 30 日盈亏', value: '+¥ 16,840', hint: '成长板块贡献最大', tone: 'positive' },
  { label: '风险等级', value: '中偏稳健', hint: '回撤控制优于上月', tone: 'warning' },
  { label: 'AI 分析完成度', value: '89%', hint: '可导出 PDF 报告', tone: 'neutral' },
];

export const holdings = [
  { name: '新能源 ETF', code: '516160', ratio: 28, profit: '+12.8%' },
  { name: '中证红利', code: '515080', ratio: 22, profit: '+4.9%' },
  { name: '贵州茅台', code: '600519', ratio: 18, profit: '+6.3%' },
  { name: 'QQQ', code: 'NASDAQ', ratio: 14, profit: '+9.1%' },
  { name: '现金与货基', code: 'CASH', ratio: 18, profit: '+1.8%' },
];

export const forecastBars = [
  { label: '10月', actual: 38, forecast: 40 },
  { label: '11月', actual: 46, forecast: 48 },
  { label: '12月', actual: 44, forecast: 49 },
  { label: '1月', actual: 54, forecast: 58 },
  { label: '2月', actual: 63, forecast: 66 },
  { label: '3月', actual: 67, forecast: 71 },
];

export const alerts = [
  '新能源相关资产已占总仓位 38%，单一赛道波动风险偏高。',
  '亏损后的快速卖出较多，存在情绪化止损倾向。',
  '若提升红利与现金配置，组合波动率预计可下降约 11%-15%。',
];

export const uploadPreview = [
  { field: '交易日期', status: '已识别', note: '标准化为 YYYY-MM-DD' },
  { field: '证券代码', status: '已识别', note: '兼容股票、ETF、基金代码' },
  { field: '成交价格', status: '已识别', note: '自动清洗货币符号与千分位' },
  { field: '交易方向', status: '待确认', note: '建议统一映射为 buy / sell / dividend' },
];

export const reportCards = [
  {
    title: '投资偏好画像',
    text: '组合明显偏向高景气成长赛道，但近期开始补充红利和现金类资产，说明你的风格正在向平衡型过渡。',
  },
  {
    title: '行为模式识别',
    text: '高收益交易多数来自中长期持有，而短线频繁调仓胜率不足 43%，适合减少高频择时。',
  },
  {
    title: '风险预警',
    text: '主要风险来自主题集中与止损纪律不稳定，市场波动放大时，新能源和海外科技仓位将最敏感。',
  },
];

export const scenarioCards = [
  { title: '稳健情景', range: '+3% ~ +6%', detail: '维持 20% 现金仓位，优先控制回撤。' },
  { title: '中性情景', range: '+7% ~ +11%', detail: '成长与红利双线配置，收益与波动相对均衡。' },
  { title: '进取情景', range: '+12% ~ +18%', detail: '提升赛道仓位，但需承担节奏失配风险。' },
];

export const historyRows = [
  ['2026-03-19', '买入', '516160', '新能源 ETF', '320', '3.28', '¥1,049.60', '已入库'],
  ['2026-03-14', '卖出', '600519', '贵州茅台', '8', '1682.00', '¥13,456.00', '已分析'],
  ['2026-03-08', '买入', '515080', '中证红利', '500', '1.42', '¥710.00', '已分析'],
  ['2026-02-27', '分红', 'CASH', '货币基金', '-', '-', '¥182.30', '已归档'],
];
