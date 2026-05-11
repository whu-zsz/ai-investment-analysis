import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Empty, List, Row, Space, Spin, Statistic, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined, ReloadOutlined, RiseOutlined, StockOutlined } from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi, marketApi } from '../api';
import type { AnalysisCandidateResponse, MarketSnapshotResponse } from '../api/types';

const { Title, Text, Paragraph } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

export default function MarketTrendPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [candidates, setCandidates] = useState<AnalysisCandidateResponse[]>([]);
  const [history, setHistory] = useState<MarketSnapshotResponse[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState('');

  const load = async (preferSymbol?: string) => {
    setLoading(true);
    setError('');
    try {
      const candidateRes = await analysisApi.getCandidates();
      const candidateList = candidateRes.candidates ?? [];
      setCandidates(candidateList);
      const nextSymbol = preferSymbol || searchParams.get('symbol') || candidateRes.default_symbol || candidateList[0]?.symbol || '';
      setSelectedSymbol(nextSymbol);
      if (!nextSymbol) {
        setHistory([]);
        return;
      }
      const historyRes = await marketApi.getSnapshotHistory({ symbol: nextSymbol, limit: 30 });
      setHistory(Array.isArray(historyRes) ? [...historyRes].reverse() : []);
      setSearchParams({ symbol: nextSymbol });
    } catch (err: any) {
      setCandidates([]);
      setHistory([]);
      setError(err?.message ?? err?.data?.message ?? '市场趋势加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const selectedCandidate = useMemo(
    () => candidates.find((item) => item.symbol === selectedSymbol) ?? null,
    [candidates, selectedSymbol],
  );

  const latestPoint = history[history.length - 1] ?? null;

  const chartOption: EChartsOption = {
    tooltip: { trigger: 'axis' },
    grid: { top: 24, left: 36, right: 20, bottom: 28, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: history.map((item) => item.snapshot_time.slice(5, 16)),
      axisLine: { lineStyle: { color: '#d9d9d9' } },
      axisLabel: { color: '#8c8c8c', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } },
      axisLabel: { color: '#8c8c8c', fontSize: 11 },
    },
    series: [
      {
        name: '最新价',
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: history.map((item) => toNumber(item.last_price)),
        lineStyle: { width: 3, color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22,119,255,0.25)' },
              { offset: 1, color: 'rgba(22,119,255,0.03)' },
            ],
          },
        },
      },
    ],
  };

  const onSelectSymbol = (symbol: string) => {
    void load(symbol);
  };

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')} style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
        返回首页
      </Button>

      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">市场趋势</Tag>
              <Tag color="blue">候选池来自用户相关标的</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>标的趋势追踪</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 620 }}>
              基于你的持仓和历史交易生成候选股票池，查看单个标的的最近市场快照走势和关键行情数据。
            </Paragraph>
          </div>
          <Button ghost icon={<ReloadOutlined />} onClick={() => void load(selectedSymbol)} loading={loading} style={{ borderRadius: 10 }}>
            刷新趋势
          </Button>
        </div>
      </Card>

      <Spin spinning={loading}>
        {error ? (
          <Card bordered={false} style={cardStyle}><Alert type="error" showIcon message={error} /></Card>
        ) : !candidates.length ? (
          <Card bordered={false} style={cardStyle}><Empty description="暂无候选标的，请先导入交易记录或生成持仓" /></Card>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={24} lg={7}>
              <Card bordered={false} style={cardStyle} title={<span><StockOutlined style={{ color: '#1677ff', marginRight: 8 }} />候选标的</span>}>
                <List
                  dataSource={candidates}
                  renderItem={(item) => (
                    <List.Item onClick={() => onSelectSymbol(item.symbol)} style={{ cursor: 'pointer', paddingInline: 0 }}>
                      <div style={{ width: '100%', padding: 12, borderRadius: 12, background: item.symbol === selectedSymbol ? '#e6f4ff' : '#fafafa', border: item.symbol === selectedSymbol ? '1px solid #91caff' : '1px solid #f0f0f0' }}>
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Space>
                            <Text strong>{item.asset_name || item.symbol}</Text>
                            <Tag color={item.is_held ? 'success' : 'default'}>{item.is_held ? '已持仓' : '历史关注'}</Tag>
                          </Space>
                          <Text type="secondary">{item.symbol}</Text>
                          <Text type={toNumber(item.change_percent) >= 0 ? 'success' : 'danger'}>
                            涨跌幅 {toNumber(item.change_percent).toFixed(2)}%
                          </Text>
                        </Space>
                      </div>
                    </List.Item>
                  )}
                />
              </Card>
            </Col>

            <Col span={24} lg={17}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Row gutter={[16, 16]}>
                  <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="候选数量" value={candidates.length} /></Card></Col>
                  <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="最新价格" value={toNumber(latestPoint?.last_price)} precision={2} prefix="¥" valueStyle={{ color: '#1677ff' }} /></Card></Col>
                  <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="最新涨跌幅" value={toNumber(latestPoint?.change_percent)} precision={2} suffix="%" valueStyle={{ color: toNumber(latestPoint?.change_percent) >= 0 ? '#52c41a' : '#ff4d4f' }} /></Card></Col>
                  <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="交易关注次数" value={selectedCandidate?.trade_count ?? 0} prefix={<RiseOutlined />} /></Card></Col>
                </Row>

                <Card bordered={false} style={cardStyle} title={<span><BarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />价格走势</span>} extra={<Text type="secondary">{selectedSymbol || '未选择标的'}</Text>}>
                  {history.length ? <ReactECharts option={chartOption} style={{ height: 320 }} /> : <Empty description="暂无该标的的历史快照" />}
                </Card>

                <Card bordered={false} style={cardStyle} title="最新快照摘要">
                  {latestPoint ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12} lg={6}><Statistic title="开盘价" value={toNumber(latestPoint.open_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="最高价" value={toNumber(latestPoint.high_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="最低价" value={toNumber(latestPoint.low_price)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="昨收价" value={toNumber(latestPoint.prev_close)} precision={2} prefix="¥" /></Col>
                      <Col span={12} lg={6}><Statistic title="成交量" value={toNumber(latestPoint.volume)} precision={0} /></Col>
                      <Col span={12} lg={6}><Statistic title="成交额" value={toNumber(latestPoint.turnover)} precision={2} /></Col>
                      <Col span={24} lg={12}><Text type="secondary">数据源：{latestPoint.source || '—'}</Text></Col>
                      <Col span={24} lg={12}><Text type="secondary">快照时间：{latestPoint.snapshot_time}</Text></Col>
                    </Row>
                  ) : (
                    <Empty description="暂无快照摘要" />
                  )}
                </Card>
              </Space>
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
}
