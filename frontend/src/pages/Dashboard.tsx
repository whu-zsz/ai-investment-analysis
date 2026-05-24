import { Alert, Avatar, Button, Card, Col, Dropdown, Empty, Row, Segmented, Skeleton, Space, Spin, Tag, Typography } from 'antd';
import {
  BulbOutlined,
  LineChartOutlined,
  RadarChartOutlined,
  RiseOutlined,
  ThunderboltOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { marketApi } from '../api';
import { useAuth } from '../hooks/useAuth';
import { computeNumericAxisRange } from '../utils/chartAxis';
import type { DashboardMarketSnapshotResponse, MarketKlineBarResponse, MarketStockKlineResponse } from '../api/types';
import type { MenuProps } from 'antd';

const { Paragraph, Text, Title } = Typography;

type DashboardRange = '3d' | '7d' | '30d';
type MarketIndexSnapshot = DashboardMarketSnapshotResponse['indices'][number];

interface ChartParam {
  axisValueLabel?: string;
  dataIndex?: number;
  name: string;
  value: number;
}

interface DashboardRangeQuery {
  period: 'day' | '60m';
  adjust: 'none';
  limit: number;
}

interface KpiChartConfig {
  seriesName: string;
  unit: string;
  labels: string[];
  values: number[];
  color: string;
  period: 'day' | '60m';
  bars?: MarketKlineBarResponse[];
  isFallback?: boolean;
}

interface DashboardInsightCard {
  symbol: string;
  title: string;
  value: number;
  precision: number;
  suffix: string;
  accent: string;
  tagColor: string;
  tagText: string;
  desc: string;
  chart: KpiChartConfig;
  hasHistory: boolean;
  snapshot: MarketIndexSnapshot;
}

const chartPalette = ['#1677ff', '#52c41a', '#722ed1', '#fa8c16', '#13c2c2', '#eb2f96'];
const dashboardRangeOptions: Array<{ label: string; value: DashboardRange }> = [
  { label: '3日', value: '3d' },
  { label: '7日', value: '7d' },
  { label: '月', value: '30d' },
];
export const dashboardRangeQueryMap: Record<DashboardRange, DashboardRangeQuery> = {
  '3d': { period: '60m', adjust: 'none', limit: 12 },
  '7d': { period: '60m', adjust: 'none', limit: 28 },
  '30d': { period: 'day', adjust: 'none', limit: 30 },
};

const statColorMap: Record<string, string> = {
  指数数量: '#1677ff',
  上涨数: '#52c41a',
  下跌数: '#ff4d4f',
  平均涨跌幅: '#722ed1',
  总成交额: '#13c2c2',
};

function toNumber(value?: string): number {
  if (!value) return 0;
  const parsed = Number.parseFloat(value.replace(/[%亿万,+]/g, ''));
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatValue(value?: string) {
  const text = value?.trim();
  return text ? text : '—';
}

function formatSignedNumber(value?: string, precision = 2, suffix = ''): string {
  const numeric = toNumber(value);
  if (!value?.trim() || !Number.isFinite(numeric)) {
    return '—';
  }
  const sign = numeric > 0 ? '+' : numeric < 0 ? '-' : '';
  return `${sign}${Math.abs(numeric).toFixed(precision)}${suffix}`;
}

function formatSignedPercent(changePercent?: string): string {
  const numeric = toNumber(changePercent);
  if (!changePercent?.trim() || !Number.isFinite(numeric)) {
    return '—';
  }
  const sign = numeric > 0 ? '+' : numeric < 0 ? '-' : '';
  return `${sign}${Math.abs(numeric).toFixed(2)}%`;
}

function formatCompactNumber(value?: string): string {
  const numeric = toNumber(value);
  if (!value?.trim() || !Number.isFinite(numeric)) {
    return '—';
  }
  const abs = Math.abs(numeric);
  if (abs >= 100000000) {
    return `${(numeric / 100000000).toFixed(2)}亿`;
  }
  if (abs >= 10000) {
    return `${(numeric / 10000).toFixed(2)}万`;
  }
  if (abs >= 1000) {
    return numeric.toFixed(0);
  }
  return numeric.toFixed(2);
}

function getTrendTag(changePercent?: string) {
  if (!changePercent?.trim()) {
    return { color: 'default', text: '涨跌未知' };
  }

  const numeric = toNumber(changePercent);
  if (numeric > 0) {
    return { color: 'green', text: `上涨 ${formatSignedPercent(changePercent)}` };
  }
  if (numeric < 0) {
    return { color: 'red', text: `下跌 ${formatSignedPercent(changePercent).replace('-', '')}` };
  }
  return { color: 'default', text: '平盘' };
}

function getDashboardRangeLabel(range: DashboardRange): string {
  return dashboardRangeOptions.find((item) => item.value === range)?.label ?? '月';
}

function getDashboardHistoryCacheKey(symbol: string, range: DashboardRange): string {
  return `${symbol}:${range}`;
}

export function formatDashboardBarTime(value: string, period: 'day' | '60m'): string {
  if (period === '60m') {
    return value.length >= 16 ? value.slice(5, 16) : value;
  }
  if (value.length >= 10) {
    return value.slice(5, 10);
  }
  return value;
}

function buildDashboardRichTooltip(chart: KpiChartConfig, color: string, params: unknown) {
  const list = params as ChartParam[];
  const data = list[0];
  if (!data) {
    return '';
  }

  const dataIndex = data.dataIndex ?? 0;
  const bar = chart.bars?.[dataIndex];
  if (!bar || chart.isFallback) {
    return `<div style="padding: 6px 8px; min-width: 140px;">
              <div style="color: #888; margin-bottom: 6px;">${data.axisValueLabel ?? data.name}</div>
              <div style="font-weight: 700; color: ${color}; font-size: 16px;">${data.value.toLocaleString()} ${chart.unit}</div>
            </div>`;
  }

  const changeColor = toNumber(bar.change_amount) >= 0 ? '#ff4d4f' : '#52c41a';
  const turnoverRate = formatValue(bar.turnover_rate);
  return `<div style="padding: 8px 10px; min-width: 220px;">
            <div style="color: #888; margin-bottom: 8px;">${bar.bar_time}</div>
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom: 8px; gap: 12px;">
              <span style="font-weight: 700; color: ${color}; font-size: 16px;">收 ${formatValue(bar.close_price)}</span>
              <span style="font-weight: 600; color: ${changeColor};">${formatSignedNumber(bar.change_amount)} / ${formatSignedPercent(bar.change_percent)}</span>
            </div>
            <div style="display:grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px 12px; color:#444;">
              <span>开 ${formatValue(bar.open_price)}</span>
              <span>高 ${formatValue(bar.high_price)}</span>
              <span>低 ${formatValue(bar.low_price)}</span>
              <span>收 ${formatValue(bar.close_price)}</span>
              <span>量 ${formatCompactNumber(bar.volume)}</span>
              <span>额 ${formatCompactNumber(bar.turnover)}</span>
              <span>振幅 ${formatSignedPercent(bar.amplitude)}</span>
              <span>换手 ${turnoverRate === '—' ? '—' : `${turnoverRate}%`}</span>
            </div>
          </div>`;
}

function normalizeKlineResponse(history: MarketStockKlineResponse): MarketStockKlineResponse {
  return {
    ...history,
    items: [...history.items].sort((a, b) => a.bar_time.localeCompare(b.bar_time)),
  };
}

function buildHistoryChart(name: string, history: MarketStockKlineResponse, color: string): KpiChartConfig {
  return {
    seriesName: name,
    unit: '点',
    labels: history.items.map((point) => formatDashboardBarTime(point.bar_time, history.period === '60m' ? '60m' : 'day')),
    values: history.items.map((point) => toNumber(point.close_price)),
    color,
    period: history.period === '60m' ? '60m' : 'day',
    bars: history.items,
  };
}

function buildFallbackChart(marketData: DashboardMarketSnapshotResponse, color: string): KpiChartConfig {
  return {
    seriesName: marketData.main_chart.index_name || '指数走势',
    unit: '点',
    labels: marketData.main_chart.series.map((point) => point.label),
    values: marketData.main_chart.series.map((point) => toNumber(point.value)),
    color,
    period: 'day',
    isFallback: true,
  };
}

function buildInsightCards(
  marketData: DashboardMarketSnapshotResponse | null,
  indexHistories: Record<string, MarketStockKlineResponse | null>,
  chartRange: DashboardRange,
): DashboardInsightCard[] {
  if (!marketData?.indices?.length) {
    return [];
  }

  const rangeLabel = getDashboardRangeLabel(chartRange);

  return marketData.indices.slice(0, 4).map((item, index) => {
    const color = chartPalette[index % chartPalette.length];
    const trend = getTrendTag(item.change_percent);
    const history = indexHistories[getDashboardHistoryCacheKey(item.symbol, chartRange)] ?? null;
    const hasHistory = Boolean(history?.items?.length);

    return {
      symbol: item.symbol,
      title: item.name,
      value: toNumber(item.last_price),
      precision: 2,
      suffix: '点',
      accent: color,
      tagColor: trend.color,
      tagText: trend.text,
      desc: hasHistory
        ? `${item.symbol} · ${rangeLabel}走势 · 高低 ${formatValue(item.high_price)}/${formatValue(item.low_price)}`
        : `${item.symbol} · ${rangeLabel}概览 · 成交额 ${formatCompactNumber(item.turnover)}`,
      chart: hasHistory
        ? buildHistoryChart(item.name, history as MarketStockKlineResponse, color)
        : { seriesName: item.name, unit: '点', labels: [], values: [], color, period: 'day' },
      hasHistory,
      snapshot: item,
    };
  });
}

function getKpiChartOption(chart: KpiChartConfig, mode: 'mini' | 'expanded'): EChartsOption {
  const isMini = mode === 'mini';
  const axisRange = computeNumericAxisRange(chart.values, {
    minPaddingRatio: 0.08,
    maxPaddingRatio: 0.08,
    minPaddingAbs: chart.period === '60m' ? 5 : 10,
    roundMode: 'magnitude',
    splitNumber: 4,
  });

  return {
    animation: true,
    tooltip: isMini
      ? { show: false }
      : {
          trigger: 'axis',
          backgroundColor: 'rgba(255, 255, 255, 0.96)',
          borderColor: '#d9e6ff',
          borderWidth: 1,
          formatter: (params: unknown) => buildDashboardRichTooltip(chart, chart.color, params)
        },
    grid: isMini
      ? { top: 6, left: 0, right: 0, bottom: 0, containLabel: false }
      : { top: 20, left: 36, right: 16, bottom: 28, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: chart.labels,
      show: !isMini,
      axisLine: isMini ? { show: false } : { lineStyle: { color: '#d9d9d9' } },
      axisTick: { show: false },
      axisLabel: { color: '#8c8c8c', fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      min: axisRange?.min,
      max: axisRange?.max,
      show: !isMini,
      axisLabel: { color: '#8c8c8c', fontSize: 11 },
      splitLine: isMini ? { show: false } : { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: chart.seriesName,
        type: 'line',
        smooth: false,
        showSymbol: false,
        data: chart.values,
        lineStyle: { width: isMini ? 2 : 3, color: chart.color },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: `${chart.color}${isMini ? '30' : '40'}` },
              { offset: 1, color: `${chart.color}08` }
            ]
          }
        }
      }
    ]
  };
}

