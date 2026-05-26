import { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, AutoComplete, Button, Card, Col, Empty, Input, List, Progress, Row, Segmented, Space, Spin, Statistic, Tag, Typography, message } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined, DeleteOutlined, FundProjectionScreenOutlined, ReloadOutlined, RiseOutlined, RobotOutlined, StarOutlined, StockOutlined, ThunderboltOutlined, WarningOutlined } from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi, marketApi } from '../api';
import { computeCandleAxisRange, computeNumericAxisRange } from '../utils/chartAxis';
import type {
  AnalysisCandidateResponse,
  MarketSnapshotResponse,
  MarketStockDetailResponse,
  MarketStockKlineResponse,
  StockNewsResponse,
  StockProfileResponse,
} from '../api/types';

const { Title, Text } = Typography;
const { Search } = Input;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

type ChartView = 'kline' | 'intraday';
type KlinePeriod = 'day' | 'week' | 'month' | '1m' | '5m' | '15m' | '60m';
type FavoriteSymbol = { symbol: string; asset_name: string; market: string };
type KlineItem = MarketStockKlineResponse['items'][number];
type LiveQuoteSnapshot = {
  symbol: string;
  name: string;
  last_price: string;
  change_amount: string;
  change_percent: string;
  turnover: string;
  turnover_rate: string;
  amplitude: string;
  volume_ratio: string;
  total_market_cap: string;
  fetched_at: string;
};
type QuoteFlashState = {
  direction: 'up' | 'down' | 'flat';
  priceDiff: number;
  changePercentDiff: number;
  touchedAt: number;
  fields: string[];
};

const FAVORITE_SYMBOLS_KEY = 'marketTrendFavoriteSymbols';
const ACTIVE_STOCK_REFRESH_MS = 10 * 1000;
const QUOTE_FLASH_WINDOW_MS = 3200;

const klinePeriodOptions: Array<{ label: string; value: KlinePeriod }> = [
  { label: '日线', value: 'day' },
  { label: '周线', value: 'week' },
  { label: '月线', value: 'month' },
  { label: '1分', value: '1m' },
  { label: '5分', value: '5m' },
  { label: '15分', value: '15m' },
  { label: '60分', value: '60m' },
];

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

function normalizeSymbolInput(value: string) {
  return value.trim().toUpperCase();
}

function stripMarketSuffix(symbol: string) {
  return normalizeSymbolInput(symbol).split('.')[0] || '';
}

function readFavoriteSymbols() {
  try {
    const raw = localStorage.getItem(FAVORITE_SYMBOLS_KEY);
    if (!raw) return [] as FavoriteSymbol[];
    const parsed = JSON.parse(raw) as FavoriteSymbol[];
    if (!Array.isArray(parsed)) return [];
    return parsed
      .map((item) => ({
        symbol: normalizeSymbolInput(item.symbol || ''),
        asset_name: item.asset_name || item.symbol || '',
        market: item.market || 'cn_stock',
      }))
      .filter((item) => item.symbol);
  } catch {
    return [];
  }
}

function writeFavoriteSymbols(items: FavoriteSymbol[]) {
  localStorage.setItem(FAVORITE_SYMBOLS_KEY, JSON.stringify(items));
}

function formatLargeNumber(value?: string) {
  const number = toNumber(value);
  if (Math.abs(number) >= 100000000) return `${(number / 100000000).toFixed(2)} 亿`;
  if (Math.abs(number) >= 10000) return `${(number / 10000).toFixed(2)} 万`;
  return number.toFixed(2);
}

function formatPercent(value?: string) {
  return `${toNumber(value).toFixed(2)}%`;
}

function formatPrice(value?: string) {
  return `¥${toNumber(value).toFixed(2)}`;
}

function formatSignedPercent(value: number) {
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}%`;
}

function average(values: number[]) {
  if (!values.length) return 0;
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function standardDeviation(values: number[]) {
  if (values.length < 2) return 0;
  const avg = average(values);
  const variance = average(values.map((value) => (value - avg) ** 2));
  return Math.sqrt(variance);
}

function movingAverage(values: number[], windowSize: number) {
  return values.map((_, index) => {
    if (index + 1 < windowSize) return null;
    return Number(average(values.slice(index + 1 - windowSize, index + 1)).toFixed(3));
  });
}

function getChangeColor(value?: string) {
  const number = toNumber(value);
  if (number > 0) return '#ff4d4f';
  if (number < 0) return '#52c41a';
  return '#1677ff';
}

function formatChangeText(changeAmount?: string, changePercent?: string) {
  const amount = toNumber(changeAmount);
  const sign = amount > 0 ? '+' : '';
  return `${sign}${amount.toFixed(2)} / ${formatPercent(changePercent)}`;
}

function buildQuoteSnapshotFromCandidate(item: AnalysisCandidateResponse): LiveQuoteSnapshot {
  return {
    symbol: item.symbol,
    name: item.asset_name || item.symbol,
    last_price: item.last_price || '0',
    change_amount: '0',
    change_percent: item.change_percent || '0',
    turnover: '0',
    turnover_rate: '0',
    amplitude: '0',
    volume_ratio: '0',
    total_market_cap: '0',
    fetched_at: '',
  };
}

function buildQuoteSnapshotFromDetail(detail: MarketStockDetailResponse | null | undefined): LiveQuoteSnapshot | null {
  if (!detail?.symbol) return null;
  return {
    symbol: detail.symbol,
    name: detail.name || detail.symbol,
    last_price: detail.last_price || '0',
    change_amount: detail.change_amount || '0',
    change_percent: detail.change_percent || '0',
    turnover: detail.turnover || '0',
    turnover_rate: detail.turnover_rate || '0',
    amplitude: detail.amplitude || '0',
    volume_ratio: detail.volume_ratio || '0',
    total_market_cap: detail.total_market_cap || '0',
    fetched_at: detail.fetched_at || '',
  };
}

function buildQuoteSnapshotFromProfile(profile: StockProfileResponse | null | undefined): LiveQuoteSnapshot | null {
  if (!profile?.symbol) return null;
  return {
    symbol: profile.symbol,
    name: profile.name || profile.symbol,
    last_price: profile.last_price || '0',
    change_amount: profile.change_amount || '0',
    change_percent: profile.change_percent || '0',
    turnover: profile.turnover || '0',
    turnover_rate: profile.turnover_rate || '0',
    amplitude: profile.amplitude || '0',
    volume_ratio: profile.volume_ratio || '0',
    total_market_cap: profile.total_market_cap || '0',
    fetched_at: profile.fetched_at || '',
  };
}

function buildQuoteFlash(previous: LiveQuoteSnapshot | undefined, next: LiveQuoteSnapshot): QuoteFlashState | null {
  if (!previous) return null;
  const fields: string[] = [];
  if (toNumber(previous.last_price) !== toNumber(next.last_price)) fields.push('last_price');
  if (toNumber(previous.change_amount) !== toNumber(next.change_amount)) fields.push('change_amount');
  if (toNumber(previous.change_percent) !== toNumber(next.change_percent)) fields.push('change_percent');
  if (toNumber(previous.turnover) !== toNumber(next.turnover)) fields.push('turnover');
  if (toNumber(previous.turnover_rate) !== toNumber(next.turnover_rate)) fields.push('turnover_rate');
  if (toNumber(previous.amplitude) !== toNumber(next.amplitude)) fields.push('amplitude');
  if (toNumber(previous.volume_ratio) !== toNumber(next.volume_ratio)) fields.push('volume_ratio');
  if (toNumber(previous.total_market_cap) !== toNumber(next.total_market_cap)) fields.push('total_market_cap');
  if (!fields.length) return null;

  const priceDiff = toNumber(next.last_price) - toNumber(previous.last_price);
  const changePercentDiff = toNumber(next.change_percent) - toNumber(previous.change_percent);
  return {
    direction: priceDiff > 0 ? 'up' : priceDiff < 0 ? 'down' : 'flat',
    priceDiff,
    changePercentDiff,
    touchedAt: Date.now(),
    fields,
  };
}

function getFlashAccentColor(flash?: QuoteFlashState) {
  if (!flash) return '#1677ff';
  if (flash.direction === 'up') return '#ff4d4f';
  if (flash.direction === 'down') return '#52c41a';
  return '#1677ff';
}

function formatPriceDelta(value: number) {
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}`;
}

function normalizeDetailLabel(value?: string) {
  const trimmed = value?.trim() ?? '';
  if (!trimmed || trimmed === '-' || trimmed === '—') return '';
  if (/^GP-[A-Z0-9-]+$/.test(trimmed)) return '';
  return trimmed;
}

function normalizeConcepts(values?: string[]) {
  if (!values?.length) return [] as string[];
  return Array.from(new Set(values.map((item) => item.trim()).filter((item) => item && item !== '腾讯行情')));
}

function marketLabel(market?: string) {
  switch (market) {
    case 'cn_index':
      return 'A 股指数';
    case 'cn_stock':
      return 'A 股个股';
    default:
      return market || '市场';
  }
}

