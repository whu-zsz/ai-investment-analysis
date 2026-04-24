import { Alert, Button, Card, Col, Descriptions, Divider, Avatar, Dropdown, Popover, Row, Space, Statistic, Tag, Typography } from 'antd';
import {
  BulbOutlined,
  LineChartOutlined,
  RadarChartOutlined,
  RiseOutlined,
  ThunderboltOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { useMemo, useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { useAuth } from '../hooks/useAuth';
import { marketApi } from '../api/index';
import type { DashboardMarketSnapshotResponse } from '../api/types';
import { mockDashboardSnapshot } from '../mockData';
import type { MenuProps } from 'antd';

const { Paragraph, Text, Title } = Typography;

interface ChartParam {
  axisValueLabel?: string;
  name: string;
  value: number;
}

interface KpiChartConfig {
  seriesName: string;
  unit: string;
  labels: string[];
  values: number[];
  color: string;
}

interface DashboardInsightCard {
  title: string;
  value: number;
  precision: number;
  suffix: string;
  accent: string;
  tagColor: string;
  tagText: string;
  desc: string;
  chart: KpiChartConfig;
}

const chartPalette = ['#1677ff', '#52c41a', '#722ed1', '#fa8c16', '#13c2c2', '#eb2f96'];

const statColorMap: Record<string, string> = {
  指数数量: '#1677ff',
  上涨数: '#52c41a',
  下跌数: '#ff4d4f',
  平均涨跌幅: '#722ed1',
  总成交额: '#13c2c2',
};

function toNumber(value?: string): number {
  if (!value) return 0;
  const parsed = Number.parseFloat(value.replace(/[%亿,+]/g, ''));
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatChangePercent(changePercent?: string): string {
  if (!changePercent) return '0.00%';
  const normalized = changePercent.trim();
  if (normalized.endsWith('%')) {
    return normalized;
  }
  const numeric = toNumber(normalized);
  return `${numeric >= 0 ? '+' : ''}${normalized}%`;
}

function getTrendTag(changePercent?: string) {
  const numeric = toNumber(changePercent);
  if (numeric > 0) {
    return { color: 'green', text: `涨幅 ${formatChangePercent(changePercent)}` };
  }
  if (numeric < 0) {
    return { color: 'red', text: `跌幅 ${formatChangePercent(changePercent).replace('-', '')}` };
  }
  return { color: 'default', text: '平盘' };
}

function buildInsightCards(marketData: DashboardMarketSnapshotResponse | null): DashboardInsightCard[] {
  if (!marketData?.indices?.length) {
    return [];
  }

  return marketData.indices.slice(0, 4).map((item, index) => {
    const color = chartPalette[index % chartPalette.length];
    const trend = getTrendTag(item.change_percent);

    return {
      title: item.name,
      value: toNumber(item.last_price),
      precision: 2,
      suffix: '点',
      accent: color,
      tagColor: trend.color,
      tagText: trend.text,
      desc: `${item.symbol} · 最新涨跌额 ${item.change_amount}`,
      chart: {
        seriesName: item.name,
        unit: '点',
        labels: marketData.main_chart.series.map(point => point.label),
        values: marketData.main_chart.series.map(point => toNumber(point.value)),
        color,
      }
    };
  });
}


function getKpiChartOption(chart: KpiChartConfig, mode: 'mini' | 'expanded'): EChartsOption {
  const isMini = mode === 'mini';

  return {
    animation: true,
    tooltip: isMini
      ? { show: false }
      : {
          trigger: 'axis',
          backgroundColor: 'rgba(255, 255, 255, 0.96)',
          borderColor: '#d9e6ff',
          borderWidth: 1,
          formatter: (params: unknown) => {
            const list = params as ChartParam[];
            const data = list[0];
            return `<div style="padding: 4px 6px;">
                      <div style="color: #888; margin-bottom: 4px;">${data.axisValueLabel ?? data.name}</div>
                      <div style="font-weight: bold; color: ${chart.color}; font-size: 16px;">${data.value.toLocaleString()} ${chart.unit}</div>
                    </div>`;
          }
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
      show: !isMini,
      axisLabel: { color: '#8c8c8c', fontSize: 11 },
      splitLine: isMini ? { show: false } : { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: chart.seriesName,
        type: 'line',
        smooth: true,
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

export default function Dashboard() {
  const navigate = useNavigate();
  const location = useLocation();
  const isPublicHome = location.pathname === '/';
  const { isLoggedIn, userInfo, logout } = useAuth();

  // 市场数据状态
  const [marketData, setMarketData] = useState<DashboardMarketSnapshotResponse | null>(null);

  // 获取市场数据
  useEffect(() => {
    const fetchMarketData = async () => {
      try {
        const res = await marketApi.getDashboardSnapshot();
        setMarketData(res);
      } catch {
        setMarketData(mockDashboardSnapshot);
      }
    };
    fetchMarketData();
  }, []);

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
    if (key === 'logout') { logout(); navigate('/'); }
  };

  const getOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => {
        const list = params as ChartParam[];
        const data = list[0];
        return `<div style="padding: 4px 6px;">
                  <div style="color: #888; margin-bottom: 4px;">${data.name}</div>
                  <div style="font-weight: bold; color: #1677ff; font-size: 16px;">${data.value.toLocaleString()} 点</div>
                </div>`;
      }
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
      data: (marketData?.main_chart.series || []).map(point => point.label),
      axisLine: { lineStyle: { color: '#d9d9d9' } }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: marketData?.main_chart.index_name || '市场走势',
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: (marketData?.main_chart.series || []).map(point => toNumber(point.value)),
        lineStyle: { width: 3, color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22, 119, 255, 0.35)' },
              { offset: 1, color: 'rgba(22, 119, 255, 0.02)' }
            ]
          }
        }
      }
    ]
  });

  const insightCards = useMemo(() => buildInsightCards(marketData), [marketData]);

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
              <Button ghost onClick={() => guardNavigate('/app/analysis')}>AI 风险分析</Button>
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

      <Row gutter={[16, 16]}>
        {insightCards.map((item) => (
          <Col xs={24} sm={12} lg={6} key={item.title}>
            <Card
              bordered={false}
              hoverable
              onClick={() => navigate('/app/analysis')}
              style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15, 23, 42, 0.06)' }}
            >
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Text type="secondary">{item.title}</Text>
                  <Statistic
                    value={item.value}
                    precision={item.precision}
                    suffix={item.suffix}
                    valueStyle={{ color: item.accent, fontSize: 30 }}
                    style={{ marginTop: 8 }}
                  />
                </div>
                <div style={{ width: 124, flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 10 }}>
                  <Tag color={item.tagColor} style={{ marginInlineEnd: 0, textAlign: 'center' }}>{item.tagText}</Tag>
                  <Popover
                    trigger="hover"
                    placement="bottom"
                    overlayStyle={{ maxWidth: 360 }}
                    content={
                      <div style={{ width: 320 }}>
                        <Text strong>{item.title}走势明细</Text>
                        <ReactECharts option={getKpiChartOption(item.chart, 'expanded')} style={{ height: 220, marginTop: 8 }} />
                      </div>
                    }
                  >
                    <div style={{ cursor: 'zoom-in', borderRadius: 12, background: '#fafcff', border: '1px solid #eef2f6', padding: '4px 6px' }}>
                      <ReactECharts option={getKpiChartOption(item.chart, 'mini')} style={{ height: 62 }} />
                    </div>
                  </Popover>
                </div>
              </div>
              <Paragraph type="secondary" style={{ margin: '14px 0 0', minHeight: 44 }}>
                {item.desc}
              </Paragraph>
            </Card>
          </Col>
        ))}
      </Row>

      <Card
        title={
          <span>
            <LineChartOutlined style={{ marginRight: 8, color: '#1677ff' }} />
            今日大盘走势
          </span>
        }
        extra={
          <Space size={8} wrap>
            <Tag color={marketData?.is_stale ? 'warning' : 'processing'} icon={<RiseOutlined />}>
              {marketData?.main_chart.index_name || '市场走势'}
            </Tag>
            <Text type="secondary">更新时间 {marketData?.snapshot_time || '--'}</Text>
          </Space>
        }
        bordered={false}
        style={{ marginTop: 24, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.06)' }}
      >
        <ReactECharts option={getOption()} style={{ height: '400px' }} />

        <Row gutter={[12, 12]} style={{ marginTop: 8 }}>
          {quickStats.map((item) => (
            <Col xs={12} md={8} lg={4} key={item.label}>
              <div style={{ background: '#f8fafc', borderRadius: 12, padding: '14px 16px', border: '1px solid #eef2f6' }}>
                <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                <div style={{ marginTop: 6, fontSize: 20, fontWeight: 700, color: item.color }}>{item.value}</div>
              </div>
            </Col>
          ))}
        </Row>
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 8 }}>
        <Col span={24} lg={12}>
          <Card
            bordered={false}
            title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />市场快照</span>}
            style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Descriptions column={1} bordered size="small">
              {(marketData?.indices || []).map((item) => (
                <Descriptions.Item key={item.symbol} label={item.name}>
                  <Space split={<Divider type="vertical" />} size={8} wrap>
                    <Text strong>{item.last_price}</Text>
                    <Text type={toNumber(item.change_percent) >= 0 ? 'success' : 'danger'}>
                      {formatChangePercent(item.change_percent)}
                    </Text>
                    <Text type="secondary">{item.symbol}</Text>
                  </Space>
                </Descriptions.Item>
              ))}
            </Descriptions>
          </Card>
        </Col>

        <Col span={24} lg={12}>
          <Card
            bordered={false}
            title={<span><ThunderboltOutlined style={{ color: '#722ed1', marginRight: 8 }} />数据状态</span>}
            style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="数据源">
                <Text strong>{marketData?.source || '未知'}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="快照时间">
                <Text strong>{marketData?.snapshot_time || '--'}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="新鲜度">
                <Tag color={marketData?.is_stale ? 'warning' : 'success'}>
                  {marketData?.is_stale ? '数据可能已过期' : '数据新鲜'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="统计口径">
                基于当前最新批次指数快照自动聚合。
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ marginTop: 16, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
      >
        <Alert
          type={marketData?.is_stale ? 'warning' : 'info'}
          showIcon
          icon={<BulbOutlined />}
          message={`数据结论：当前共追踪 ${marketData?.indices.length || 0} 个指数，${marketData?.stats.find(item => item.label === '上涨数')?.value || '0'} 个上涨，${marketData?.stats.find(item => item.label === '下跌数')?.value || '0'} 个下跌。`}
          description={
            <Space direction="vertical" size={6}>
              <Text>主图展示 {marketData?.main_chart.index_name || '市场走势'}，数据来源为 {marketData?.source || '未知'}。</Text>
              <Text>{marketData?.is_stale ? '当前快照超过阈值，建议先刷新行情数据后再继续分析。' : '当前快照处于有效窗口，可继续联调其他依赖 dashboard 的页面。'}</Text>
            </Space>
          }
        />
      </Card>
    </div>
  );
}