function getMainChartOption(chart: KpiChartConfig | null): EChartsOption {
  const chartColor = chart?.color ?? '#1677ff';
  const axisRange = computeNumericAxisRange(chart?.values ?? [], {
    minPaddingRatio: 0.08,
    maxPaddingRatio: 0.08,
    minPaddingAbs: chart?.period === '60m' ? 5 : 10,
    roundMode: 'magnitude',
    splitNumber: 4,
  });

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => buildDashboardRichTooltip(chart ?? { seriesName: '', unit: '点', labels: [], values: [], color: chartColor, period: 'day' }, chartColor, params)
    },
    grid: {
      top: '10%',
      left: '3%',
      right: '4%',
      bottom: '8%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: chart?.labels ?? [],
      axisLine: { lineStyle: { color: '#d9d9d9' } }
    },
    yAxis: {
      type: 'value',
      min: axisRange?.min,
      max: axisRange?.max,
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: chart?.seriesName ?? '',
        type: 'line',
        smooth: false,
        showSymbol: false,
        data: chart?.values ?? [],
        lineStyle: { width: 3, color: chartColor },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: `${chartColor}40` },
              { offset: 1, color: `${chartColor}08` }
            ]
          }
        }
      }
    ]
  };
}

export default function Dashboard() {
  const navigate = useNavigate();
  const location = useLocation();
  const isPublicHome = location.pathname === '/';
  const { isLoggedIn, userInfo, logoutWithRevoke } = useAuth();
  const requestRef = useRef(0);

  const [marketData, setMarketData] = useState<DashboardMarketSnapshotResponse | null>(null);
  const [indexHistories, setIndexHistories] = useState<Record<string, MarketStockKlineResponse | null>>({});
  const [activeIndexSymbol, setActiveIndexSymbol] = useState('');
  const [chartRange, setChartRange] = useState<DashboardRange>('30d');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [error, setError] = useState('');
  const [historyError, setHistoryError] = useState('');

  const fetchIndexHistories = async (
    indices: DashboardMarketSnapshotResponse['indices'],
    range: DashboardRange,
    requestId: number,
  ) => {
    const topIndices = indices.slice(0, 4);
    if (!topIndices.length) {
      if (requestId !== requestRef.current) return;
      setIndexHistories({});
      setActiveIndexSymbol('');
      setHistoryError('');
      setHistoryLoading(false);
      return;
    }

    const missingIndices = topIndices.filter((item) => !(getDashboardHistoryCacheKey(item.symbol, range) in indexHistories));
    if (!missingIndices.length) {
      if (requestId !== requestRef.current) return;
      const firstAvailable = topIndices.find((item) => indexHistories[getDashboardHistoryCacheKey(item.symbol, range)]?.items?.length);
      setActiveIndexSymbol((prev) => {
        if (prev && indexHistories[getDashboardHistoryCacheKey(prev, range)]?.items?.length) {
          return prev;
        }
        if (prev && topIndices.some((item) => item.symbol === prev) && !firstAvailable) {
          return prev;
        }
        return firstAvailable?.symbol ?? topIndices[0]?.symbol ?? '';
      });
      setHistoryError(firstAvailable ? '' : '指数走势暂时不可用，主图已切换为市场概览。');
      setHistoryLoading(false);
      return;
    }

    setHistoryLoading(true);
    setHistoryError('');

    const results = await Promise.allSettled(
      missingIndices.map((item) => marketApi.getStockKlines(item.symbol, dashboardRangeQueryMap[range])),
    );

    if (requestId !== requestRef.current) {
      return;
    }

    const nextHistories: Record<string, MarketStockKlineResponse | null> = {};
    missingIndices.forEach((item, index) => {
      const result = results[index];
      nextHistories[getDashboardHistoryCacheKey(item.symbol, range)] = result.status === 'fulfilled'
        ? normalizeKlineResponse(result.value)
        : null;
    });

    const mergedHistories = { ...indexHistories, ...nextHistories };
    const firstAvailable = topIndices.find((item) => mergedHistories[getDashboardHistoryCacheKey(item.symbol, range)]?.items?.length);

    setIndexHistories((prev) => ({ ...prev, ...nextHistories }));
    setActiveIndexSymbol((prev) => {
      if (prev && mergedHistories[getDashboardHistoryCacheKey(prev, range)]?.items?.length) {
        return prev;
      }
      if (prev && topIndices.some((item) => item.symbol === prev) && !firstAvailable) {
        return prev;
      }
      return firstAvailable?.symbol ?? topIndices[0]?.symbol ?? '';
    });
    setHistoryError(firstAvailable ? '' : '指数走势暂时不可用，主图已回退为聚合走势。');
    setHistoryLoading(false);
  };

  const fetchMarketData = async (options?: { silent?: boolean }) => {
    const requestId = ++requestRef.current;
    const isSilent = options?.silent === true;
    if (isSilent) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError('');
    setHistoryError('');
    try {
      const res = await marketApi.getDashboardSnapshot();
      if (requestId !== requestRef.current) return;
      setMarketData(res);
      setActiveIndexSymbol((prev) => prev || res.indices[0]?.symbol || '');
      void fetchIndexHistories(res.indices ?? [], chartRange, requestId);
    } catch (err: unknown) {
      if (requestId !== requestRef.current) return;
      const apiError = err as { message?: string; data?: { message?: string } };
      if (!isSilent) {
        setMarketData(null);
        setIndexHistories({});
        setActiveIndexSymbol('');
        setError(apiError.message ?? apiError.data?.message ?? '市场快照加载失败');
      }
    } finally {
      if (requestId === requestRef.current) {
        if (isSilent) {
          setRefreshing(false);
        } else {
          setLoading(false);
        }
      }
    }
  };

  useEffect(() => {
    void fetchMarketData();
  }, []);

  useEffect(() => {
    const timer = window.setInterval(() => {
      void fetchMarketData({ silent: true });
    }, 30000);
    return () => window.clearInterval(timer);
  }, [chartRange]);

  useEffect(() => {
    if (!marketData?.indices?.length) {
      return;
    }
    const requestId = requestRef.current;
    void fetchIndexHistories(marketData.indices, chartRange, requestId);
  }, [chartRange, marketData]);

  const guardNavigate = (path: string) => {
    if (!isLoggedIn) navigate('/login', { state: { from: path } });
    else navigate(path);
  };

  const userMenuItems: MenuProps['items'] = [
    { key: 'profile', icon: <SettingOutlined />, label: '个人中心' },
    { type: 'divider' },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
  ];

  const handleUserMenu: MenuProps['onClick'] = ({ key }) => {
    if (key === 'profile') navigate('/profile');
    if (key === 'logout') void logoutWithRevoke('/');
  };

  const insightCards = useMemo(
    () => buildInsightCards(marketData, indexHistories, chartRange),
    [chartRange, indexHistories, marketData],
  );

  const activeInsightCard = useMemo(() => {
    if (!insightCards.length) {
      return null;
    }
    return insightCards.find((item) => item.symbol === activeIndexSymbol)
      ?? insightCards.find((item) => item.hasHistory)
      ?? insightCards[0];
  }, [activeIndexSymbol, insightCards]);

  const activeSnapshot = activeInsightCard?.snapshot ?? marketData?.indices?.[0] ?? null;

  const mainChart = useMemo(() => {
    if (!marketData || !activeInsightCard) {
      return null;
    }
    if (activeInsightCard.hasHistory) {
      return activeInsightCard.chart;
    }
    if (marketData.main_chart.series.length) {
      return buildFallbackChart(marketData, activeInsightCard.accent);
    }
    return null;
  }, [activeInsightCard, marketData]);

  const isFallbackChart = Boolean(activeInsightCard && !activeInsightCard.hasHistory && mainChart?.labels.length);

  const quickStats = useMemo(() => {
    if (!marketData?.stats?.length) {
      return [];
    }

    return marketData.stats.map((item) => ({
      label: item.label,
      value: item.value,
      color: statColorMap[item.label] || '#1677ff',
    }));
  }, [marketData]);

  const mainSummaryItems = activeSnapshot ? [
    { label: '今开', value: formatValue(activeSnapshot.open_price) },
    { label: '最高', value: formatValue(activeSnapshot.high_price) },
    { label: '最低', value: formatValue(activeSnapshot.low_price) },
    { label: '昨收', value: formatValue(activeSnapshot.prev_close) },
    { label: '成交量', value: formatCompactNumber(activeSnapshot.volume) },
    { label: '成交额', value: formatCompactNumber(activeSnapshot.turnover) },
  ] : [];

  return (
    <div style={{ padding: isPublicHome ? '24px' : '4px' }}>
      {isPublicHome && (
        <Card
          bordered={false}
          style={{
            marginBottom: 24,
            borderRadius: 20,
            background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
            boxShadow: '0 18px 40px rgba(22,119,255,0.18)'
          }}
          bodyStyle={{ padding: 28 }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
            <div>
              <Space size={12} wrap style={{ marginBottom: 12 }}>
                <Tag color="processing">AI 驱动</Tag>
                <Tag color="gold">实时市场洞察</Tag>
              </Space>
              <Title level={2} style={{ margin: 0, color: '#fff' }}>观势智投</Title>
              <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 720 }}>
                聚合指数趋势、仓位风险、交易效率与 AI 建议，快速看到今日市场温度与组合状态。
              </Paragraph>
            </div>
            <Space wrap>
              <Button ghost onClick={() => guardNavigate('/app/upload')}>上传记录</Button>
              <Button ghost onClick={() => guardNavigate('/app/portfolio')}>持仓总览</Button>
              <Button ghost onClick={() => guardNavigate('/app/market-trend')}>市场趋势</Button>
              <Button ghost onClick={() => guardNavigate('/app/analysis')}>AI 风险分析</Button>
              <Button ghost onClick={() => guardNavigate('/app/recommendation')}>AI 推荐</Button>
              <Button ghost onClick={() => guardNavigate('/app/prediction')}>趋势预测</Button>
              <Button ghost onClick={() => guardNavigate('/app/history')}>历史归档</Button>
              {isLoggedIn ? (
                <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenu }} placement="bottomRight" arrow>
                  <Button
                    type="primary"
                    icon={<Avatar size={18} icon={<UserOutlined />} style={{ background: 'rgba(255,255,255,0.25)', verticalAlign: 'middle' }} />}
                    style={{ display: 'flex', alignItems: 'center', gap: 6 }}
                  >
                    {userInfo?.username ?? '用户'}
                  </Button>
                </Dropdown>
              ) : (
                <Button type="primary" icon={<UserOutlined />} onClick={() => navigate('/login')}>
                  登录
                </Button>
              )}
            </Space>
          </div>
        </Card>
      )}

      {loading ? (
        <Skeleton active paragraph={{ rows: 12 }} />
      ) : error ? (
        <Card bordered={false} style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}>
          <Alert
            type="error"
            showIcon
            message={error}
            action={<Button size="small" onClick={() => void fetchMarketData()}>重试</Button>}
          />
        </Card>
      ) : !marketData?.indices?.length ? (
        <Card bordered={false} style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}>
          <Empty description="当前暂无可用市场快照数据" />
        </Card>
      ) : (
        <>
          <Row gutter={[16, 16]}>
            {insightCards.map((item) => {
              const isActive = item.symbol === activeInsightCard?.symbol;
              return (
                <Col xs={24} sm={12} lg={6} key={item.symbol}>
                  <Card
                    bordered={false}
                    hoverable
                    onClick={() => setActiveIndexSymbol(item.symbol)}
                    style={{
                      borderRadius: 16,
                      border: isActive ? '1px solid #91caff' : '1px solid transparent',
                      background: isActive ? '#f7fbff' : '#fff',
                      boxShadow: isActive ? '0 10px 28px rgba(22,119,255,0.10)' : '0 6px 22px rgba(15, 23, 42, 0.06)'
                    }}
                    bodyStyle={{ padding: 18 }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}>
                      <div style={{ minWidth: 0 }}>
                        <Text strong style={{ display: 'block', fontSize: 16 }}>{item.title}</Text>
                        <Text type="secondary">{item.symbol}</Text>
                      </div>
                      <Tag color={item.tagColor} style={{ marginInlineEnd: 0 }}>{item.tagText}</Tag>
                    </div>

                    <div style={{ marginTop: 14 }}>
                      <div style={{ fontSize: 32, lineHeight: 1.1, fontWeight: 700, color: item.accent }}>
                        {item.value.toFixed(item.precision)}
                        <span style={{ fontSize: 14, marginLeft: 4, color: '#8c8c8c' }}>{item.suffix}</span>
                      </div>
                      <Space size={12} wrap style={{ marginTop: 8 }}>
                        <Text type={toNumber(item.snapshot.change_amount) >= 0 ? 'success' : 'danger'}>
                          {formatSignedNumber(item.snapshot.change_amount)} 点
                        </Text>
                        <Text type={toNumber(item.snapshot.change_percent) >= 0 ? 'success' : 'danger'}>
                          {formatSignedPercent(item.snapshot.change_percent)}
                        </Text>
                      </Space>
                    </div>

                    <div style={{ marginTop: 14, borderRadius: 12, background: '#fafcff', border: '1px solid #eef2f6', padding: '6px 8px' }}>
                      {item.hasHistory ? (
                        <ReactECharts option={getKpiChartOption(item.chart, 'mini')} style={{ height: 64 }} />
                      ) : (
                        <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#8c8c8c', fontSize: 12 }}>
                          暂无走势
                        </div>
                      )}
                    </div>

                    <Row gutter={[8, 8]} style={{ marginTop: 14 }}>
                      <Col span={12}>
                        <div style={{ borderRadius: 10, background: '#fafafa', padding: '10px 12px' }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>区间高低</Text>
                          <div style={{ marginTop: 4, fontWeight: 600 }}>{formatValue(item.snapshot.high_price)} / {formatValue(item.snapshot.low_price)}</div>
                        </div>
                      </Col>
                      <Col span={12}>
                        <div style={{ borderRadius: 10, background: '#fafafa', padding: '10px 12px' }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>成交额</Text>
                          <div style={{ marginTop: 4, fontWeight: 600 }}>{formatCompactNumber(item.snapshot.turnover)}</div>
                        </div>
                      </Col>
                    </Row>

                    <Paragraph type="secondary" style={{ margin: '12px 0 0', minHeight: 44 }}>
                      {item.desc}
                    </Paragraph>
                  </Card>
                </Col>
              );
            })}
          </Row>

          <Card
            title={
              <Space size={8} wrap>
                <LineChartOutlined style={{ color: '#1677ff' }} />
                <span>{activeInsightCard ? `${activeInsightCard.title}${getDashboardRangeLabel(chartRange)}走势` : '指数走势'}</span>
              </Space>
            }
            extra={
              <Space size={8} wrap>
                <Segmented
                  options={dashboardRangeOptions}
                  value={chartRange}
                  onChange={(value) => setChartRange(value as DashboardRange)}
                />
                <Button type="text" size="small" icon={<ReloadOutlined />} loading={refreshing} onClick={() => void fetchMarketData({ silent: true })}>
                  刷新
                </Button>
              </Space>
            }
            bordered={false}
            style={{ marginTop: 24, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.06)' }}
          >
            {historyError ? <Alert type="warning" showIcon message={historyError} style={{ marginBottom: 16 }} /> : null}

            <Row gutter={[20, 20]} align="stretch">
              <Col xs={24} xl={16}>
                <Spin spinning={historyLoading}>
                  {mainChart?.labels.length ? (
                    <ReactECharts option={getMainChartOption(mainChart)} style={{ height: '420px' }} />
                  ) : (
                    <Empty description="当前暂无可用指数走势数据" />
                  )}
                </Spin>
              </Col>
              <Col xs={24} xl={8}>
                <div style={{ height: '100%', borderRadius: 16, border: '1px solid #eef2f6', background: '#fafcff', padding: 20 }}>
                  <Space direction="vertical" size={18} style={{ width: '100%' }}>
                    <div>
                      <Space size={8} wrap>
                        <Text strong style={{ fontSize: 18 }}>{activeSnapshot?.name ?? '—'}</Text>
                        <Tag color="blue">{activeSnapshot?.symbol ?? '—'}</Tag>
                        <Tag color={marketData.is_stale ? 'warning' : 'success'}>
                          {marketData.is_stale ? '更新较早' : '实时跟踪'}
                        </Tag>
                      </Space>
                      <div style={{ marginTop: 14, fontSize: 38, lineHeight: 1, fontWeight: 700, color: activeInsightCard?.accent ?? '#1677ff' }}>
                        {formatValue(activeSnapshot?.last_price)}
                      </div>
                      <Space size={12} wrap style={{ marginTop: 10 }}>
                        <Text type={toNumber(activeSnapshot?.change_amount) >= 0 ? 'success' : 'danger'} style={{ fontSize: 16 }}>
                          {formatSignedNumber(activeSnapshot?.change_amount)} 点
                        </Text>
                        <Text type={toNumber(activeSnapshot?.change_percent) >= 0 ? 'success' : 'danger'} style={{ fontSize: 16 }}>
                          {formatSignedPercent(activeSnapshot?.change_percent)}
                        </Text>
                      </Space>
                    </div>

                    <Row gutter={[12, 12]}>
                      {mainSummaryItems.map((item) => (
                        <Col span={12} key={item.label}>
                          <div style={{ borderRadius: 12, background: '#fff', border: '1px solid #eef2f6', padding: '12px 14px' }}>
                            <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                            <div style={{ marginTop: 4, fontSize: 16, fontWeight: 600 }}>{item.value}</div>
                          </div>
                        </Col>
                      ))}
                    </Row>

                    <div style={{ borderTop: '1px dashed #d9e6ff', paddingTop: 14 }}>
                      <Space size={[8, 8]} wrap>
                        <Tag color="processing" icon={<RiseOutlined />}>{formatValue(marketData.source)}</Tag>
                        <Tag color="default">{activeSnapshot?.market || '指数'}</Tag>
                        {isFallbackChart ? <Tag color="default">市场概览</Tag> : null}
                      </Space>
                      <div style={{ marginTop: 10, display: 'flex', flexDirection: 'column', gap: 4 }}>
                        <Text type="secondary">源数据时间 {formatValue(marketData.snapshot_time)}</Text>
                        <Text type="secondary">刷新入库时间 {formatValue(marketData.refreshed_at)}</Text>
                      </div>
                    </div>
                  </Space>
                </div>
              </Col>
            </Row>

            <Row gutter={[12, 12]} style={{ marginTop: 18 }}>
              {quickStats.map((item) => (
                <Col xs={12} md={8} xl={4} key={item.label}>
                  <div style={{ background: '#f8fafc', borderRadius: 12, padding: '14px 16px', border: '1px solid #eef2f6' }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                    <div style={{ marginTop: 6, fontSize: 22, fontWeight: 700, color: item.color }}>{item.value}</div>
                  </div>
                </Col>
              ))}
            </Row>
          </Card>

          <Card
            bordered={false}
            title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />指数对比明细</span>}
            style={{ marginTop: 16, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Row gutter={[16, 16]}>
              {marketData.indices.slice(0, 4).map((item) => {
                const isActive = item.symbol === activeInsightCard?.symbol;
                return (
                  <Col xs={24} md={12} key={item.symbol}>
                    <div
                      onClick={() => setActiveIndexSymbol(item.symbol)}
                      style={{
                        cursor: 'pointer',
                        height: '100%',
                        borderRadius: 16,
                        border: isActive ? '1px solid #91caff' : '1px solid #eef2f6',
                        background: isActive ? '#f7fbff' : '#fff',
                        padding: 18,
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}>
                        <div>
                          <Text strong style={{ display: 'block', fontSize: 16 }}>{item.name}</Text>
                          <Text type="secondary">{item.symbol}</Text>
                        </div>
                        <Tag color={toNumber(item.change_percent) >= 0 ? 'green' : 'red'}>
                          {formatSignedPercent(item.change_percent)}
                        </Tag>
                      </div>

                      <div style={{ marginTop: 14, display: 'flex', alignItems: 'baseline', gap: 10, flexWrap: 'wrap' }}>
                        <span style={{ fontSize: 28, fontWeight: 700 }}>{formatValue(item.last_price)}</span>
                        <Text type={toNumber(item.change_amount) >= 0 ? 'success' : 'danger'}>
                          {formatSignedNumber(item.change_amount)} 点
                        </Text>
                      </div>

                      <Row gutter={[12, 12]} style={{ marginTop: 14 }}>
                        {[
                          { label: '今开', value: formatValue(item.open_price) },
                          { label: '最高', value: formatValue(item.high_price) },
                          { label: '最低', value: formatValue(item.low_price) },
                          { label: '昨收', value: formatValue(item.prev_close) },
                          { label: '成交量', value: formatCompactNumber(item.volume) },
                          { label: '成交额', value: formatCompactNumber(item.turnover) },
                        ].map((metric) => (
                          <Col span={8} key={metric.label}>
                            <div style={{ borderRadius: 12, background: '#fafafa', padding: '10px 12px', minHeight: 68 }}>
                              <Text type="secondary" style={{ fontSize: 12 }}>{metric.label}</Text>
                              <div style={{ marginTop: 4, fontWeight: 600 }}>{metric.value}</div>
                            </div>
                          </Col>
                        ))}
                      </Row>
                    </div>
                  </Col>
                );
              })}
            </Row>
          </Card>

          <Card
            bordered={false}
            style={{ marginTop: 16, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Alert
              type={marketData.is_stale ? 'warning' : 'info'}
              showIcon
              icon={<BulbOutlined />}
              message={`数据结论：当前共追踪 ${marketData.indices.length} 个指数，${marketData.stats.find((item) => item.label === '上涨数')?.value ?? '—'} 个上涨，${marketData.stats.find((item) => item.label === '下跌数')?.value ?? '—'} 个下跌。`}
              description={
                <Space direction="vertical" size={6}>
                  <Text>
                    当前主图展示 {activeInsightCard?.title ?? marketData.main_chart.index_name}
                    {isFallbackChart ? '，该指数暂缺区间走势，已切换为市场概览。' : `，当前区间为${getDashboardRangeLabel(chartRange)}，可点击顶部或下方卡片切换其他指数。`}
                  </Text>
                  <Text>{marketData.is_stale ? '行情更新相对较早，可手动刷新查看最新走势。' : '当前行情更新正常，可直接查看指数走势与成交情况。'}</Text>
                </Space>
              }
            />
          </Card>

          <Card
            bordered={false}
            style={{ marginTop: 16, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Space size={10} wrap>
              <ThunderboltOutlined style={{ color: '#722ed1' }} />
              <Text strong>行情时间</Text>
              <Tag color={marketData.is_stale ? 'warning' : 'success'}>
                {marketData.is_stale ? '更新较早' : '已更新'}
              </Tag>
              <Text type="secondary">来源 {formatValue(marketData.source)}</Text>
              <Text type="secondary">行情时间 {formatValue(marketData.snapshot_time)}</Text>
              <Text type="secondary">更新时间 {formatValue(marketData.refreshed_at)}</Text>
            </Space>
          </Card>
        </>
      )}
    </div>
  );
}
