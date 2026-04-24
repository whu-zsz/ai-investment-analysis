import { useState, useEffect } from 'react';
import {
  Table, Card, Typography, Tag, Space, Button,
  Row, Col, Statistic, Alert, Spin, Empty,
} from 'antd';
import {
  HistoryOutlined,
  RiseOutlined, FallOutlined, ArrowLeftOutlined, BulbOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { transactionApi } from '../api/index';
import type { TransactionResponse, TransactionStats } from '../api/types';

const { Text, Title, Paragraph } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

export default function HistoryPage() {
  const navigate = useNavigate();
  const [transactions, setTransactions] = useState<TransactionResponse[]>([]);
  const [stats, setStats] = useState<TransactionStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);

  useEffect(() => { void fetchData(); }, [page]);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const [listRes, statsRes] = await Promise.all([
        transactionApi.getList({ page, page_size: 10 }),
        transactionApi.getStats(),
      ]);
      setTransactions(listRes.transactions);
      setTotal(listRes.total);
      setStats(statsRes);
    } catch (err: any) {
      setTransactions([]);
      setTotal(0);
      setStats(null);
      setError(err?.message ?? err?.data?.message ?? '交易记录加载失败');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<TransactionResponse> = [
    {
      title: '交易时间', dataIndex: 'transaction_date',
      sorter: (a, b) => a.transaction_date.localeCompare(b.transaction_date),
      render: (text) => <Text type="secondary" style={{ fontSize: 13 }}>{text}</Text>,
    },
    {
      title: '标的名称', dataIndex: 'asset_name',
      render: (text, row) => (
        <div>
          <Text strong>{text}</Text>
          <div style={{ color: '#8c8c8c', fontSize: 11 }}>{row.asset_code}</div>
        </div>
      ),
    },
    {
      title: '操作类型', dataIndex: 'transaction_type',
      render: (type: string) => (
        <Tag
          icon={type === 'buy' ? <RiseOutlined /> : <FallOutlined />}
          color={type === 'buy' ? 'processing' : 'default'}
          style={{ borderRadius: 20, padding: '2px 10px' }}
        >
          {type === 'buy' ? '买入' : type === 'sell' ? '卖出' : '分红'}
        </Tag>
      ),
    },
    {
      title: '成交均价', dataIndex: 'price_per_unit',
      render: (val) => `¥${parseFloat(val).toLocaleString()}`,
    },
    {
      title: '成交数量', dataIndex: 'quantity',
      render: (val) => parseFloat(val).toLocaleString(),
    },
    {
      title: '成交额', dataIndex: 'total_amount',
      render: (val) => <Text strong>¥{parseFloat(val).toLocaleString()}</Text>,
    },
    {
      title: '盈亏', dataIndex: 'profit',
      render: (val?: string) => {
        if (!val) return <Text type="secondary">—</Text>;
        const num = parseFloat(val);
        return (
          <Text strong style={{ color: num >= 0 ? '#52c41a' : '#ff4d4f' }}>
            {num >= 0 ? '+' : ''}¥{num.toLocaleString()}
          </Text>
        );
      },
    },
  ];

  const summaryStats = stats ? [
    { label: '总交易次数', value: stats.total_transactions, suffix: '次', color: '#1677ff' },
    { label: '买入次数', value: stats.buy_count, suffix: '次', color: '#52c41a' },
    { label: '卖出次数', value: stats.sell_count, suffix: '次', color: '#722ed1' },
    { label: '累计投入', value: `¥${parseFloat(stats.total_investment).toLocaleString()}`, suffix: '', color: '#262626' },
    { label: '累计盈亏', value: `¥${parseFloat(stats.total_profit).toLocaleString()}`, suffix: '', color: parseFloat(stats.total_profit) >= 0 ? '#52c41a' : '#ff4d4f' },
  ] : [];

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
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
              <Tag color="processing">历史归档</Tag>
              <Tag color="blue">分页流水与汇总统计</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>历史交易归档</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              当前页面展示后端已提供的分页交易流水和汇总统计，不再展示未接入的筛选或导出能力。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              累计盈亏 {stats ? `¥${parseFloat(stats.total_profit).toLocaleString()}` : '—'}
            </Tag>
            <Tag color="processing" icon={<HistoryOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              共 {stats?.total_transactions ?? '—'} 笔交易
            </Tag>
          </Space>
        </div>
      </Card>

      <Spin spinning={loading}>
        {error ? (
          <Card bordered={false} style={cardStyle}>
            <Alert type="error" showIcon message={error} />
          </Card>
        ) : (
          <>
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              {summaryStats.map(item => (
                <Col xs={12} sm={8} lg={4} key={item.label}>
                  <Card bordered={false} style={cardStyle}>
                    <Statistic
                      title={item.label}
                      value={item.value}
                      suffix={<span style={{ fontSize: 14, color: '#bfbfbf' }}>{item.suffix}</span>}
                      valueStyle={{ color: item.color, fontSize: 22 }}
                    />
                  </Card>
                </Col>
              ))}
            </Row>

            <Card bordered={false} style={cardStyle}
              title={<span><HistoryOutlined style={{ color: '#1677ff', marginRight: 8 }} />交易明细流水</span>}>
              {transactions.length ? (
                <Table
                  columns={columns}
                  dataSource={transactions.map(t => ({ ...t, key: t.id }))}
                  loading={false}
                  pagination={{
                    total, current: page, pageSize: 10,
                    onChange: (p) => setPage(p),
                    showTotal: (t) => `共 ${t} 条`,
                  }}
                  size="middle"
                />
              ) : (
                <Empty description="暂无交易记录" />
              )}
            </Card>

            <Card bordered={false} style={{ ...cardStyle, marginTop: 16 }}>
              <Alert type="info" showIcon icon={<BulbOutlined />}
                message="说明"
                description="本页当前仅展示真实的交易流水分页结果与汇总统计。若需要筛选、导出或更多分析能力，需要后端先提供对应接口。"
              />
            </Card>
          </>
        )}
      </Spin>
    </div>
  );
}
