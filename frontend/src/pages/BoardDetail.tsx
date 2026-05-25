import { ArrowLeftOutlined, BarChartOutlined, ReloadOutlined, RiseOutlined, RobotOutlined, StockOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Col, Empty, List, Popover, Row, Space, Spin, Statistic, Table, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { ColumnsType } from 'antd/es/table';
import type { EChartsOption } from 'echarts';
import { marketApi } from '../api';
import type { BoardConstituentResponse, BoardNewsResponse, MarketBoardDetailResponse, MarketStockKlineResponse } from '../api/types';

const { Title, Text } = Typography;

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatPrice(value?: string) {
  const number = toNumber(value);
  return Number.isFinite(number) ? `¥${number.toFixed(2)}` : '—';
}

function formatSignedPercent(value?: string) {
  const number = toNumber(value);
  if (!Number.isFinite(number)) return '—';
  const sign = number > 0 ? '+' : '';
  return `${sign}${number.toFixed(2)}%`;
}

function formatCompactNumber(value?: string) {
  const number = toNumber(value);
  if (!Number.isFinite(number) || number === 0) return '—';
  if (Math.abs(number) >= 100000000) return `${(number / 100000000).toFixed(2)} 亿`;
  if (Math.abs(number) >= 10000) return `${(number / 10000).toFixed(2)} 万`;
  return number.toFixed(2);
}

function getChangeColor(value?: string) {
  const number = toNumber(value);
  if (number > 0) return '#ff4d4f';
  if (number < 0) return '#52c41a';
  return '#1677ff';
}

function buildAdvanceDeclineOption(detail: MarketBoardDetailResponse | null): EChartsOption {
  if (!detail) return {};
  return {
    tooltip: { trigger: 'item' },
    color: ['#ff4d4f', '#52c41a', '#91caff'],
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        label: { formatter: '{b}\n{c}' },
        data: [
          { name: '上涨', value: detail.board.rise_count },
          { name: '下跌', value: detail.board.fall_count },
          { name: '平盘', value: detail.board.flat_count },
        ],
      },
    ],
  };
}

function buildTopMoveOption(items: BoardConstituentResponse[], title: string, color: string): EChartsOption {
  return {
    title: { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 80, right: 20, top: 50, bottom: 24 },
    xAxis: { type: 'value', axisLabel: { formatter: '{value}%' } },
    yAxis: {
      type: 'category',
      data: items.map((item) => item.name || item.symbol),
      axisLabel: { width: 84, overflow: 'truncate' },
    },
    series: [
      {
        type: 'bar',
        data: items.map((item) => toNumber(item.change_percent)),
        itemStyle: { color, borderRadius: [0, 6, 6, 0] },
      },
    ],
  };
}

function buildTurnoverOption(items: BoardConstituentResponse[]): EChartsOption {
  return {
    title: { text: '成交额前排', left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 80, right: 20, top: 50, bottom: 24 },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: items.map((item) => item.name || item.symbol),
      axisLabel: { width: 84, overflow: 'truncate' },
    },
    series: [
      {
        type: 'bar',
        data: items.map((item) => toNumber(item.turnover)),
        itemStyle: { color: '#1677ff', borderRadius: [0, 6, 6, 0] },
      },
    ],
  };
}

function buildMiniTrendOption(name: string, kline: MarketStockKlineResponse | null): EChartsOption {
  const items = kline?.items ?? [];
  return {
    animation: false,
    title: {
      text: `${name} 近 20 日走势`,
      left: 10,
      top: 8,
      textStyle: { fontSize: 13, fontWeight: 600, color: '#173142' },
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const point = Array.isArray(params) ? params[0] : params;
        if (!point) return '';
        return `${point.axisValue}<br/>收盘价：¥${Number(point.data ?? 0).toFixed(2)}`;
      },
    },
    grid: { left: 16, right: 16, top: 44, bottom: 24 },
    xAxis: {
      type: 'category',
      data: items.map((item) => item.bar_time.slice(5, 10)),
      boundaryGap: false,
      axisLabel: { show: false },
      axisTick: { show: false },
      axisLine: { show: false },
    },
    yAxis: {
      type: 'value',
      scale: true,
      axisLabel: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#edf2ee' } },
    },
    series: [
      {
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: items.map((item) => toNumber(item.close_price)),
        lineStyle: { width: 2, color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22, 119, 255, 0.28)' },
              { offset: 1, color: 'rgba(22, 119, 255, 0.04)' },
            ],
          },
        },
      },
    ],
  };
}