function getKlinePeriodLabel(period: KlinePeriod) {
  return klinePeriodOptions.find((item) => item.value === period)?.label ?? '日线';
}

function resolveMarketSymbol(symbol: string, candidates: AnalysisCandidateResponse[], favorites: FavoriteSymbol[]) {
  const normalized = normalizeSymbolInput(symbol);
  if (!normalized) return '';
  const availableSymbols = [
    ...candidates.map((item) => item.symbol),
    ...favorites.map((item) => item.symbol),
  ].map(normalizeSymbolInput).filter(Boolean);
  return availableSymbols.find((item) => item === normalized)
    || availableSymbols.find((item) => stripMarketSuffix(item) === normalized)
    || normalized;
}

function formatKlineAxisTime(value: string, period: KlinePeriod) {
  if (period === '5m' || period === '15m' || period === '60m') return value.slice(5, 16);
  if (period === 'month') return value.slice(0, 7);
  return value.slice(2, 10);
}

function buildKlineOption(kline: MarketStockKlineResponse | null, period: KlinePeriod): EChartsOption {
  const items = kline?.items ?? [];
  const priceAxisRange = computeCandleAxisRange(
    items.map((item) => toNumber(item.high_price)),
    items.map((item) => toNumber(item.low_price)),
    { paddingRatio: 0.02, minPaddingAbs: period === '1m' || period === '5m' || period === '15m' || period === '60m' ? 0.02 : 0.1, roundMode: 'step', splitNumber: 4 },
  );
  const volumeAxisRange = computeNumericAxisRange(items.map((item) => toNumber(item.volume)), {
    minPaddingRatio: 0,
    maxPaddingRatio: 0.1,
    minPaddingAbs: 1,
    includeZero: true,
  });

  return {
    animation: false,
    tooltip: { trigger: 'axis' },
    axisPointer: { link: [{ xAxisIndex: 'all' }] },
    grid: [
      { top: 24, left: 36, right: 20, height: 220, containLabel: true },
      { left: 36, right: 20, top: 272, height: 88, containLabel: true },
    ],
    xAxis: [
      { type: 'category', data: items.map((item) => formatKlineAxisTime(item.bar_time, period)), boundaryGap: true, axisLine: { lineStyle: { color: '#d9d9d9' } }, axisLabel: { color: '#8c8c8c', fontSize: 11 } },
      { type: 'category', gridIndex: 1, data: items.map((item) => formatKlineAxisTime(item.bar_time, period)), boundaryGap: true, axisLine: { lineStyle: { color: '#d9d9d9' } }, axisLabel: { show: false } },
    ],
    yAxis: [
      { type: 'value', scale: true, min: priceAxisRange?.min, max: priceAxisRange?.max, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }, axisLabel: { color: '#8c8c8c', fontSize: 11 } },
      { type: 'value', gridIndex: 1, min: 0, max: volumeAxisRange?.max, splitNumber: 2, splitLine: { show: false }, axisLabel: { color: '#8c8c8c', fontSize: 11 } },
    ],
    series: [
      { name: 'K线', type: 'candlestick', data: items.map((item) => [toNumber(item.open_price), toNumber(item.close_price), toNumber(item.low_price), toNumber(item.high_price)]), itemStyle: { color: '#ff4d4f', color0: '#52c41a', borderColor: '#ff4d4f', borderColor0: '#52c41a' } },
      { name: '成交量', type: 'bar', xAxisIndex: 1, yAxisIndex: 1, data: items.map((item) => toNumber(item.volume)), itemStyle: { color: '#91caff' } },
    ],
  };
}

function buildIntradayOption(kline: MarketStockKlineResponse | null, period: KlinePeriod): EChartsOption {
  const items = kline?.items ?? [];
  const prices = items.map((item) => toNumber(item.close_price));
  const prevClose = items.length ? toNumber(items[0].open_price) || toNumber(items[0].close_price) : 0;
  const axisRange = computeNumericAxisRange(prices.length ? [...prices, ...(prevClose > 0 ? [prevClose] : [])] : [0], {
    minPaddingRatio: 0.015,
    maxPaddingRatio: 0.015,
    minPaddingAbs: period === '1m' ? 0.01 : 0.02,
    roundMode: 'step',
    splitNumber: 4,
  });

  return {
    tooltip: { trigger: 'axis' },
    grid: [
      { top: 24, left: 36, right: 18, height: 240, containLabel: true },
      { left: 36, right: 18, top: 286, height: 82, containLabel: true },
    ],
    xAxis: [
      {
        type: 'category',
        boundaryGap: false,
        data: items.map((item) => formatKlineAxisTime(item.bar_time, period)),
        axisLine: { lineStyle: { color: '#d9d9d9' } },
        axisLabel: { color: '#8c8c8c', fontSize: 11 },
      },
      {
        type: 'category',
        gridIndex: 1,
        boundaryGap: false,
        data: items.map((item) => formatKlineAxisTime(item.bar_time, period)),
        axisLine: { lineStyle: { color: '#d9d9d9' } },
        axisLabel: { show: false },
      },
    ],
    yAxis: [
      {
        type: 'value',
        min: axisRange?.min,
        max: axisRange?.max,
        splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } },
        axisLabel: { color: '#8c8c8c', fontSize: 11 },
      },
      {
        type: 'value',
        gridIndex: 1,
        min: 0,
        splitLine: { show: false },
        axisLabel: { color: '#8c8c8c', fontSize: 10 },
      },
    ],
    series: [
      {
        name: '价格',
        type: 'line',
        smooth: false,
        showSymbol: false,
        data: prices,
        lineStyle: { width: 3, color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22,119,255,0.22)' },
              { offset: 1, color: 'rgba(22,119,255,0.03)' },
            ],
          },
        },
        markLine: prevClose > 0 ? {
          silent: true,
          symbol: 'none',
          lineStyle: { color: '#bfbfbf', type: 'dashed' },
          data: [{ yAxis: prevClose }],
        } : undefined,
      },
      {
        name: '成交量',
        type: 'bar',
        xAxisIndex: 1,
        yAxisIndex: 1,
        data: items.map((item) => toNumber(item.volume)),
        itemStyle: {
          color: (params: any) => {
            const item = items[params.dataIndex];
            return toNumber(item?.close_price) >= toNumber(item?.open_price) ? '#ff7875' : '#73d13d';
          },
        },
      },
    ],
  };
}

