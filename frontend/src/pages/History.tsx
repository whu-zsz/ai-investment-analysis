import { useEffect, useMemo, useState } from 'react';
import {
  Table, Card, Typography, Tag, Input, Space, Button,
  Row, Col, Statistic, Alert
} from 'antd';
import {
  SearchOutlined, DownloadOutlined, HistoryOutlined,
  RiseOutlined, FallOutlined, ArrowLeftOutlined, BulbOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../types';

const { Text, Title, Paragraph } = Typography;

interface TransactionItem {
  id: number;
  transaction_date: string;
  transaction_type: string;
  asset_type: string;
  asset_code: string;
  asset_name: string;
  quantity: string;
  price_per_unit: string;
  total_amount: string;
  commission: string;
  profit: string | null;
  notes: string | null;
  created_at: string;
}

interface TransactionListResponse {
  transactions: TransactionItem[];
  total: number;
  page: number;
  page_size: number;
}

interface TransactionStats {
  total_transactions: number;
  buy_count: number;
  sell_count: number;
  total_investment: string;
  total_profit: string;
}

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

export default function HistoryPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [stats, setStats] = useState<TransactionStats | null>(null);
  const [records, setRecords] = useState<TransactionItem[]>([]);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const [listResponse, statsResponse] = await Promise.all([
          api.getTransactions({ page: 1, page_size: 50 }),
          api.getTransactionStats(),
        ]);
        const listData = listResponse.data as TransactionListResponse;
        setRecords(listData.transactions);
        setStats(statsResponse.data as TransactionStats);
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, []);

  const filteredRecords = useMemo(() => {
    if (!search.trim()) return records;
    return records.filter(record =>
      `${record.asset_name} ${record.asset_code}`.toLowerCase().includes(search.trim().toLowerCase())
    );
  }, [records, search]);

  const columns: ColumnsType<TransactionItem> = [
    {
      title: '交易时间', dataIndex: 'transaction_date',
      sorter: (a, b) => a.transaction_date.localeCompare(b.transaction_date),
      render: text => <Text type="secondary" style={{ fontSize: 13 }}>{text}</Text>,
    },
    {
      title: '标的名称', key: 'asset',
      render: (_, row) => <Text strong>{row.asset_name} ({row.asset_code})</Text>,
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
      render: val => `¥${Number(val).toLocaleString()}`,
    },
    {
      title: '成交数量', dataIndex: 'quantity',
      render: val => Number(val).toLocaleString(),
    },
    {
      title: '成交额', dataIndex: 'total_amount',
      render: val => <Text strong>¥{Number(val).toLocaleString()}</Text>,
    },
    {
      title: '浮动盈亏', dataIndex: 'profit',
      render: (val: string | null) => {
        const number = Number(val ?? 0);
        return (
          <Text strong style={{ color: number >= 0 ? '#52c41a' : '#ff4d4f' }}>
            {number >= 0 ? '+' : ''}¥{number.toLocaleString()}
          </Text>
        );
      },
    },
    {
      title: '备注', dataIndex: 'notes',
      render: value => value ?? '—',
    },
  ];

  const summaryStats = [
    { label: '总交易次数', value: stats?.total_transactions ?? 0, suffix: '次', color: '#1677ff' },
    { label: '总成交额', value: Number(stats?.total_investment ?? 0), suffix: '元', color: '#262626' },
    { label: '累计盈亏', value: Number(stats?.total_profit ?? 0), suffix: '元', color: '#52c41a' },
    { label: '买入次数', value: stats?.buy_count ?? 0, suffix: '次', color: '#1677ff' },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}
      >
        返回首页
      </Button>

      <Card
        bordered={false}
        style={{
          marginBottom: 24, borderRadius: 20,
          background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
          boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
        }}
        bodyStyle={{ padding: 28 }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">历史归档</Tag>
              <Tag color="blue">真实交易接口</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>历史交易归档</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              交易明细与统计已接入后端 `/transactions` 与 `/transactions/stats`。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              累计盈亏 ¥{Number(stats?.total_profit ?? 0).toLocaleString()}
            </Tag>
            <Tag color="processing" icon={<HistoryOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              总交易 {stats?.total_transactions ?? 0} 笔
            </Tag>
          </Space>
        </div>
      </Card>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        {summaryStats.map(item => (
          <Col xs={12} sm={6} key={item.label}>
            <Card bordered={false} style={cardStyle}>
              <Statistic
                title={item.label}
                value={item.value}
                suffix={<span style={{ fontSize: 14, color: '#bfbfbf' }}>{item.suffix}</span>}
                valueStyle={{ color: item.color, fontSize: 26 }}
              />
            </Card>
          </Col>
        ))}
      </Row>

      <Card bordered={false} style={{ ...cardStyle, marginBottom: 16 }}>
        <Space wrap size="middle">
          <Input
            placeholder="搜索标的名称 / 代码"
            prefix={<SearchOutlined />}
            style={{ width: 260, borderRadius: 10 }}
            allowClear
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
          <Button type="primary" icon={<SearchOutlined />} style={{ borderRadius: 10 }}>查询</Button>
          <Button icon={<DownloadOutlined />} style={{ borderRadius: 10 }} disabled>导出 CSV</Button>
        </Space>
      </Card>

      <Card bordered={false} style={cardStyle} title={<span><HistoryOutlined style={{ color: '#1677ff', marginRight: 8 }} />交易明细流水</span>}>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={filteredRecords}
          loading={loading}
          pagination={{ pageSize: 10 }}
          size="middle"
        />
      </Card>

      <Card bordered={false} style={{ ...cardStyle, marginTop: 16 }}>
        <Alert
          type="info"
          showIcon
          icon={<BulbOutlined />}
          message="当前历史页已切换为真实后端数据；导出功能因后端暂无专门接口，先保留占位。"
        />
      </Card>
    </div>
  );
}
