import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Empty, Row, Space, Spin, Statistic, Table, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, FundOutlined, RiseOutlined, FallOutlined, LineChartOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { portfolioApi } from '../api';
import type { PortfolioResponse } from '../api/types';

const { Title, Text, Paragraph } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
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

  const summary = useMemo(() => {
    const totalMarketValue = portfolios.reduce((sum, item) => sum + toNumber(item.market_value), 0);
    const totalProfitLoss = portfolios.reduce((sum, item) => sum + toNumber(item.profit_loss), 0);
    const profitableCount = portfolios.filter((item) => toNumber(item.profit_loss) >= 0).length;
    return {
      totalMarketValue,
      totalProfitLoss,
      profitableCount,
    };
  }, [portfolios]);

  const columns: ColumnsType<PortfolioResponse> = [
    {
      title: '标的',
      dataIndex: 'asset_name',
      render: (_, row) => (
        <div>
          <Text strong>{row.asset_name}</Text>
          <div style={{ color: '#8c8c8c', fontSize: 12 }}>{row.asset_code}</div>
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'asset_type',
      render: (value: string) => <Tag>{value || 'stock'}</Tag>,
    },
    {
      title: '持仓数量',
      dataIndex: 'total_quantity',
      render: (value: string) => toNumber(value).toLocaleString(),
    },
    {
      title: '成本价',
      dataIndex: 'average_cost',
      render: (value: string) => `¥${toNumber(value).toLocaleString()}`,
    },
    {
      title: '现价',
      dataIndex: 'current_price',
      render: (value: string) => `¥${toNumber(value).toLocaleString()}`,
    },
    {
      title: '市值',
      dataIndex: 'market_value',
      render: (value: string) => <Text strong>¥{toNumber(value).toLocaleString()}</Text>,
    },
    {
      title: '盈亏',
      dataIndex: 'profit_loss',
      render: (value: string, row) => {
        const profit = toNumber(value);
        const profitPercent = toNumber(row.profit_loss_percent);
        return (
          <Space direction="vertical" size={0}>
            <Text strong style={{ color: profit >= 0 ? '#52c41a' : '#ff4d4f' }}>
              {profit >= 0 ? '+' : ''}¥{profit.toLocaleString()}
            </Text>
            <Text type={profitPercent >= 0 ? 'success' : 'danger'}>
              {profitPercent >= 0 ? '+' : ''}{profitPercent.toFixed(2)}%
            </Text>
          </Space>
        );
      },
    },
    {
      title: '更新时间',
      dataIndex: 'last_updated',
      render: (value: string) => <Text type="secondary">{value}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, row) => (
        <Button type="link" icon={<LineChartOutlined />} onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(row.asset_code)}`)}>
          查看趋势
        </Button>
      ),
    },
  ];

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
              <Tag color="processing">持仓总览</Tag>
              <Tag color="blue">来自后端实时持仓</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>当前持仓视图</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 620 }}>
              聚合展示你的当前持仓、市值和浮盈亏，并可直接跳转查看单个标的的市场趋势。
            </Paragraph>
          </div>
          <Button ghost onClick={() => void load()} style={{ borderRadius: 10 }}>刷新持仓</Button>
        </div>
      </Card>

      <Spin spinning={loading}>
        {error ? (
          <Card bordered={false} style={cardStyle}>
            <Alert type="error" showIcon message={error} />
          </Card>
        ) : !portfolios.length ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description="暂无持仓数据，请先导入交易记录" />
          </Card>
        ) : (
          <>
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              <Col xs={12} sm={12} lg={6}>
                <Card bordered={false} style={cardStyle}><Statistic title="持仓标的数" value={portfolios.length} prefix={<FundOutlined />} /></Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card bordered={false} style={cardStyle}><Statistic title="总市值" value={summary.totalMarketValue} precision={2} prefix="¥" valueStyle={{ color: '#1677ff' }} /></Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card bordered={false} style={cardStyle}><Statistic title="总浮盈亏" value={summary.totalProfitLoss} precision={2} prefix="¥" valueStyle={{ color: summary.totalProfitLoss >= 0 ? '#52c41a' : '#ff4d4f' }} /></Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card bordered={false} style={cardStyle}><Statistic title="盈利持仓数" value={summary.profitableCount} prefix={summary.profitableCount >= portfolios.length / 2 ? <RiseOutlined /> : <FallOutlined />} valueStyle={{ color: '#52c41a' }} /></Card>
              </Col>
            </Row>

            <Card bordered={false} style={cardStyle} title={<span><FundOutlined style={{ color: '#1677ff', marginRight: 8 }} />持仓明细</span>}>
              <Table rowKey="id" columns={columns} dataSource={portfolios} pagination={false} scroll={{ x: 'max-content' }} />
            </Card>
          </>
        )}
      </Spin>
    </div>
  );
}