export default function BoardDetailPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const boardType = searchParams.get('type') || '';
  const boardCode = searchParams.get('code') || '';
  const [detail, setDetail] = useState<MarketBoardDetailResponse | null>(null);
  const [boardNews, setBoardNews] = useState<BoardNewsResponse | null>(null);
  const [hoverKlines, setHoverKlines] = useState<Record<string, MarketStockKlineResponse | null>>({});
  const [hoverLoading, setHoverLoading] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!boardType || !boardCode) return;
    let active = true;
    const fetchDetail = async () => {
      setLoading(true);
      setError('');
      try {
        const [result, newsResult] = await Promise.all([
          marketApi.getBoardDetail(boardType, boardCode, { limit: 80 }),
          marketApi.getBoardNews(boardType, boardCode, { limit: 10 }),
        ]);
        if (!active) return;
        setDetail(result);
        setBoardNews(newsResult);
      } catch (err) {
        if (!active) return;
        setError(err instanceof Error ? err.message : '板块详情加载失败');
      } finally {
        if (active) setLoading(false);
      }
    };
    void fetchDetail();
    return () => {
      active = false;
    };
  }, [boardCode, boardType]);

  const loadHoverKline = async (symbol: string) => {
    if (!symbol || hoverKlines[symbol] || hoverLoading[symbol]) return;
    setHoverLoading((prev) => ({ ...prev, [symbol]: true }));
    try {
      const res = await marketApi.getStockKlines(symbol, { period: 'day', adjust: 'qfq', limit: 20 });
      setHoverKlines((prev) => ({ ...prev, [symbol]: res }));
    } catch {
      setHoverKlines((prev) => ({ ...prev, [symbol]: null }));
    } finally {
      setHoverLoading((prev) => ({ ...prev, [symbol]: false }));
    }
  };

  const renderConstituentName = (record: BoardConstituentResponse) => {
    const kline = hoverKlines[record.symbol];
    const loadingKline = hoverLoading[record.symbol];

    return (
      <Popover
        trigger="hover"
        placement="rightTop"
        mouseEnterDelay={0.2}
        overlayStyle={{ maxWidth: 360 }}
        onOpenChange={(open) => {
          if (open) {
            void loadHoverKline(record.symbol);
          }
        }}
        content={(
          <div style={{ width: 320 }}>
            {loadingKline ? (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 180 }}>
                <Spin size="small" />
              </div>
            ) : kline?.items?.length ? (
              <Space direction="vertical" size={8} style={{ width: '100%' }}>
                <ReactECharts option={buildMiniTrendOption(record.name || record.symbol, kline)} style={{ height: 190 }} />
                <Space size={12} wrap>
                  <Text type="secondary">最新价 {formatPrice(record.last_price)}</Text>
                  <Text style={{ color: getChangeColor(record.change_percent) }}>{formatSignedPercent(record.change_percent)}</Text>
                </Space>
              </Space>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无近期走势数据" />
            )}
          </div>
        )}
      >
        <Button type="link" style={{ padding: 0, height: 'auto' }} onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(record.symbol)}`)}>
          {record.name || record.symbol}
        </Button>
      </Popover>
    );
  };

  const newsItems = boardNews?.items ?? [];
  const latestNews = newsItems[0] ?? null;

  const columns = useMemo<ColumnsType<BoardConstituentResponse>>(() => [
    {
      title: '成分',
      dataIndex: 'name',
      key: 'name',
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          {renderConstituentName(record)}
          <Text type="secondary">{record.symbol}</Text>
        </Space>
      ),
    },
    {
      title: '最新价',
      dataIndex: 'last_price',
      key: 'last_price',
      align: 'right',
      render: (value) => formatPrice(value),
    },
    {
      title: '涨跌幅',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'right',
      render: (value) => <Text style={{ color: getChangeColor(value), fontWeight: 600 }}>{formatSignedPercent(value)}</Text>,
    },
    {
      title: '成交额',
      dataIndex: 'turnover',
      key: 'turnover',
      align: 'right',
      render: (value) => formatCompactNumber(value),
    },
    {
      title: '总市值',
      dataIndex: 'total_market_cap',
      key: 'total_market_cap',
      align: 'right',
      render: (value) => formatCompactNumber(value),
    },
    {
      title: '状态',
      dataIndex: 'has_snapshot',
      key: 'has_snapshot',
      align: 'center',
      render: (value: boolean) => value ? <Tag color="success">已定价</Tag> : <Tag color="warning">停牌</Tag>,
    },
  ], [navigate]);

  const coverageCards = detail?.coverage ?? [];
  const boardName = detail?.board.name || boardCode || '板块';

  return (
    <div style={{ padding: 24, background: 'linear-gradient(180deg, #f5f8ff 0%, #ffffff 42%)', minHeight: '100vh' }}>
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>
            <div>
              <Title level={3} style={{ margin: 0 }}>{boardName}</Title>
              <Text type="secondary">{boardType === 'industry' ? '行业板块' : boardType === 'concept' ? '概念板块' : boardType} · {boardCode}</Text>
            </div>
          </Space>
          <Space>
            <Button icon={<RobotOutlined />} type="primary" onClick={() => navigate(`/app/chat?kind=board&boardType=${encodeURIComponent(boardType)}&code=${encodeURIComponent(boardCode)}&name=${encodeURIComponent(boardName)}`)}>AI 板块对话</Button>
            <Button icon={<ReloadOutlined />} onClick={() => window.location.reload()}>刷新</Button>
          </Space>
        </Space>

        {error ? <Alert type="warning" showIcon message={error} /> : null}
        {detail?.message ? <Alert type={detail.is_partial ? 'warning' : 'info'} showIcon message={detail.message} /> : null}

        <Spin spinning={loading}>
          {!boardType || !boardCode ? (
            <Empty description="缺少板块参数" />
          ) : !detail ? (
            <Card bordered={false} style={{ borderRadius: 16 }}>
              <Empty description="暂无板块详情" />
            </Card>
          ) : (
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <Row gutter={[16, 16]}>
                <Col xs={24} xl={8}>
                  <Card bordered={false} style={{ borderRadius: 16, height: '100%' }}>
                    <Space direction="vertical" size={16} style={{ width: '100%' }}>
                      <Space align="center">
                        <RiseOutlined style={{ color: getChangeColor(detail.board.change_percent), fontSize: 18 }} />
                        <Text strong>{detail.board.name}</Text>
                        <Tag color={detail.board.board_type === 'industry' ? 'blue' : 'purple'}>{detail.board.board_type}</Tag>
                      </Space>
                      <Statistic title="板块涨跌幅" value={toNumber(detail.board.change_percent)} precision={2} suffix="%" valueStyle={{ color: getChangeColor(detail.board.change_percent) }} />
                      <Row gutter={[12, 12]}>
                        <Col span={12}><Statistic title="板块成交额" value={formatCompactNumber(detail.board.turnover)} /></Col>
                        <Col span={12}><Statistic title="成分股数量" value={detail.board.stock_count} /></Col>
                      </Row>
                      <Text type="secondary">快照时间：{detail.snapshot_time || '—'}</Text>
                      <Text type="secondary">刷新时间：{detail.refreshed_at || '—'}</Text>
                    </Space>
                  </Card>
                </Col>
                <Col xs={24} xl={16}>
                  <Row gutter={[12, 12]}>
                    {coverageCards.map((item) => (
                      <Col xs={12} md={8} key={item.label}>
                        <Card bordered={false} style={{ borderRadius: 16, background: '#f8fbff' }}>
                          <Statistic title={item.label} value={item.value} />
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </Col>
              </Row>

              <Row gutter={[16, 16]}>
                <Col xs={24} xl={8}>
                  <Card bordered={false} style={{ borderRadius: 16 }} title={<Space><BarChartOutlined /><span>涨跌分布</span></Space>}>
                    <ReactECharts option={buildAdvanceDeclineOption(detail)} style={{ height: 280 }} />
                  </Card>
                </Col>
                <Col xs={24} xl={8}>
                  <Card bordered={false} style={{ borderRadius: 16 }} title={<Space><RiseOutlined /><span>领涨成分</span></Space>}>
                    <ReactECharts option={buildTopMoveOption(detail.top_gainers.slice(0, 8), '涨幅前 8', '#ff4d4f')} style={{ height: 280 }} />
                  </Card>
                </Col>
                <Col xs={24} xl={8}>
                  <Card bordered={false} style={{ borderRadius: 16 }} title={<Space><StockOutlined /><span>成交额前排</span></Space>}>
                    <ReactECharts option={buildTurnoverOption(detail.top_turnover.slice(0, 8))} style={{ height: 280 }} />
                  </Card>
                </Col>
              </Row>

              <Row gutter={[16, 16]}>
                <Col xs={24} xl={12}>
                  <Card bordered={false} style={{ borderRadius: 16 }} title="领涨成分股">
                    <Table rowKey="symbol" dataSource={detail.top_gainers.slice(0, 8)} columns={columns} pagination={false} size="small" scroll={{ x: 720 }} />
                  </Card>
                </Col>
                <Col xs={24} xl={12}>
                  <Card bordered={false} style={{ borderRadius: 16 }} title="拖累成分股">
                    <Table rowKey="symbol" dataSource={detail.top_losers.slice(0, 8)} columns={columns} pagination={false} size="small" scroll={{ x: 720 }} />
                  </Card>
                </Col>
              </Row>

              <Card bordered={false} style={{ borderRadius: 16 }} title="板块新闻" extra={<Text type="secondary">覆盖 {boardNews?.coverage || '—'} · 最新 {latestNews?.published_at || '—'}</Text>}>
                {newsItems.length ? (
                  <List
                    dataSource={newsItems}
                    renderItem={(item) => (
                      <List.Item>
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <a href={item.url} target="_blank" rel="noreferrer" style={{ fontWeight: 600, color: '#1677ff' }}>
                            {item.title}
                          </a>
                          <Text type="secondary">{item.summary || '暂无摘要'}</Text>
                          <Space wrap size={[8, 8]}>
                            <Tag color={item.is_recent ? 'success' : 'default'}>{item.is_recent ? '近 7 日' : '较早新闻'}</Tag>
                            <Text type="secondary">{item.source || item.provider}</Text>
                            <Text type="secondary">{item.published_at}</Text>
                          </Space>
                        </Space>
                      </List.Item>
                    )}
                  />
                ) : (
                  <Empty description="暂无可追溯板块新闻" />
                )}
              </Card>

              <Card bordered={false} style={{ borderRadius: 16 }} title="全部成分股">
                <Table
                  rowKey="symbol"
                  dataSource={detail.constituents}
                  columns={columns}
                  pagination={{ pageSize: 20, showSizeChanger: false }}
                  scroll={{ x: 860 }}
                />
              </Card>
            </Space>
          )}
        </Spin>
      </Space>
    </div>
  );
}
