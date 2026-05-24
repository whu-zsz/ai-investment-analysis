import { useEffect, useMemo, useState } from 'react';
import { Alert, AutoComplete, Button, Card, Col, Empty, Input, List, Row, Segmented, Space, Spin, Statistic, Tag, Typography, message } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined, DeleteOutlined, ReloadOutlined, RiseOutlined, RobotOutlined, StarOutlined, StockOutlined } from '@ant-design/icons';
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
} from '../api/types';

const { Title, Text } = Typography;
const { Search } = Input;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

type ChartView = 'kline' | 'snapshot';
type KlinePeriod = 'day' | 'week' | 'month' | '5m' | '15m' | '60m';
type FavoriteSymbol = { symbol: string; asset_name: string; market: string };

const FAVORITE_SYMBOLS_KEY = 'marketTrendFavoriteSymbols';

const klinePeriodOptions: Array<{ label: string; value: KlinePeriod }> = [
  { label: '日线', value: 'day' },
  { label: '周线', value: 'week' },
  { label: '月线', value: 'month' },
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
    { paddingRatio: 0.02, minPaddingAbs: period === '5m' || period === '15m' || period === '60m' ? 0.02 : 0.1, roundMode: 'magnitude', splitNumber: 4 },
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

function buildSnapshotOption(history: MarketSnapshotResponse[]): EChartsOption {
  const axisRange = computeNumericAxisRange(history.map((item) => toNumber(item.last_price)), {
    minPaddingRatio: 0.06,
    maxPaddingRatio: 0.06,
    minPaddingAbs: 0.1,
    roundMode: 'magnitude',
    splitNumber: 4,
  });

  return {
    tooltip: { trigger: 'axis' },
    grid: { top: 24, left: 36, right: 20, bottom: 28, containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: history.map((item) => item.snapshot_time.slice(5, 16)), axisLine: { lineStyle: { color: '#d9d9d9' } }, axisLabel: { color: '#8c8c8c', fontSize: 11 } },
    yAxis: { type: 'value', min: axisRange?.min, max: axisRange?.max, splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }, axisLabel: { color: '#8c8c8c', fontSize: 11 } },
    series: [{ name: '最新价', type: 'line', smooth: false, showSymbol: false, data: history.map((item) => toNumber(item.last_price)), lineStyle: { width: 3, color: '#1677ff' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(22,119,255,0.25)' }, { offset: 1, color: 'rgba(22,119,255,0.03)' }] } } }],
  };
}

export default function MarketTrendPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [klineLoading, setKlineLoading] = useState(false);
  const [error, setError] = useState('');
  const [detailError, setDetailError] = useState('');
  const [klineError, setKlineError] = useState('');
  const [candidates, setCandidates] = useState<AnalysisCandidateResponse[]>([]);
  const [favoriteSymbols, setFavoriteSymbols] = useState<FavoriteSymbol[]>(() => readFavoriteSymbols());
  const [favoriteQuery, setFavoriteQuery] = useState('');
  const [marketSearchResults, setMarketSearchResults] = useState<MarketSnapshotResponse[]>([]);
  const [history, setHistory] = useState<MarketSnapshotResponse[]>([]);
  const [detail, setDetail] = useState<MarketStockDetailResponse | null>(null);
  const [kline, setKline] = useState<MarketStockKlineResponse | null>(null);
  const [selectedSymbol, setSelectedSymbol] = useState('');
  const [chartView, setChartView] = useState<ChartView>('kline');
  const [klinePeriod, setKlinePeriod] = useState<KlinePeriod>('day');
  const [showAllConcepts, setShowAllConcepts] = useState(false);

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

  const load = async (preferSymbol?: string, forceRefresh = false, period: KlinePeriod = klinePeriod) => {
    setLoading(true);
    setError('');
    setDetailError('');
    setKlineError('');
    try {
      const candidateRes = await analysisApi.getCandidates();
      const candidateList = candidateRes.candidates ?? [];
      const favorites = readFavoriteSymbols();
      setCandidates(candidateList);
      setFavoriteSymbols(favorites);
      const nextSymbol = preferSymbol || searchParams.get('symbol') || candidateRes.default_symbol || candidateList[0]?.symbol || favorites[0]?.symbol || '';
      setSelectedSymbol(nextSymbol);
      setShowAllConcepts(false);
      if (!nextSymbol) {
        setHistory([]);
        setDetail(null);
        setKline(null);
        return;
      }

      const [historyRes, detailRes, klineRes] = await Promise.allSettled([
        marketApi.getSnapshotHistory({ symbol: nextSymbol, limit: 30 }),
        marketApi.getStockDetail(nextSymbol, forceRefresh ? { refresh: true } : undefined),
        marketApi.getStockKlines(nextSymbol, { period, adjust: 'qfq', limit: 60, refresh: forceRefresh }),
      ]);

      setHistory(historyRes.status === 'fulfilled' ? [...historyRes.value].reverse() : []);
      if (detailRes.status === 'fulfilled') {
        setDetail(detailRes.value);
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
      setSearchParams({ symbol: nextSymbol });
    } catch (err: any) {
      setCandidates([]);
      setHistory([]);
      setDetail(null);
      setKline(null);
      setError(err?.message ?? err?.data?.message ?? '市场趋势加载失败');
    } finally {
      setLoading(false);
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
  const latestPoint = history[history.length - 1] ?? null;
  const latestPrice = detail?.last_price ?? latestPoint?.last_price;
  const latestChangePercent = detail?.change_percent ?? latestPoint?.change_percent;
  const latestChangeAmount = detail?.change_amount ?? latestPoint?.change_amount;
  const industryLabel = normalizeDetailLabel(detail?.industry);
  const regionLabel = normalizeDetailLabel(detail?.region);
  const metaSummary = [industryLabel, regionLabel].filter(Boolean).join(' · ');
  const conceptList = normalizeConcepts(detail?.concepts);
  const visibleConcepts = showAllConcepts ? conceptList : conceptList.slice(0, 8);
  const chartTitle = chartView === 'kline' ? `${getKlinePeriodLabel(klinePeriod)} K 线` : '快照走势';

  const summaryCards = [
    { title: '成交额', value: formatLargeNumber(detail?.turnover ?? latestPoint?.turnover), color: '#1677ff' },
    { title: '换手率', value: formatPercent(detail?.turnover_rate), color: '#722ed1' },
    { title: '振幅', value: formatPercent(detail?.amplitude), color: '#fa8c16' },
    { title: '量比', value: toNumber(detail?.volume_ratio).toFixed(2), color: '#13c2c2' },
    { title: '总市值', value: formatLargeNumber(detail?.total_market_cap), color: '#2f54eb' },
    { title: '关注次数', value: String(selectedCandidate?.trade_count ?? 0), color: '#52c41a' },
  ];

  const onSelectSymbol = (symbol: string) => void load(symbol, false, klinePeriod);

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

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')} style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>返回首页</Button>

      <Card bordered={false} style={{ marginBottom: 24, borderRadius: 20, background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)', boxShadow: '0 18px 40px rgba(22,119,255,0.18)' }} bodyStyle={{ padding: 28 }}>
        <Row gutter={[20, 20]} align="middle">
          <Col span={24} xl={15}>
            <Space size={10} wrap style={{ marginBottom: 14 }}>
              <Tag color="processing">个股详情</Tag>
              <Tag color="blue">{marketLabel(detail?.market || latestPoint?.market || selectedCandidate?.market)}</Tag>
              <Tag color={selectedCandidate?.is_held ? 'success' : 'default'}>{selectedCandidate?.is_held ? '当前持仓' : '关注标的'}</Tag>
              {selectedCandidate?.sources.some((source) => source.type === 'favorite') ? <Tag color="gold">自选</Tag> : null}
              {detail?.is_stale ? <Tag color="warning">缓存数据</Tag> : <Tag color="success">数据较新</Tag>}
              {detail?.refresh_triggered || kline?.refresh_triggered ? <Tag color="gold">已触发刷新</Tag> : null}
            </Space>
            <Space direction="vertical" size={6} style={{ width: '100%' }}>
              <Space align="end" size={14} wrap>
                <Title level={2} style={{ margin: 0, color: '#fff' }}>{detail?.name || selectedCandidate?.asset_name || '标的详情'}</Title>
                <Text style={{ color: 'rgba(255,255,255,0.78)', fontSize: 16 }}>{selectedSymbol || '—'}</Text>
              </Space>
              <Space align="end" size={16} wrap>
                <Text style={{ color: '#fff', fontSize: 40, fontWeight: 700, lineHeight: 1 }}>{formatPrice(latestPrice)}</Text>
                <Text style={{ color: getChangeColor(latestChangePercent), fontSize: 18, fontWeight: 600, lineHeight: 1.3 }}>{formatChangeText(latestChangeAmount, latestChangePercent)}</Text>
              </Space>
              {metaSummary ? <Text style={{ color: 'rgba(255,255,255,0.82)' }}>{metaSummary}</Text> : null}
              <Text style={{ color: 'rgba(255,255,255,0.72)' }}>更新时间 {detail?.fetched_at || latestPoint?.snapshot_time || '—'}</Text>
              <Button type="primary" icon={<RobotOutlined />} onClick={() => navigate(`/app/stock-chat?symbol=${encodeURIComponent(selectedSymbol)}`)} disabled={!selectedSymbol} style={{ marginTop: 6, borderRadius: 10, boxShadow: '0 8px 24px rgba(22,119,255,0.28)' }}>AI 分析对话</Button>
            </Space>
          </Col>
          <Col span={24} xl={9}>
            <Row gutter={[12, 12]}>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>今开</span>} value={toNumber(detail?.open_price ?? latestPoint?.open_price)} precision={2} prefix="¥" valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>昨收</span>} value={toNumber(detail?.prev_close ?? latestPoint?.prev_close)} precision={2} prefix="¥" valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>最高</span>} value={toNumber(detail?.high_price ?? latestPoint?.high_price)} precision={2} prefix="¥" valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
              <Col span={12}><Card bordered={false} bodyStyle={{ padding: 16 }} style={{ borderRadius: 14, background: 'rgba(255,255,255,0.14)' }}><Statistic title={<span style={{ color: 'rgba(255,255,255,0.75)' }}>最低</span>} value={toNumber(detail?.low_price ?? latestPoint?.low_price)} precision={2} prefix="¥" valueStyle={{ color: '#fff', fontSize: 20 }} /></Card></Col>
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
              <Empty description="暂无候选标的，请先加入自选，或导入交易记录 / 生成持仓" />
            </Space>
          </Card>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={24} lg={7}>
              <Card bordered={false} style={cardStyle} title={<span><StockOutlined style={{ color: '#1677ff', marginRight: 8 }} />关注标的</span>} extra={<Text type="secondary">持仓 / 自选</Text>}>
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
                  return (
                    <List.Item style={{ paddingInline: 0 }}>
                      <div onClick={() => onSelectSymbol(item.symbol)} style={{ width: '100%', cursor: 'pointer', padding: 12, borderRadius: 12, background: item.symbol === selectedSymbol ? '#e6f4ff' : '#fafafa', border: item.symbol === selectedSymbol ? '1px solid #91caff' : '1px solid #f0f0f0' }}>
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
                            <Text type={getChangeColor(item.change_percent) === '#ff4d4f' ? 'danger' : getChangeColor(item.change_percent) === '#52c41a' ? 'success' : undefined}>涨跌幅 {formatPercent(item.change_percent)}</Text>
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
                  {summaryCards.map((item) => <Col key={item.title} xs={12} xl={8}><Card bordered={false} style={cardStyle}><Statistic title={item.title} value={item.value} valueStyle={{ color: item.color, fontSize: 24 }} prefix={item.title === '关注次数' ? <RiseOutlined /> : undefined} /></Card></Col>)}
                </Row>

                <Card bordered={false} style={cardStyle} title={<span><BarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />{chartTitle}</span>} extra={<Space wrap size={12}><Segmented options={[{ label: 'K线', value: 'kline' }, { label: '快照走势', value: 'snapshot' }]} value={chartView} onChange={(value) => setChartView(value as ChartView)} />{chartView === 'kline' ? <Segmented options={klinePeriodOptions} value={klinePeriod} onChange={(value) => onSelectPeriod(value as KlinePeriod)} /> : null}<Button icon={<ReloadOutlined />} onClick={() => void load(selectedSymbol, true, klinePeriod)} loading={loading || klineLoading}>强制刷新</Button></Space>}>
                  {chartView === 'kline' ? <Spin spinning={klineLoading}>{klineError ? <Alert type="warning" showIcon message={klineError} style={{ marginBottom: 16 }} /> : null}{kline?.items?.length ? <ReactECharts option={buildKlineOption(kline, klinePeriod)} style={{ height: 380 }} /> : <Empty description="暂无该标的的 K 线数据" />}</Spin> : history.length ? <ReactECharts option={buildSnapshotOption(history)} style={{ height: 380 }} /> : <Empty description="暂无该标的的历史快照" />}
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