function buildPriceMomentumOption(kline: MarketStockKlineResponse | null, period: KlinePeriod): EChartsOption {
  const items = kline?.items ?? [];
  const fullCloses = items.map((item) => toNumber(item.close_price));
  const fullMa5 = movingAverage(fullCloses, 5);
  const fullMa20 = movingAverage(fullCloses, 20);
  const displayCount = period === '5m' || period === '15m' ? 20 : 18;
  const startIndex = Math.max(items.length - displayCount, 0);
  const displayItems = items.slice(startIndex);
  const closes = fullCloses.slice(startIndex);
  const ma5 = fullMa5.slice(startIndex);
  const ma20 = fullMa20.slice(startIndex);
  const axisRange = computeNumericAxisRange([...closes, ...ma5.filter((value): value is number => value !== null), ...ma20.filter((value): value is number => value !== null)], {
    minPaddingRatio: 0.04,
    maxPaddingRatio: 0.04,
    minPaddingAbs: period === '1m' || period === '5m' || period === '15m' || period === '60m' ? 0.02 : 0.1,
    roundMode: 'step',
    splitNumber: 4,
  });

  return {
    tooltip: { trigger: 'axis' },
    legend: { top: 0, right: 8, itemWidth: 10, itemHeight: 10 },
    grid: { top: 34, left: 36, right: 18, bottom: 28, containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: displayItems.map((item) => formatKlineAxisTime(item.bar_time, period)), axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    yAxis: { type: 'value', min: axisRange?.min, max: axisRange?.max, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }, axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    series: [
      { name: '收盘', type: 'line', smooth: false, showSymbol: false, symbol: 'circle', emphasis: { focus: 'series' }, data: closes, lineStyle: { width: 3, color: '#1677ff' } },
      { name: 'MA5', type: 'line', smooth: false, showSymbol: false, symbol: 'circle', emphasis: { focus: 'series' }, data: ma5, lineStyle: { width: 2, color: '#fa8c16' } },
      { name: 'MA20', type: 'line', smooth: false, showSymbol: false, symbol: 'circle', emphasis: { focus: 'series' }, data: ma20, lineStyle: { width: 2, color: '#722ed1' } },
    ],
  };
}

function buildVolumeTrendOption(kline: MarketStockKlineResponse | null, period: KlinePeriod): EChartsOption {
  const items = kline?.items ?? [];
  const volumes = items.map((item) => toNumber(item.volume));
  const volumeAxisRange = computeNumericAxisRange(volumes, {
    minPaddingRatio: 0,
    maxPaddingRatio: 0.12,
    minPaddingAbs: 1,
    includeZero: true,
  });

  return {
    tooltip: { trigger: 'axis' },
    grid: { top: 24, left: 36, right: 18, bottom: 28, containLabel: true },
    xAxis: { type: 'category', data: items.map((item) => formatKlineAxisTime(item.bar_time, period)), axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    yAxis: { type: 'value', min: 0, max: volumeAxisRange?.max, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }, axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    series: [{
      name: '成交量',
      type: 'bar',
      data: volumes,
      itemStyle: {
        color: (params: any) => {
          const item = items[params.dataIndex];
          return toNumber(item?.close_price) >= toNumber(item?.open_price) ? '#ff7875' : '#73d13d';
        },
      },
    }],
  };
}

function buildReturnDistributionOption(kline: MarketStockKlineResponse | null): EChartsOption {
  const items = kline?.items ?? [];
  const buckets = [
    { label: '<-3%', min: -Infinity, max: -3, count: 0 },
    { label: '-3~-1%', min: -3, max: -1, count: 0 },
    { label: '-1~0%', min: -1, max: 0, count: 0 },
    { label: '0~1%', min: 0, max: 1, count: 0 },
    { label: '1~3%', min: 1, max: 3, count: 0 },
    { label: '>3%', min: 3, max: Infinity, count: 0 },
  ];
  items.forEach((item, index) => {
    if (index === 0) return;
    const previousClose = toNumber(items[index - 1].close_price);
    const close = toNumber(item.close_price);
    if (previousClose <= 0 || close <= 0) return;
    const dailyReturn = ((close - previousClose) / previousClose) * 100;
    const bucket = buckets.find((candidate) => dailyReturn >= candidate.min && dailyReturn < candidate.max);
    if (bucket) bucket.count += 1;
  });

  return {
    tooltip: { trigger: 'axis' },
    grid: { top: 24, left: 34, right: 16, bottom: 28, containLabel: true },
    xAxis: { type: 'category', data: buckets.map((item) => item.label), axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }, axisLabel: { color: '#8c8c8c', fontSize: 10 } },
    series: [{ name: '天数', type: 'bar', data: buckets.map((item) => item.count), itemStyle: { color: '#1677ff', borderRadius: [6, 6, 0, 0] } }],
  };
}

function buildTrendMetrics(items: KlineItem[]) {
  const closes = items.map((item) => toNumber(item.close_price)).filter((value) => value > 0);
  const highs = items.map((item) => toNumber(item.high_price)).filter((value) => value > 0);
  const lows = items.map((item) => toNumber(item.low_price)).filter((value) => value > 0);
  const volumes = items.map((item) => toNumber(item.volume)).filter((value) => value >= 0);
  const latestClose = closes.at(-1) ?? 0;
  const previousClose = closes.at(-2) ?? 0;
  const high = highs.length ? Math.max(...highs) : 0;
  const low = lows.length ? Math.min(...lows) : 0;
  const ma5 = closes.length >= 5 ? average(closes.slice(-5)) : 0;
  const ma20 = closes.length >= 20 ? average(closes.slice(-20)) : 0;
  const recentVolume = volumes.length ? average(volumes.slice(-5)) : 0;
  const baseVolume = volumes.length >= 20 ? average(volumes.slice(-20, -5)) : average(volumes.slice(0, -5));
  const returns = closes.slice(1).map((close, index) => {
    const previous = closes[index];
    return previous > 0 ? ((close - previous) / previous) * 100 : 0;
  });
  const rangePosition = high > low && latestClose > 0 ? ((latestClose - low) / (high - low)) * 100 : 0;
  const maxDrawdown = closes.reduce((state, close) => {
    const peak = Math.max(state.peak, close);
    const drawdown = peak > 0 ? ((close - peak) / peak) * 100 : 0;
    return { peak, maxDrawdown: Math.min(state.maxDrawdown, drawdown) };
  }, { peak: 0, maxDrawdown: 0 });

  return {
    sampleSize: closes.length,
    latestClose,
    dayReturn: previousClose > 0 ? ((latestClose - previousClose) / previousClose) * 100 : 0,
    rangeHigh: high,
    rangeLow: low,
    rangePosition,
    ma5,
    ma20,
    maBias: ma20 > 0 ? ((latestClose - ma20) / ma20) * 100 : 0,
    volumeRatio: baseVolume > 0 ? recentVolume / baseVolume : 0,
    volatility: standardDeviation(returns.slice(-20)),
    maxDrawdown: maxDrawdown.maxDrawdown,
  };
}

export default function MarketTrendPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);
  const [profileLoading, setProfileLoading] = useState(false);
  const [newsLoading, setNewsLoading] = useState(false);
  const [klineLoading, setKlineLoading] = useState(false);
  const [intradayLoading, setIntradayLoading] = useState(false);
  const [error, setError] = useState('');
  const [detailError, setDetailError] = useState('');
  const [klineError, setKlineError] = useState('');
  const [intradayError, setIntradayError] = useState('');
  const [candidates, setCandidates] = useState<AnalysisCandidateResponse[]>([]);
  const [favoriteSymbols, setFavoriteSymbols] = useState<FavoriteSymbol[]>(() => readFavoriteSymbols());
  const [favoriteQuery, setFavoriteQuery] = useState('');
  const [marketSearchResults, setMarketSearchResults] = useState<MarketSnapshotResponse[]>([]);
  const [history, setHistory] = useState<MarketSnapshotResponse[]>([]);
  const [detail, setDetail] = useState<MarketStockDetailResponse | null>(null);
  const [profile, setProfile] = useState<StockProfileResponse | null>(null);
  const [stockNews, setStockNews] = useState<StockNewsResponse | null>(null);
  const [kline, setKline] = useState<MarketStockKlineResponse | null>(null);
  const [intradayKline, setIntradayKline] = useState<MarketStockKlineResponse | null>(null);
  const [selectedSymbol, setSelectedSymbol] = useState('');
  const [chartView, setChartView] = useState<ChartView>('kline');
  const [klinePeriod, setKlinePeriod] = useState<KlinePeriod>('day');
  const [showAllConcepts, setShowAllConcepts] = useState(false);
  const [liveQuotes, setLiveQuotes] = useState<Record<string, LiveQuoteSnapshot>>({});
  const [quoteFlashes, setQuoteFlashes] = useState<Record<string, QuoteFlashState>>({});
  const loadRequestRef = useRef(0);
  const refreshInFlightRef = useRef(false);
  const flashTimersRef = useRef<Record<string, number>>({});

  const commitLiveQuotes = (entries: Array<LiveQuoteSnapshot | null | undefined>) => {
    const normalizedEntries = entries.filter((entry): entry is LiveQuoteSnapshot => Boolean(entry?.symbol));
    if (!normalizedEntries.length) return;

    const pendingFlashes: Record<string, QuoteFlashState> = {};
    setLiveQuotes((previous) => {
      const next = { ...previous };
      normalizedEntries.forEach((entry) => {
        const flash = buildQuoteFlash(previous[entry.symbol], entry);
        if (flash) pendingFlashes[entry.symbol] = flash;
        next[entry.symbol] = entry;
      });
      return next;
    });

    if (!Object.keys(pendingFlashes).length) return;

    setQuoteFlashes((previous) => ({ ...previous, ...pendingFlashes }));
    Object.entries(pendingFlashes).forEach(([symbol]) => {
      if (flashTimersRef.current[symbol]) {
        window.clearTimeout(flashTimersRef.current[symbol]);
      }
      flashTimersRef.current[symbol] = window.setTimeout(() => {
        setQuoteFlashes((previous) => {
          const next = { ...previous };
          delete next[symbol];
          return next;
        });
        delete flashTimersRef.current[symbol];
      }, QUOTE_FLASH_WINDOW_MS);
    });
  };

  const loadKline = async (symbol: string, period: KlinePeriod, refresh = false) => {
    if (!symbol) {
      setKline(null);
      return;
    }
    setKlineLoading(true);
    setKlineError('');
    try {
      const res = await marketApi.getStockKlines(symbol, { period, adjust: 'qfq', limit: 60, refresh });
      setKline(res);
    } catch (err: any) {
      setKline(null);
      setKlineError(err?.message ?? err?.data?.message ?? 'K线数据暂时不可用');
    } finally {
      setKlineLoading(false);
    }
  };

  const loadSupplementalData = async (symbol: string) => {
    if (!symbol) return;
    setProfileLoading(true);
    setNewsLoading(true);
    const [profileRes, newsRes] = await Promise.allSettled([
      marketApi.getStockProfile(symbol),
      marketApi.getStockNews(symbol, { limit: 10 }),
    ]);
    if (profileRes.status === 'fulfilled') {
      setProfile(profileRes.value);
      commitLiveQuotes([buildQuoteSnapshotFromProfile(profileRes.value)]);
    } else {
      setProfile(null);
    }
    if (newsRes.status === 'fulfilled') {
      setStockNews(newsRes.value);
    } else {
      setStockNews(null);
    }
    setProfileLoading(false);
    setNewsLoading(false);
  };

  const load = async (preferSymbol?: string, forceRefresh = false, period: KlinePeriod = klinePeriod) => {
    const requestId = ++loadRequestRef.current;
    setLoading(true);
    setError('');
    setDetailError('');
    setKlineError('');
    try {
      const candidateRes = await analysisApi.getCandidates();
      if (requestId !== loadRequestRef.current) return;
      const candidateList = candidateRes.candidates ?? [];
      const favorites = readFavoriteSymbols();
      setCandidates(candidateList);
      setFavoriteSymbols(favorites);
      setLiveQuotes((previous) => {
        const next = { ...previous };
        candidateList.forEach((item) => {
          if (!next[item.symbol]) {
            next[item.symbol] = buildQuoteSnapshotFromCandidate(item);
          }
        });
        return next;
      });
      const requestedSymbol = preferSymbol || searchParams.get('symbol') || candidateRes.default_symbol || candidateList[0]?.symbol || favorites[0]?.symbol || '';
      const nextSymbol = resolveMarketSymbol(requestedSymbol, candidateList, favorites);
      setSelectedSymbol(nextSymbol);
      setShowAllConcepts(false);
      if (!nextSymbol) {
        setHistory([]);
        setDetail(null);
        setProfile(null);
        setStockNews(null);
        setKline(null);
        setIntradayKline(null);
        return;
      }

      setSearchParams({ symbol: nextSymbol });
      setProfile(null);
      setStockNews(null);
      setDetailLoading(true);
      setKlineLoading(true);
      setProfileLoading(true);
      setNewsLoading(true);

      void loadSupplementalData(nextSymbol);
      void loadIntraday(nextSymbol, forceRefresh);

      const [historyRes, detailRes, klineRes] = await Promise.allSettled([
        marketApi.getSnapshotHistory({ symbol: nextSymbol, limit: 30 }),
        marketApi.getStockDetail(nextSymbol, forceRefresh ? { refresh: true } : undefined),
        marketApi.getStockKlines(nextSymbol, { period, adjust: 'qfq', limit: 60, refresh: forceRefresh }),
      ]);
      if (requestId !== loadRequestRef.current) return;

      setHistory(historyRes.status === 'fulfilled' ? [...historyRes.value].reverse() : []);
      if (detailRes.status === 'fulfilled') {
        setDetail(detailRes.value);
        commitLiveQuotes([buildQuoteSnapshotFromDetail(detailRes.value)]);
      } else {
        setDetail(null);
        setDetailError(detailRes.reason?.message ?? detailRes.reason?.data?.message ?? '详细行情暂时不可用');
      }
      if (klineRes.status === 'fulfilled') {
        setKline(klineRes.value);
      } else {
        setKline(null);
        setKlineError(klineRes.reason?.message ?? klineRes.reason?.data?.message ?? 'K线数据暂时不可用');
      }
      if (historyRes.status === 'rejected' && detailRes.status === 'rejected' && klineRes.status === 'rejected') {
        throw detailRes.reason ?? klineRes.reason ?? historyRes.reason;
      }
    } catch (err: any) {
      if (requestId !== loadRequestRef.current) return;
      setCandidates([]);
      setHistory([]);
      setDetail(null);
      setProfile(null);
      setStockNews(null);
      setKline(null);
      setError(err?.message ?? err?.data?.message ?? '市场趋势加载失败');
    } finally {
      if (requestId === loadRequestRef.current) {
        setLoading(false);
        setDetailLoading(false);
        setKlineLoading(false);
      }
    }
  };

  useEffect(() => {
    void load(undefined, false, klinePeriod);
  }, []);

  useEffect(() => {
    const query = normalizeSymbolInput(favoriteQuery);
    if (query.length < 2) {
      setMarketSearchResults([]);
      return;
    }
    const timer = window.setTimeout(() => {
      void marketApi.searchStocks({ q: query, limit: 12 })
        .then(setMarketSearchResults)
        .catch(() => setMarketSearchResults([]));
    }, 250);
    return () => window.clearTimeout(timer);
  }, [favoriteQuery]);

  const candidateList = useMemo(() => {
    const bySymbol = new Map<string, AnalysisCandidateResponse>();
    candidates.forEach((item) => bySymbol.set(item.symbol, item));
    favoriteSymbols.forEach((favorite) => {
      const existing = bySymbol.get(favorite.symbol);
      if (existing) {
        bySymbol.set(favorite.symbol, {
          ...existing,
          sources: existing.sources.some((source) => source.type === 'favorite') ? existing.sources : [...existing.sources, { type: 'favorite' }],
        });
        return;
      }
      bySymbol.set(favorite.symbol, {
        symbol: favorite.symbol,
        asset_name: favorite.asset_name || favorite.symbol,
        asset_type: 'stock',
        market: favorite.market || 'cn_stock',
        sources: [{ type: 'favorite' }],
        is_held: false,
        trade_count: 0,
        last_price: '',
        change_percent: '',
      });
    });
    return Array.from(bySymbol.values()).sort((a, b) => {
      const aFavorite = a.sources.some((source) => source.type === 'favorite');
      const bFavorite = b.sources.some((source) => source.type === 'favorite');
      return Number(b.is_held) - Number(a.is_held)
        || Number(bFavorite) - Number(aFavorite)
        || (b.trade_count ?? 0) - (a.trade_count ?? 0)
        || a.symbol.localeCompare(b.symbol);
    });
  }, [candidates, favoriteSymbols]);

  const selectedCandidate = useMemo(() => candidateList.find((item) => item.symbol === selectedSymbol) ?? null, [candidateList, selectedSymbol]);
  const favoriteOptions = useMemo(() => {
    const query = normalizeSymbolInput(favoriteQuery);
    if (!query) return [];
    const optionMap = new Map<string, { value: string; label: React.ReactNode }>();
    marketSearchResults.forEach((item) => {
      optionMap.set(item.symbol, {
        value: item.symbol,
        label: (
          <Space split={<Text type="secondary">|</Text>}>
            <Text strong>{item.symbol}</Text>
            <Text>{item.name || item.symbol}</Text>
            <Tag color="blue">市场</Tag>
          </Space>
        ),
      });
    });
    candidateList
      .filter((item) => item.symbol.includes(query) || (item.asset_name || '').toUpperCase().includes(query))
      .slice(0, 8)
      .forEach((item) => {
        if (optionMap.has(item.symbol)) return;
        optionMap.set(item.symbol, {
          value: item.symbol,
          label: (
            <Space split={<Text type="secondary">|</Text>}>
              <Text strong>{item.symbol}</Text>
              <Text>{item.asset_name || item.symbol}</Text>
              {item.is_held ? <Tag color="success">持仓</Tag> : item.sources.some((source) => source.type === 'favorite') ? <Tag color="gold">自选</Tag> : <Tag>历史</Tag>}
            </Space>
          ),
        });
      });
    if (!optionMap.has(query) && /^[A-Z0-9.]{2,16}$/.test(query)) {
      optionMap.set(query, { value: query, label: <Text>添加 <Text strong>{query}</Text> 为自选</Text> });
    }
    return Array.from(optionMap.values()).slice(0, 12);
  }, [candidateList, favoriteQuery, marketSearchResults]);
  const selectedLiveQuote = selectedSymbol ? liveQuotes[selectedSymbol] : undefined;
  const selectedFlash = selectedSymbol ? quoteFlashes[selectedSymbol] : undefined;
  const latestPoint = history[history.length - 1] ?? null;
  const latestPrice = selectedLiveQuote?.last_price || profile?.last_price || detail?.last_price || latestPoint?.last_price;
  const latestChangePercent = selectedLiveQuote?.change_percent || profile?.change_percent || detail?.change_percent || latestPoint?.change_percent;
  const latestChangeAmount = selectedLiveQuote?.change_amount || profile?.change_amount || detail?.change_amount || latestPoint?.change_amount;
  const industryLabel = normalizeDetailLabel(profile?.industry || detail?.industry);
  const regionLabel = normalizeDetailLabel(profile?.region || detail?.region);
  const metaSummary = [industryLabel, regionLabel].filter(Boolean).join(' · ');
  const boardMemberships = profile?.boards ?? [];
  const boardByName = new Map(boardMemberships.map((board) => [board.name, board]));
  const industryBoard = industryLabel ? boardMemberships.find((board) => board.board_type === 'industry' && board.name === industryLabel) : undefined;
  const conceptList = normalizeConcepts(profile?.concepts?.length ? profile.concepts : detail?.concepts);
  const visibleConcepts = showAllConcepts ? conceptList : conceptList.slice(0, 8);
  const visibleBoards = boardMemberships.slice(0, 10);
  const newsItems = useMemo(() => {
    const items = [...(stockNews?.items ?? [])];
    items.sort((a, b) => {
      const left = a.published_at ? new Date(a.published_at).getTime() : 0;
      const right = b.published_at ? new Date(b.published_at).getTime() : 0;
      return right - left;
    });
    return items;
  }, [stockNews]);
  const recentNewsCount = newsItems.filter((item) => item.is_recent).length;
  const newestNews = newsItems[0] ?? null;
  const intradayPeriod = (intradayKline?.period as KlinePeriod | undefined) || '5m';
  const chartTitle = chartView === 'kline' ? `${getKlinePeriodLabel(klinePeriod)} K 线` : `${getKlinePeriodLabel(intradayPeriod)} 盘中走势`;
  const trendMetrics = useMemo(() => buildTrendMetrics(kline?.items ?? []), [kline]);
  const latestAboveMa20 = trendMetrics.ma20 > 0 && trendMetrics.latestClose >= trendMetrics.ma20;
  const trendBiasLabel = latestAboveMa20 ? '站上 MA20' : trendMetrics.ma20 > 0 ? '低于 MA20' : '均线不足';
  const rangePositionColor = trendMetrics.rangePosition >= 80 ? '#ff4d4f' : trendMetrics.rangePosition <= 20 ? '#52c41a' : '#1677ff';
  const alertItems = [
    trendMetrics.sampleSize < 20 ? { color: 'warning', icon: <WarningOutlined />, text: `K 线样本仅 ${trendMetrics.sampleSize} 条，均线和波动判断可信度有限。` } : null,
    trendMetrics.rangePosition >= 85 ? { color: 'red', icon: <WarningOutlined />, text: `价格位于当前样本区间高位 ${trendMetrics.rangePosition.toFixed(0)}%，追涨需要控制仓位。` } : null,
    trendMetrics.rangePosition <= 15 && trendMetrics.sampleSize >= 10 ? { color: 'green', icon: <ThunderboltOutlined />, text: `价格接近样本区间低位 ${trendMetrics.rangePosition.toFixed(0)}%，更适合结合基本面做低位观察。` } : null,
    trendMetrics.volumeRatio >= 1.8 ? { color: 'orange', icon: <ThunderboltOutlined />, text: `近 5 根成交量约为前期均量 ${trendMetrics.volumeRatio.toFixed(2)} 倍，短线资金活跃度上升。` } : null,
    trendMetrics.maxDrawdown <= -8 ? { color: 'red', icon: <WarningOutlined />, text: `样本内最大回撤 ${trendMetrics.maxDrawdown.toFixed(2)}%，波动风险偏高。` } : null,
    detail?.is_stale ? { color: 'warning', icon: <WarningOutlined />, text: '详细行情来自缓存，建议点击强制刷新后再做判断。' } : null,
    !newsItems.length ? { color: 'warning', icon: <WarningOutlined />, text: '当前未拉取到可追溯新闻，AI 分析时不要假设近期事件。' } : null,
    newsItems.length && recentNewsCount === 0 ? { color: 'warning', icon: <WarningOutlined />, text: '新闻源存在但近期新闻不足，短线事件驱动判断需要降权。' } : null,
    industryLabel ? { color: 'green', icon: <ThunderboltOutlined />, text: `已识别所属行业/板块：${industryLabel}${conceptList.length ? `，概念标签 ${conceptList.slice(0, 3).join('、')}` : ''}。` } : null,
  ].filter(Boolean) as Array<{ color: string; icon: React.ReactNode; text: string }>;

  const summaryCards = [
    { title: '成交额', value: formatLargeNumber(selectedLiveQuote?.turnover || profile?.turnover || detail?.turnover || latestPoint?.turnover), color: '#1677ff', field: 'turnover' },
    { title: '换手率', value: formatPercent(selectedLiveQuote?.turnover_rate || profile?.turnover_rate || detail?.turnover_rate), color: '#722ed1', field: 'turnover_rate' },
    { title: '振幅', value: formatPercent(selectedLiveQuote?.amplitude || profile?.amplitude || detail?.amplitude), color: '#fa8c16', field: 'amplitude' },
    { title: '量比', value: toNumber(selectedLiveQuote?.volume_ratio || profile?.volume_ratio || detail?.volume_ratio).toFixed(2), color: '#13c2c2', field: 'volume_ratio' },
    { title: '总市值', value: formatLargeNumber(selectedLiveQuote?.total_market_cap || profile?.total_market_cap || detail?.total_market_cap), color: '#2f54eb', field: 'total_market_cap' },
    { title: '关注次数', value: String(selectedCandidate?.trade_count ?? 0), color: '#52c41a' },
  ];

  const onSelectSymbol = (symbol: string) => void load(symbol, false, klinePeriod);

  useEffect(() => {
    if (!selectedSymbol) {
      return;
    }

    const refreshViewedSymbol = async () => {
      if (document.visibilityState !== 'visible' || refreshInFlightRef.current) {
        return;
      }
      refreshInFlightRef.current = true;

      const symbolsToRefresh = Array.from(new Set(candidateList.map((item) => item.symbol).filter(Boolean)));

      try {
        const [historyRes, detailRes, klineRes, newsRes, watchDetailResults] = await Promise.all([
          Promise.allSettled([
            marketApi.getSnapshotHistory({ symbol: selectedSymbol, limit: 30 }),
            marketApi.getStockDetail(selectedSymbol, { refresh: true }),
            marketApi.getStockKlines(selectedSymbol, { period: klinePeriod, adjust: 'qfq', limit: 60, refresh: true }),
            marketApi.getStockNews(selectedSymbol, { limit: 10 }),
          ]),
          Promise.allSettled(symbolsToRefresh.map((symbol) => marketApi.getStockDetail(symbol, { refresh: true }))),
        ]).then(([mainResults, watchResults]) => [mainResults[0], mainResults[1], mainResults[2], mainResults[3], watchResults] as const);

        if (historyRes.status === 'fulfilled') {
          setHistory([...historyRes.value].reverse());
        }

        if (detailRes.status === 'fulfilled') {
          setDetail(detailRes.value);
          setDetailError('');
          commitLiveQuotes([buildQuoteSnapshotFromDetail(detailRes.value)]);
          try {
            const nextProfile = await marketApi.getStockProfile(selectedSymbol);
            setProfile(nextProfile);
            commitLiveQuotes([buildQuoteSnapshotFromProfile(nextProfile)]);
          } catch {
            // Keep the last profile snapshot when the lightweight refresh cannot update it.
          }
        } else {
          setDetailError(detailRes.reason?.message ?? detailRes.reason?.data?.message ?? '详细行情暂时不可用');
        }

        if (klineRes.status === 'fulfilled') {
          setKline(klineRes.value);
          setKlineError('');
        } else {
          setKlineError(klineRes.reason?.message ?? klineRes.reason?.data?.message ?? 'K线数据暂时不可用');
        }

        if (newsRes.status === 'fulfilled') {
          setStockNews(newsRes.value);
        }

        await loadIntraday(selectedSymbol, true);

        const quoteEntries = watchDetailResults
          .filter((result): result is PromiseFulfilledResult<MarketStockDetailResponse> => result.status === 'fulfilled')
          .map((result) => buildQuoteSnapshotFromDetail(result.value));
        commitLiveQuotes(quoteEntries);
      } finally {
        refreshInFlightRef.current = false;
      }
    };

    const timer = window.setInterval(() => {
      void refreshViewedSymbol();
    }, ACTIVE_STOCK_REFRESH_MS);

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        void refreshViewedSymbol();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [candidateList, klinePeriod, selectedSymbol]);

  const addFavoriteSymbol = async (value: string) => {
    const symbol = normalizeSymbolInput(value);
    if (!symbol) {
      message.warning('请输入股票代码');
      return;
    }
    if (!/^[A-Z0-9.]{2,16}$/.test(symbol)) {
      message.warning('股票代码格式不正确');
      return;
    }
    let assetName = symbol;
    let market = 'cn_stock';
    try {
      const stockDetail = await marketApi.getStockDetail(symbol);
      assetName = stockDetail.name || symbol;
      market = stockDetail.market || market;
    } catch {
      message.info('已加入自选，暂时未获取到名称，选择后仍会尝试拉取行情');
    }
    const nextFavorites = [{ symbol, asset_name: assetName, market }, ...favoriteSymbols.filter((item) => item.symbol !== symbol)];
    setFavoriteSymbols(nextFavorites);
    setFavoriteQuery('');
    writeFavoriteSymbols(nextFavorites);
    message.success(`已加入自选：${assetName}`);
    void load(symbol, false, klinePeriod);
  };

  const removeFavoriteSymbol = (symbol: string) => {
    const nextFavorites = favoriteSymbols.filter((item) => item.symbol !== symbol);
    setFavoriteSymbols(nextFavorites);
    writeFavoriteSymbols(nextFavorites);
    message.success(`已移除自选：${symbol}`);
    if (selectedSymbol === symbol) {
      const nextSymbol = candidateList.find((item) => item.symbol !== symbol)?.symbol || nextFavorites[0]?.symbol || '';
      if (nextSymbol) void load(nextSymbol, false, klinePeriod);
    }
  };

  const onSelectPeriod = (period: KlinePeriod) => {
    setKlinePeriod(period);
    if (selectedSymbol) void loadKline(selectedSymbol, period);
  };

  const loadIntraday = async (symbol: string, refresh = false) => {
    if (!symbol) {
      setIntradayKline(null);
      setIntradayError('');
      return;
    }
    setIntradayLoading(true);
    setIntradayError('');
    try {
      const oneMinute = await marketApi.getStockKlines(symbol, { period: '1m', adjust: 'qfq', limit: 120, refresh });
      setIntradayKline(oneMinute);
      setIntradayError('');
      return;
    } catch (oneMinuteErr: any) {
      try {
        const fallback = await marketApi.getStockKlines(symbol, { period: '5m', adjust: 'qfq', limit: 120, refresh });
        setIntradayKline(fallback);
        setIntradayError(oneMinuteErr?.message ?? oneMinuteErr?.data?.message ?? '');
      } catch {
        setIntradayKline(null);
        setIntradayError('盘中走势暂时不可用');
      }
    } finally {
      setIntradayLoading(false);
    }
  };

  const flashCardStyle = (flash?: QuoteFlashState): React.CSSProperties | undefined => {
    if (!flash) return undefined;
    return {
      position: 'relative',
      overflow: 'hidden',
      border: `1px solid ${getFlashAccentColor(flash)}`,
      boxShadow: `0 0 0 1px ${getFlashAccentColor(flash)}22, 0 10px 26px ${getFlashAccentColor(flash)}1f`,
      transition: 'all 0.28s ease',
    };
  };

  const flashValueStyle = (baseColor: string, field: string): React.CSSProperties => {
    const active = selectedFlash?.fields.includes(field);
    return {
      color: active ? getFlashAccentColor(selectedFlash) : baseColor,
      fontSize: 24,
      transition: 'color 0.28s ease, transform 0.28s ease, text-shadow 0.28s ease',
      transform: active ? 'scale(1.05)' : 'scale(1)',
      textShadow: active ? `0 0 18px ${getFlashAccentColor(selectedFlash)}55` : 'none',
    };
  };

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')} style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>返回首页</Button>

      <Card bordered={false} style={{ marginBottom: 24, borderRadius: 20, background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)', boxShadow: '0 18px 40px rgba(22,119,255,0.18)', ...(selectedFlash ? { outline: `1px solid ${getFlashAccentColor(selectedFlash)}`, boxShadow: `0 18px 40px rgba(22,119,255,0.18), 0 0 0 1px ${getFlashAccentColor(selectedFlash)}44, 0 0 28px ${getFlashAccentColor(selectedFlash)}2a` } : {}) }} bodyStyle={{ padding: 28 }}>
        <Row gutter={[20, 20]} align="middle">
          <Col span={24} xl={15}>
            <Space size={10} wrap style={{ marginBottom: 14 }}>
              <Tag color="processing">个股详情</Tag>
              <Tag color="blue">{marketLabel(profile?.market || detail?.market || latestPoint?.market || selectedCandidate?.market)}</Tag>
              <Tag color={selectedCandidate?.is_held ? 'success' : 'default'}>{selectedCandidate?.is_held ? '当前持仓' : '关注列表'}</Tag>
              {selectedCandidate?.sources.some((source) => source.type === 'favorite') ? <Tag color="gold">自选</Tag> : null}
              {detail?.is_stale ? <Tag color="warning">缓存数据</Tag> : <Tag color="success">数据较新</Tag>}
              {detail?.refresh_triggered || kline?.refresh_triggered ? <Tag color="gold">已触发刷新</Tag> : null}
            </Space>
            <Space direction="vertical" size={6} style={{ width: '100%' }}>
              <Space align="end" size={14} wrap>
                <Title level={2} style={{ margin: 0, color: '#fff' }}>{profile?.name || detail?.name || selectedCandidate?.asset_name || '证券详情'}</Title>
                <Text style={{ color: 'rgba(255,255,255,0.78)', fontSize: 16 }}>{selectedSymbol || '—'}</Text>
              </Space>
              <Space align="end" size={16} wrap>
                <Text style={{ color: '#fff', fontSize: 40, fontWeight: 700, lineHeight: 1, transition: 'transform 0.28s ease, text-shadow 0.28s ease', transform: selectedFlash?.fields.includes('last_price') ? 'scale(1.04)' : 'scale(1)', textShadow: selectedFlash?.fields.includes('last_price') ? `0 0 18px ${getFlashAccentColor(selectedFlash)}88` : 'none' }}>{formatPrice(latestPrice)}</Text>
                <Text style={{ color: getChangeColor(latestChangePercent), fontSize: 18, fontWeight: 600, lineHeight: 1.3 }}>{formatChangeText(latestChangeAmount, latestChangePercent)}</Text>
                {selectedFlash ? (
                  <Tag color={selectedFlash.direction === 'up' ? 'error' : selectedFlash.direction === 'down' ? 'success' : 'processing'} style={{ marginInlineStart: 0, borderRadius: 999, paddingInline: 10, fontWeight: 600 }}>
                    {selectedFlash.direction === 'up' ? '↑' : selectedFlash.direction === 'down' ? '↓' : '•'} {formatPriceDelta(selectedFlash.priceDiff)} / {formatSignedPercent(selectedFlash.changePercentDiff)}
                  </Tag>
                ) : null}
              </Space>
              {metaSummary ? <Text style={{ color: 'rgba(255,255,255,0.82)' }}>{metaSummary}</Text> : null}
              <Text style={{ color: 'rgba(255,255,255,0.72)' }}>更新时间 {profile?.fetched_at || detail?.fetched_at || latestPoint?.snapshot_time || '—'}</Text>
              <Button type="primary" icon={<RobotOutlined />} onClick={() => navigate(`/app/chat?kind=stock&symbol=${encodeURIComponent(selectedSymbol)}`)} disabled={!selectedSymbol} style={{ marginTop: 6, borderRadius: 10, boxShadow: '0 8px 24px rgba(22,119,255,0.28)' }}>AI 分析对话</Button>
            </Space>
          </Col>
          <Col span={24} xl={9}>
            <Row gutter={[12, 12]}>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>成交额</span>} value={formatLargeNumber(profile?.turnover || detail?.turnover || latestPoint?.turnover)} valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>换手率</span>} value={toNumber(profile?.turnover_rate || detail?.turnover_rate)} precision={2} suffix="%" valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>总市值</span>} value={formatLargeNumber(profile?.total_market_cap || detail?.total_market_cap)} valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>近期新闻</span>} value={`${recentNewsCount}/${newsItems.length}`} valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
            </Row>
          </Col>
        </Row>
      </Card>

      <Spin spinning={loading}>
        {error ? (
          <Card bordered={false} style={cardStyle}><Alert type="error" showIcon message={error} /></Card>
        ) : !candidateList.length ? (
          <Card bordered={false} style={cardStyle}>
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <AutoComplete
                value={favoriteQuery}
                options={favoriteOptions}
                onChange={setFavoriteQuery}
                onSelect={(value) => void addFavoriteSymbol(value)}
                style={{ width: '100%' }}
              >
                <Search enterButton="加入自选" placeholder="输入股票代码 / 名称，例如 000858 或 五粮液" onSearch={(value) => void addFavoriteSymbol(value)} />
              </AutoComplete>
              <Empty description="暂无候选证券，请先加入自选，或导入交易记录 / 生成持仓" />
            </Space>
          </Card>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={24} lg={7}>
              <Card bordered={false} style={cardStyle} title={<span><StockOutlined style={{ color: '#1677ff', marginRight: 8 }} />关注证券</span>} extra={<Text type="secondary">持仓 / 自选</Text>}>
                <AutoComplete
                  value={favoriteQuery}
                  options={favoriteOptions}
                  onChange={setFavoriteQuery}
                  onSelect={(value) => void addFavoriteSymbol(value)}
                  style={{ width: '100%', marginBottom: 14 }}
                >
                  <Search enterButton="加入自选" placeholder="输入股票代码 / 名称，例如 000858 或 五粮液" onSearch={(value) => void addFavoriteSymbol(value)} />
                </AutoComplete>
                <List dataSource={candidateList} renderItem={(item) => {
                  const isFavorite = item.sources.some((source) => source.type === 'favorite');
                  const liveQuote = liveQuotes[item.symbol];
                  const flash = quoteFlashes[item.symbol];
                  const itemChangePercent = liveQuote?.change_percent || item.change_percent;
                  const itemPrice = liveQuote?.last_price || item.last_price;
                  return (
                    <List.Item style={{ paddingInline: 0 }}>
                      <div onClick={() => onSelectSymbol(item.symbol)} style={{ width: '100%', cursor: 'pointer', padding: 12, borderRadius: 12, background: item.symbol === selectedSymbol ? '#e6f4ff' : flash ? `${getFlashAccentColor(flash)}10` : '#fafafa', border: item.symbol === selectedSymbol ? '1px solid #91caff' : flash ? `1px solid ${getFlashAccentColor(flash)}66` : '1px solid #f0f0f0', boxShadow: flash ? `0 8px 22px ${getFlashAccentColor(flash)}18` : 'none', transition: 'all 0.28s ease' }}>
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
                            <Space wrap>
                              <Text strong>{item.asset_name || item.symbol}</Text>
                              {item.is_held ? <Tag color="success">已持仓</Tag> : null}
                              {isFavorite ? <Tag icon={<StarOutlined />} color="gold">自选</Tag> : null}
                              {!item.is_held && !isFavorite ? <Tag>历史关注</Tag> : null}
                            </Space>
                            {isFavorite ? <Button danger type="text" size="small" icon={<DeleteOutlined />} onClick={(event) => { event.stopPropagation(); removeFavoriteSymbol(item.symbol); }} /> : null}
                          </Space>
                          <Text type="secondary">{item.symbol}</Text>
                          <Space split={<Text type="secondary">|</Text>} size={6} wrap>
                            <Text type={getChangeColor(itemChangePercent) === '#ff4d4f' ? 'danger' : getChangeColor(itemChangePercent) === '#52c41a' ? 'success' : undefined} style={{ fontWeight: flash?.fields.includes('change_percent') ? 700 : 500, transition: 'all 0.28s ease', textShadow: flash?.fields.includes('change_percent') ? `0 0 14px ${getFlashAccentColor(flash)}55` : 'none' }}>涨跌幅 {formatPercent(itemChangePercent)}</Text>
                            <Text style={{ color: '#595959', fontWeight: flash?.fields.includes('last_price') ? 700 : 500, transition: 'all 0.28s ease', textShadow: flash?.fields.includes('last_price') ? `0 0 14px ${getFlashAccentColor(flash)}55` : 'none' }}>现价 {formatPrice(itemPrice)}</Text>
                            {flash ? <Tag color={flash.direction === 'up' ? 'error' : flash.direction === 'down' ? 'success' : 'processing'} style={{ borderRadius: 999 }}>{
                              flash.direction === 'up' ? `+${flash.priceDiff.toFixed(2)}` : flash.direction === 'down' ? flash.priceDiff.toFixed(2) : '更新'
                            }</Tag> : null}
                            <Text type="secondary">关注 {item.trade_count ?? 0} 次</Text>
                          </Space>
                        </Space>
                      </div>
                    </List.Item>
                  );
                }} />
              </Card>
            </Col>

            <Col span={24} lg={17}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Row gutter={[16, 16]}>
                  {summaryCards.map((item) => <Col key={item.title} xs={12} xl={8}><Card bordered={false} style={{ ...cardStyle, ...(item.field ? flashCardStyle(selectedFlash?.fields.includes(item.field) ? selectedFlash : undefined) : undefined) }}><Statistic title={item.title} value={item.value} valueStyle={item.field ? flashValueStyle(item.color, item.field) : { color: item.color, fontSize: 24 }} prefix={item.title === '关注次数' ? <RiseOutlined /> : undefined} /></Card></Col>)}
                </Row>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={10}>
                    <Card bordered={false} style={cardStyle} title="证券介绍" extra={<Text type="secondary">来源 {profile?.company_profile?.source || profile?.source || detail?.source || '—'}</Text>}>
                      <Spin spinning={profileLoading}>
                        <Space direction="vertical" size={12} style={{ width: '100%' }}>
                          <Space wrap>
                          {profile?.company_profile?.market_label ? <Tag color="geekblue">{profile.company_profile.market_label}</Tag> : null}
                          {profile?.company_profile?.industry_label ? <Tag color="blue">{profile.company_profile.industry_label}</Tag> : industryLabel ? (
                            <Tag
                              color="blue"
                              style={industryBoard ? { cursor: 'pointer' } : undefined}
                              onClick={industryBoard ? () => navigate(`/app/board?type=${encodeURIComponent(industryBoard.board_type)}&code=${encodeURIComponent(industryBoard.code)}`) : undefined}
                            >
                              行业 {industryLabel}
                            </Tag>
                          ) : null}
                          {regionLabel ? <Tag color="cyan">地区 {regionLabel}</Tag> : null}
                          <Tag color={profile?.is_stale || detail?.is_stale ? 'warning' : 'success'}>{profile?.is_stale || detail?.is_stale ? '缓存行情' : '行情较新'}</Tag>
                        </Space>
                        <Text strong>{profile?.company_profile?.company_name || profile?.name || detail?.name || selectedCandidate?.asset_name || selectedSymbol}</Text>
                        <Text style={{ lineHeight: 1.8 }}>
                          {profile?.company_profile?.introduction || profile?.company_profile?.business || profile?.description || `${profile?.name || detail?.name || selectedCandidate?.asset_name || selectedSymbol} 当前暂无完整公司简介，页面已优先展示可信行情、板块标签、K线趋势和可追溯新闻。`}
                        </Text>
                        {profile?.company_profile?.business ? <Text type="secondary">主营业务：{profile.company_profile.business}</Text> : null}
                        <Row gutter={[12, 12]}>
                          <Col span={12}><Statistic title="成立日期" value={profile?.company_profile?.founded_at || '—'} /></Col>
                          <Col span={12}><Statistic title="上市日期" value={profile?.company_profile?.listed_at || '—'} /></Col>
                          <Col span={12}><Statistic title="流通市值" value={formatLargeNumber(profile?.float_market_cap || detail?.float_market_cap)} /></Col>
                          <Col span={12}><Statistic title="振幅" value={toNumber(profile?.amplitude || detail?.amplitude)} precision={2} suffix="%" /></Col>
                          <Col span={12}><Statistic title="涨停价" value={toNumber(profile?.limit_up || detail?.limit_up)} precision={2} prefix="¥" /></Col>
                          <Col span={12}><Statistic title="跌停价" value={toNumber(profile?.limit_down || detail?.limit_down)} precision={2} prefix="¥" /></Col>
                        </Row>
                        {profile?.company_profile?.website ? <Text type="secondary">官网：{profile.company_profile.website}</Text> : null}
                        </Space>
                      </Spin>
                    </Card>
                  </Col>
                  <Col span={24} xl={14}>
                    <Card bordered={false} style={cardStyle} title="板块与概念" extra={<Text type="secondary">用于辅助判断，不替代财报分析</Text>}>
                      <Space direction="vertical" size={14} style={{ width: '100%' }}>
                        <Alert
                          type={industryLabel ? 'info' : 'warning'}
                          showIcon
                          message={industryLabel ? `所属行业：${industryLabel}` : '暂未识别可靠行业信息'}
                          description={conceptList.length ? `概念覆盖：${conceptList.slice(0, 6).join('、')}` : '当前数据源暂未返回概念标签。'}
                        />
                        <Space wrap size={[8, 8]}>
                          {visibleConcepts.length ? visibleConcepts.map((concept) => {
                            const conceptBoard = boardByName.get(concept);
                            return (
                              <Tag
                                key={concept}
                                color="processing"
                                style={conceptBoard ? { cursor: 'pointer' } : undefined}
                                onClick={conceptBoard ? () => navigate(`/app/board?type=${encodeURIComponent(conceptBoard.board_type)}&code=${encodeURIComponent(conceptBoard.code)}`) : undefined}
                              >
                                {concept}
                              </Tag>
                            );
                          }) : <Text type="secondary">暂无可靠概念标签</Text>}
                          {conceptList.length > 8 ? <Button type="link" size="small" style={{ paddingInline: 0 }} onClick={() => setShowAllConcepts((value) => !value)}>{showAllConcepts ? '收起' : `展开全部 (${conceptList.length})`}</Button> : null}
                        </Space>
                        <Space wrap size={[8, 8]}>
                          <Text type="secondary">所属板块</Text>
                          {visibleBoards.length ? visibleBoards.map((board) => (
                            <Tag
                              key={`${board.board_type}-${board.code}`}
                              color={board.board_type === 'industry' ? 'blue' : 'purple'}
                              style={{ cursor: 'pointer' }}
                              onClick={() => navigate(`/app/board?type=${encodeURIComponent(board.board_type)}&code=${encodeURIComponent(board.code)}`)}
                            >
                              {board.name}
                            </Tag>
                          )) : <Text type="secondary">暂无可映射板块</Text>}
                        </Space>
                        <Row gutter={[12, 12]}>
                          <Col span={8}><Statistic title="成交额" value={formatLargeNumber(profile?.turnover || detail?.turnover)} /></Col>
                          <Col span={8}><Statistic title="换手率" value={toNumber(profile?.turnover_rate || detail?.turnover_rate)} precision={2} suffix="%" /></Col>
                          <Col span={8}><Statistic title="量比" value={toNumber(profile?.volume_ratio || detail?.volume_ratio)} precision={2} /></Col>
                        </Row>
                      </Space>
                    </Card>
                  </Col>
                </Row>

                <Card bordered={false} style={cardStyle} title={<span><FundProjectionScreenOutlined style={{ color: '#1677ff', marginRight: 8 }} />趋势体检</span>} extra={<Text type="secondary">基于当前 {getKlinePeriodLabel(klinePeriod)} 样本计算</Text>}>
                  <Row gutter={[16, 16]}>
                    <Col xs={24} md={12} xl={6}>
                      <Statistic title="区间位置" value={Math.max(0, Math.min(100, trendMetrics.rangePosition))} precision={0} suffix="%" valueStyle={{ color: rangePositionColor }} />
                      <Progress percent={Math.max(0, Math.min(100, Number(trendMetrics.rangePosition.toFixed(0))))} strokeColor={rangePositionColor} showInfo={false} />
                      <Text type="secondary">低点 {formatPrice(String(trendMetrics.rangeLow))} / 高点 {formatPrice(String(trendMetrics.rangeHigh))}</Text>
                    </Col>
                    <Col xs={24} md={12} xl={6}>
                      <Statistic title="短线涨跌" value={formatSignedPercent(trendMetrics.dayReturn)} valueStyle={{ color: getChangeColor(String(trendMetrics.dayReturn)) }} />
                      <Text type="secondary">最近两根 K 线收盘价变化</Text>
                    </Col>
                    <Col xs={24} md={12} xl={6}>
                      <Statistic title={trendBiasLabel} value={formatSignedPercent(trendMetrics.maBias)} valueStyle={{ color: latestAboveMa20 ? '#ff4d4f' : '#52c41a' }} />
                      <Text type="secondary">MA5 {formatPrice(String(trendMetrics.ma5))} / MA20 {formatPrice(String(trendMetrics.ma20))}</Text>
                    </Col>
                    <Col xs={24} md={12} xl={6}>
                      <Statistic title="波动 / 量能" value={`${trendMetrics.volatility.toFixed(2)}% / ${trendMetrics.volumeRatio.toFixed(2)}x`} valueStyle={{ color: trendMetrics.volumeRatio >= 1.5 ? '#fa8c16' : '#1677ff' }} />
                      <Text type="secondary">近 20 根收益波动，近 5 根量能相对前期</Text>
                    </Col>
                  </Row>
                </Card>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={10}>
                    <Card bordered={false} style={cardStyle} title="重点提示">
                      {alertItems.length ? (
                        <Space direction="vertical" size={10} style={{ width: '100%' }}>
                          {alertItems.map((item) => (
                            <Alert key={item.text} type={item.color === 'red' ? 'error' : item.color === 'green' ? 'success' : 'warning'} showIcon icon={item.icon} message={item.text} />
                          ))}
                        </Space>
                      ) : (
                        <Alert type="info" showIcon message="当前样本没有明显异常信号，建议结合新闻、财报和持仓成本继续判断。" />
                      )}
                    </Card>
                  </Col>
                  <Col span={24} xl={14}>
                    <Card bordered={false} style={cardStyle} title="量价关系">
                      {kline?.items?.length ? <ReactECharts option={buildVolumeTrendOption(kline, klinePeriod)} style={{ height: 260 }} /> : <Empty description="暂无成交量数据" />}
                    </Card>
                  </Col>
                </Row>

                <Card bordered={false} style={cardStyle} title={<span><BarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />{chartTitle}</span>} extra={<Space wrap size={12}><Segmented options={[{ label: 'K线', value: 'kline' }, { label: '盘中走势', value: 'intraday' }]} value={chartView} onChange={(value) => setChartView(value as ChartView)} />{chartView === 'kline' ? <Segmented options={klinePeriodOptions} value={klinePeriod} onChange={(value) => onSelectPeriod(value as KlinePeriod)} /> : intradayKline?.period === '5m' ? <Tag color="gold">1分不可用，已降级到5分</Tag> : intradayKline?.period === '1m' ? <Tag color="green">1分级别</Tag> : <Tag>等待盘中数据</Tag>}<Button icon={<ReloadOutlined />} onClick={() => void load(selectedSymbol, true, klinePeriod)} loading={detailLoading || klineLoading || intradayLoading}>强制刷新</Button></Space>}>
                  {chartView === 'kline'
                    ? <Spin spinning={klineLoading}>{klineError ? <Alert type="warning" showIcon message={klineError} style={{ marginBottom: 16 }} /> : null}{kline?.items?.length ? <ReactECharts option={buildKlineOption(kline, klinePeriod)} style={{ height: 380 }} /> : <Empty description="暂无该标的的 K 线数据" />}</Spin>
                    : <Spin spinning={intradayLoading}>
                        {intradayError ? <Alert type="warning" showIcon message={intradayError} style={{ marginBottom: 16 }} /> : null}
                        {intradayKline?.items?.length
                          ? <ReactECharts option={buildIntradayOption(intradayKline, intradayPeriod)} style={{ height: 380 }} />
                          : <Empty description="暂无该标的的盘中走势数据" />}
                      </Spin>}
                </Card>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={14}>
                    <Card bordered={false} style={cardStyle} title="均线动量">
                      {kline?.items?.length ? <ReactECharts option={buildPriceMomentumOption(kline, klinePeriod)} style={{ height: 300 }} /> : <Empty description="暂无均线数据" />}
                    </Card>
                  </Col>
                  <Col span={24} xl={10}>
                    <Card bordered={false} style={cardStyle} title="涨跌分布">
                      {kline?.items?.length ? <ReactECharts option={buildReturnDistributionOption(kline)} style={{ height: 300 }} /> : <Empty description="暂无涨跌分布数据" />}
                    </Card>
                  </Col>
                </Row>

                <Card
                  bordered={false}
                  style={cardStyle}
                  title="新闻与公告"
                  extra={<Text type="secondary">覆盖 {stockNews?.coverage || '—'} · 最新 {newestNews?.published_at || '—'}</Text>}
                >
                  <Spin spinning={newsLoading}>
                  {newsItems.length ? (
                    <List
                      dataSource={newsItems}
                      renderItem={(item) => (
                        <List.Item style={{ paddingInline: 0 }}>
                          <Space direction="vertical" size={6} style={{ width: '100%' }}>
                            <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
                              <a href={item.url} target="_blank" rel="noreferrer" style={{ fontWeight: 600 }}>{item.title}</a>
                              <Space wrap>
                                <Tag color={item.is_recent ? 'green' : 'default'}>{item.is_recent ? '近期' : '较早'}</Tag>
                                <Tag>{item.source || item.provider || '来源未知'}</Tag>
                              </Space>
                            </Space>
                            {item.summary ? <Text type="secondary">{item.summary}</Text> : null}
                            <Text type="secondary">{item.published_at || '发布时间未知'}</Text>
                          </Space>
                        </List.Item>
                      )}
                    />
                  ) : (
                    <Empty description="暂无可追溯新闻；AI 分析时将不会编造资讯。" />
                  )}
                  </Spin>
                </Card>

                <Card bordered={false} style={cardStyle} title="详细行情" extra={<Text type="secondary">数据源 {detail?.source || latestPoint?.source || '—'}</Text>}>
                  {detailError ? <Alert type="warning" showIcon message={detailError} style={{ marginBottom: 16 }} /> : null}
                  {detail || latestPoint ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12} lg={6}><Statistic title="开盘价" value={toNumber(detail?.open_price ?? latestPoint?.open_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="最高价" value={toNumber(detail?.high_price ?? latestPoint?.high_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="最低价" value={toNumber(detail?.low_price ?? latestPoint?.low_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="昨收价" value={toNumber(detail?.prev_close ?? latestPoint?.prev_close)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="成交量" value={formatLargeNumber(detail?.volume ?? latestPoint?.volume)} /></Col>
                      <Col span={12} lg={6}><Statistic title="成交额" value={formatLargeNumber(detail?.turnover ?? latestPoint?.turnover)} /></Col>
                      <Col span={12} lg={6}><Statistic title="量比" value={toNumber(detail?.volume_ratio).toFixed(2)} /></Col>
                      <Col span={12} lg={6}><Statistic title="换手率" value={toNumber(detail?.turnover_rate)} precision={2} suffix="%" /></Col>
                      <Col span={12} lg={6}><Statistic title="振幅" value={toNumber(detail?.amplitude)} precision={2} suffix="%" /></Col>
                      <Col span={12} lg={6}><Statistic title="均价" value={toNumber(detail?.average_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="涨停价" value={toNumber(detail?.limit_up)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="跌停价" value={toNumber(detail?.limit_down)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="总市值" value={formatLargeNumber(detail?.total_market_cap)} /></Col>
                      <Col span={12} lg={6}><Statistic title="流通市值" value={formatLargeNumber(detail?.float_market_cap)} /></Col>
                      <Col span={12} lg={6}><Statistic title="总股本" value={formatLargeNumber(detail?.total_shares)} /></Col>
                      <Col span={12} lg={6}><Statistic title="流通股本" value={formatLargeNumber(detail?.float_shares)} /></Col>
                      <Col span={24}><Space direction="vertical" size={10} style={{ width: '100%' }}><Space split={<Text type="secondary">|</Text>} wrap>{industryLabel ? <Text type="secondary">行业：{industryLabel}</Text> : null}{regionLabel ? <Text type="secondary">地区：{regionLabel}</Text> : null}<Text type="secondary">数据时间：{detail?.fetched_at || latestPoint?.snapshot_time || '—'}</Text></Space><Space wrap size={[8, 8]}><Text type="secondary">概念标签</Text>{visibleConcepts.length ? visibleConcepts.map((concept) => <Tag key={concept}>{concept}</Tag>) : <Text type="secondary">暂无可靠标签</Text>}{conceptList.length > 8 ? <Button type="link" size="small" style={{ paddingInline: 0 }} onClick={() => setShowAllConcepts((value) => !value)}>{showAllConcepts ? '收起' : `展开全部 (${conceptList.length})`}</Button> : null}</Space></Space></Col>
                    </Row>
                  ) : <Empty description="暂无详细行情数据" />}
                </Card>
              </Space>
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
}
