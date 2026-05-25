import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Empty, Progress, Row, Space, Spin, Statistic, Table, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined, FundOutlined, LineChartOutlined, ReloadOutlined, RobotOutlined, WarningOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import type { ColumnsType } from 'antd/es/table';
import { portfolioApi } from '../api';
import type { PortfolioResponse } from '../api/types';

const { Title, Text, Paragraph } = Typography;
const cardStyle = { borderRadius: 18, boxShadow: '0 10px 28px rgba(15,23,42,0.07)' };

type EnrichedPortfolio = PortfolioResponse & {
  quantityValue: number;
  averageCostValue: number;
  costAmount: number;
  currentPriceValue: number;
  marketValueNumber: number;
  profitLossNumber: number;
  profitLossPercentNumber: number;
  allocationPercent: number;
  priceReady: boolean;
};

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

function money(value: number, digits = 2) {
  return `¥${value.toLocaleString(undefined, { minimumFractionDigits: digits, maximumFractionDigits: digits })}`;
}

function signedMoney(value: number) {
  const sign = value > 0 ? '+' : value < 0 ? '-' : '';
  return `${sign}${money(Math.abs(value))}`;
}

function signedPercent(value: number) {
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}%`;
}

function profitColor(value: number) {
  if (value > 0) return '#d4380d';
  if (value < 0) return '#237804';
  return '#31566f';
}

function assetTypeLabel(value?: string) {
  if (value === 'fund') return '基金';
  if (value === 'stock') return '股票';
  return value || '其他';
}

function isPriced(item: PortfolioResponse) {
  return toNumber(item.current_price) > 0 && toNumber(item.market_value) > 0;
}

function buildAllocationOption(items: EnrichedPortfolio[]): EChartsOption {
  const priced = items.filter((item) => item.priceReady && item.marketValueNumber > 0);
  return {
    tooltip: { trigger: 'item', formatter: '{b}<br/>市值：¥{c}<br/>占比：{d}%' },
    legend: { bottom: 0, type: 'scroll' },
    series: [{
      name: '持仓占比',
      type: 'pie',
      radius: ['48%', '72%'],
      center: ['50%', '44%'],
      avoidLabelOverlap: true,
      label: { formatter: '{b}\n{d}%' },
      data: priced.map((item) => ({ name: item.asset_name || item.asset_code, value: Number(item.marketValueNumber.toFixed(2)) })),
    }],
  };
}

function buildProfitOption(items: EnrichedPortfolio[]): EChartsOption {
  const sorted = [...items].sort((a, b) => a.profitLossNumber - b.profitLossNumber);
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 48, right: 24, top: 24, bottom: 36, containLabel: true },
    xAxis: { type: 'category', data: sorted.map((item) => item.asset_name || item.asset_code), axisLabel: { interval: 0, rotate: 18 } },
    yAxis: { type: 'value', axisLabel: { formatter: (value: number) => `${(value / 1000).toFixed(0)}k` } },
    series: [{
      name: '浮动盈亏',
      type: 'bar',
      data: sorted.map((item) => ({
        value: Number(item.profitLossNumber.toFixed(2)),
        itemStyle: { color: profitColor(item.profitLossNumber) },
      })),
      label: { show: true, position: 'top', formatter: ({ value }) => money(Number(value), 0) },
    }],
  };
}

function buildTypeOption(items: EnrichedPortfolio[]): EChartsOption {
  const grouped = items.reduce<Record<string, number>>((acc, item) => {
    const label = assetTypeLabel(item.asset_type);
    acc[label] = (acc[label] ?? 0) + item.marketValueNumber;
    return acc;
  }, {});
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{
      name: '资产类型',
      type: 'pie',
      radius: ['42%', '68%'],
      center: ['50%', '44%'],
      data: Object.entries(grouped).map(([name, value]) => ({ name, value: Number(value.toFixed(2)) })),
    }],
  };
}

export default function PortfolioPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [portfolios, setPortfolios] = useState<PortfolioResponse[]>([]);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await portfolioApi.getList();
      setPortfolios(Array.isArray(res) ? res : []);
    } catch (err: any) {
      setPortfolios([]);
      setError(err?.message ?? err?.data?.message ?? '持仓数据加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const enriched = useMemo<EnrichedPortfolio[]>(() => {
    const totalMarketValue = portfolios.reduce((sum, item) => sum + toNumber(item.market_value), 0);
    return portfolios.map((item) => {
      const quantityValue = toNumber(item.total_quantity);
      const averageCostValue = toNumber(item.average_cost);
      const costAmount = quantityValue * averageCostValue;
      const marketValueNumber = toNumber(item.market_value);
      const profitLossNumber = toNumber(item.profit_loss);
      return {
        ...item,
        quantityValue,
        averageCostValue,
        costAmount,
        currentPriceValue: toNumber(item.current_price),
        marketValueNumber,
        profitLossNumber,
        profitLossPercentNumber: toNumber(item.profit_loss_percent),
        allocationPercent: totalMarketValue > 0 ? (marketValueNumber / totalMarketValue) * 100 : 0,
        priceReady: isPriced(item),
      };
    });
  }, [portfolios]);

  const summary = useMemo(() => {
    const totalMarketValue = enriched.reduce((sum, item) => sum + item.marketValueNumber, 0);
    const totalCost = enriched.reduce((sum, item) => sum + item.costAmount, 0);
    const totalProfitLoss = enriched.reduce((sum, item) => sum + item.profitLossNumber, 0);
    const pricedCount = enriched.filter((item) => item.priceReady).length;
    const unpriced = enriched.filter((item) => !item.priceReady);
    const maxLoss = enriched.reduce<EnrichedPortfolio | null>((current, item) => !current || item.profitLossNumber < current.profitLossNumber ? item : current, null);
    const maxAllocation = enriched.reduce<EnrichedPortfolio | null>((current, item) => !current || item.allocationPercent > current.allocationPercent ? item : current, null);
    return {
      totalMarketValue,
      totalCost,
      totalProfitLoss,
      totalReturn: totalCost > 0 ? (totalProfitLoss / totalCost) * 100 : 0,
      pricedCount,
      pricedRatio: enriched.length ? (pricedCount / enriched.length) * 100 : 0,
      unpriced,
      maxLoss,
      maxAllocation,
    };
  }, [enriched]);

  const riskTips = useMemo(() => {
    const tips: Array<{ type: 'info' | 'warning' | 'error' | 'success'; message: string; description?: string }> = [];
    if (summary.maxLoss && summary.maxLoss.profitLossNumber < 0) {
      tips.push({
        type: 'error',
        message: `最大亏损：${summary.maxLoss.asset_name}`,
        description: `${signedMoney(summary.maxLoss.profitLossNumber)}，收益率 ${signedPercent(summary.maxLoss.profitLossPercentNumber)}`,
      });
    }
    if (summary.maxAllocation) {
      tips.push({
        type: summary.maxAllocation.allocationPercent >= 40 ? 'warning' : 'info',
        message: `最大仓位：${summary.maxAllocation.asset_name}`,
        description: `占组合市值 ${summary.maxAllocation.allocationPercent.toFixed(2)}%`,
      });
    }
    if (summary.unpriced.length) {
      tips.push({
        type: 'warning',
        message: `待取价标的：${summary.unpriced.map((item) => item.asset_name || item.asset_code).join('、')}`,
        description: '系统没有写入不确定价格，避免把指数或错误标的当成持仓市价。',
      });
    }
    if (!tips.length) {
      tips.push({ type: 'success', message: '持仓估值完整', description: '全部标的已有可用价格，可用于组合市值和浮盈亏计算。' });
    }
    return tips;
  }, [summary]);

  const columns: ColumnsType<EnrichedPortfolio> = [
    {
      title: '标的',
      dataIndex: 'asset_name',
      fixed: 'left',
      width: 190,
      render: (_, row) => (
        <Space direction="vertical" size={2}>
          <Text strong>{row.asset_name}</Text>
          <Text type="secondary">{row.asset_code}</Text>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'asset_type',
      width: 92,
      render: (value: string) => <Tag color={value === 'fund' ? 'purple' : 'blue'}>{assetTypeLabel(value)}</Tag>,
    },
    {
      title: '价格状态',
      key: 'price_status',
      width: 110,
      render: (_, row) => row.priceReady ? <Tag color="success">已定价</Tag> : <Tag color="warning">待取价</Tag>,
      filters: [
        { text: '已定价', value: 'priced' },
        { text: '待取价', value: 'unpriced' },
      ],
      onFilter: (value, row) => value === 'priced' ? row.priceReady : !row.priceReady,
    },
    {
      title: '持仓数量',
      dataIndex: 'total_quantity',
      align: 'right',
      sorter: (a, b) => a.quantityValue - b.quantityValue,
      render: (_, row) => row.quantityValue.toLocaleString(),
    },
    {
      title: '持仓成本',
      key: 'cost_amount',
      align: 'right',
      sorter: (a, b) => a.costAmount - b.costAmount,
      render: (_, row) => (
        <Space direction="vertical" size={0} align="end">
          <Text>{money(row.costAmount)}</Text>
          <Text type="secondary">成本价 {money(row.averageCostValue)}</Text>
        </Space>
      ),
    },
    {
      title: '现价',
      dataIndex: 'current_price',
      align: 'right',
      render: (_, row) => row.priceReady ? money(row.currentPriceValue) : <Text type="secondary">待取价</Text>,
    },
    {
      title: '市值 / 占比',
      key: 'market_value',
      align: 'right',
      sorter: (a, b) => a.marketValueNumber - b.marketValueNumber,
      render: (_, row) => (
        <Space direction="vertical" size={4} style={{ minWidth: 120 }}>
          <Text strong>{money(row.marketValueNumber)}</Text>
          <Progress percent={Number(row.allocationPercent.toFixed(2))} size="small" showInfo={false} />
          <Text type="secondary">{row.allocationPercent.toFixed(2)}%</Text>
        </Space>
      ),
    },
    {
      title: '浮动盈亏',
      key: 'profit_loss',
      align: 'right',
      sorter: (a, b) => a.profitLossNumber - b.profitLossNumber,
      render: (_, row) => (
        <Space direction="vertical" size={0} align="end">
          <Text strong style={{ color: profitColor(row.profitLossNumber) }}>{signedMoney(row.profitLossNumber)}</Text>
          <Text style={{ color: profitColor(row.profitLossPercentNumber) }}>{signedPercent(row.profitLossPercentNumber)}</Text>
        </Space>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'last_updated',
      width: 168,
      render: (value: string) => <Text type="secondary">{value}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 170,
      render: (_, row) => (
        <Space>
          <Button type="link" icon={<LineChartOutlined />} onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(row.asset_code)}`)}>
            趋势
          </Button>
          <Button type="link" icon={<RobotOutlined />} onClick={() => navigate(`/app/chat?kind=stock&symbol=${encodeURIComponent(row.asset_code)}`)}>
            AI
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ minHeight: '100vh', padding: 24, background: 'linear-gradient(180deg, #f5f8ff 0%, #eef4ff 44%, #f7f8fb 100%)' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')} style={{ marginBottom: 16, color: '#31566f', paddingLeft: 0 }}>
        返回首页
      </Button>

      <Card bordered={false} style={{ marginBottom: 18, borderRadius: 22, background: 'linear-gradient(135deg, #0f172a 0%, #1457d9 58%, #69b1ff 100%)', boxShadow: '0 18px 46px rgba(20,87,217,0.18)' }} bodyStyle={{ padding: 30 }}>
        <Row gutter={[20, 20]} align="middle">
          <Col span={24} lg={16}>
            <Space size={10} wrap style={{ marginBottom: 12 }}>
              <Tag color="processing">持仓总览</Tag>
              <Tag color="blue">估值仪表盘</Tag>
              <Tag color={summary.pricedRatio === 100 ? 'success' : 'warning'}>可定价 {summary.pricedRatio.toFixed(0)}%</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>当前持仓视图</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 760 }}>
              聚合展示组合市值、投入成本、浮动盈亏、仓位分布和待取价风险，帮助快速判断资金集中度与收益贡献。
            </Paragraph>
          </Col>
          <Col span={24} lg={8} style={{ textAlign: 'right' }}>
            <Button ghost icon={<ReloadOutlined />} loading={loading} onClick={() => void load()} style={{ borderRadius: 10 }}>
              刷新估值
            </Button>
          </Col>
        </Row>
      </Card>

      <Spin spinning={loading}>
        {error ? (
          <Card bordered={false} style={cardStyle}><Alert type="error" showIcon message={error} /></Card>
        ) : !enriched.length ? (
          <Card bordered={false} style={cardStyle}><Empty description="暂无持仓数据，请先导入交易记录" /></Card>
        ) : (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Row gutter={[14, 14]}>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="总市值" value={summary.totalMarketValue} precision={2} prefix="¥" valueStyle={{ color: '#1457d9' }} /></Card></Col>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="总成本" value={summary.totalCost} precision={2} prefix="¥" /></Card></Col>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="总浮盈亏" value={summary.totalProfitLoss} precision={2} prefix="¥" valueStyle={{ color: profitColor(summary.totalProfitLoss) }} /></Card></Col>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="总收益率" value={summary.totalReturn} precision={2} suffix="%" valueStyle={{ color: profitColor(summary.totalReturn) }} /></Card></Col>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="持仓数量" value={enriched.length} prefix={<FundOutlined />} /></Card></Col>
              <Col xs={12} lg={4}><Card bordered={false} style={cardStyle}><Statistic title="可定价比例" value={summary.pricedRatio} precision={0} suffix="%" prefix={<BarChartOutlined />} /></Card></Col>
            </Row>

            <Row gutter={[16, 16]}>
              <Col span={24} xl={8}>
                <Card bordered={false} style={cardStyle} title="持仓市值占比">
                  {summary.totalMarketValue > 0 ? <ReactECharts option={buildAllocationOption(enriched)} style={{ height: 320 }} /> : <Empty description="暂无可定价市值" />}
                </Card>
              </Col>
              <Col span={24} xl={8}>
                <Card bordered={false} style={cardStyle} title="盈亏贡献">
                  <ReactECharts option={buildProfitOption(enriched)} style={{ height: 320 }} />
                </Card>
              </Col>
              <Col span={24} xl={8}>
                <Card bordered={false} style={cardStyle} title="资产类型分布">
                  {summary.totalMarketValue > 0 ? <ReactECharts option={buildTypeOption(enriched)} style={{ height: 320 }} /> : <Empty description="暂无可定价市值" />}
                </Card>
              </Col>
            </Row>

            <Card bordered={false} style={cardStyle} title={<span><WarningOutlined style={{ color: '#fa8c16', marginRight: 8 }} />重点提示</span>}>
              <Row gutter={[12, 12]}>
                {riskTips.map((tip) => (
                  <Col span={24} lg={8} key={`${tip.message}-${tip.description}`}>
                    <Alert type={tip.type} showIcon message={tip.message} description={tip.description} />
                  </Col>
                ))}
              </Row>
            </Card>

            <Card bordered={false} style={cardStyle} title={<span><FundOutlined style={{ color: '#1677ff', marginRight: 8 }} />持仓明细</span>}>
              <Table
                rowKey="id"
                columns={columns}
                dataSource={enriched}
                pagination={{ pageSize: 8, showSizeChanger: true }}
                scroll={{ x: 1320 }}
              />
            </Card>
          </Space>
        )}
      </Spin>
    </div>
  );
}
