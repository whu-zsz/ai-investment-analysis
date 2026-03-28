import { useState } from 'react';
import { Table, Card, Typography, Tag, Input, Space, Button, DatePicker, ConfigProvider, theme, Row, Col } from 'antd';
import { SearchOutlined, DownloadOutlined, HistoryOutlined, RiseOutlined, FallOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import PageHeader from '../components/PageHeader';

const { Text } = Typography;
const { RangePicker } = DatePicker;

interface TradeRecord {
  key: string;
  date: string;
  asset: string;
  type: '买入' | '卖出';
  price: number;
  amount: number;
  total: number;
  status: '已成交' | '已撤单';
  pnl: number;
}

const cardStyle = {
  background: 'rgba(15, 23, 42, 0.6)',
  border: '1px solid rgba(255,255,255,0.08)',
  borderRadius: 16,
  backdropFilter: 'blur(10px)',
};

export default function HistoryPage() {
  const [loading] = useState(false);

  const mockData: TradeRecord[] = [
    { key: '1', date: '2024-03-22 14:30', asset: '腾讯控股 (0700.HK)',   type: '买入', price: 290.5,  amount: 100,  total: 29050, status: '已成交', pnl:  1240 },
    { key: '2', date: '2024-03-21 10:15', asset: '贵州茅台 (600519)',    type: '卖出', price: 1720.0, amount: 10,   total: 17200, status: '已成交', pnl:  -380 },
    { key: '3', date: '2024-03-20 09:45', asset: '纳指100ETF (513100)', type: '买入', price: 1.25,   amount: 5000, total:  6250, status: '已成交', pnl:   860 },
    { key: '4', date: '2024-03-19 15:00', asset: '英伟达 (NVDA.US)',     type: '买入', price: 890.2,  amount: 5,    total:  4451, status: '已成交', pnl:  2100 },
    { key: '5', date: '2024-03-18 11:20', asset: '招商银行 (600036)',    type: '卖出', price: 32.1,   amount: 1000, total: 32100, status: '已成交', pnl:  -120 },
  ];

  const columns: ColumnsType<TradeRecord> = [
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>交易时间</Text>,
      dataIndex: 'date',
      sorter: (a, b) => a.date.localeCompare(b.date),
      render: (text) => <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 13 }}>{text}</Text>,
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>标的名称</Text>,
      dataIndex: 'asset',
      render: (text) => <Text strong style={{ color: '#fff' }}>{text}</Text>,
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>操作类型</Text>,
      dataIndex: 'type',
      render: (type: '买入' | '卖出') => (
        <Tag
          icon={type === '买入' ? <RiseOutlined /> : <FallOutlined />}
          color={type === '买入' ? 'processing' : 'default'}   // 蓝色/灰色，去掉橙青
          style={{ borderRadius: 20, padding: '2px 10px' }}
        >
          {type}
        </Tag>
      ),
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>成交均价</Text>,
      dataIndex: 'price',
      render: (val) => <Text style={{ color: 'rgba(255,255,255,0.65)' }}>¥{val.toLocaleString()}</Text>,
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>成交数量</Text>,
      dataIndex: 'amount',
      render: (val) => <Text style={{ color: 'rgba(255,255,255,0.65)' }}>{val.toLocaleString()}</Text>,
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>成交额</Text>,
      dataIndex: 'total',
      render: (val) => <Text strong style={{ color: '#fff' }}>¥{val.toLocaleString()}</Text>,
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>浮动盈亏</Text>,
      dataIndex: 'pnl',
      render: (val: number) => (
        <Text strong style={{ color: val >= 0 ? '#52c41a' : '#ff4d4f' }}>
          {val >= 0 ? '+' : ''}¥{val.toLocaleString()}
        </Text>
      ),
    },
    {
      title: <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>状态</Text>,
      dataIndex: 'status',
      render: (status) => (
        <Tag style={{ borderRadius: 20, background: 'rgba(255,255,255,0.07)', border: 'none', color: 'rgba(255,255,255,0.45)' }}>
          {status}
        </Tag>
      ),
    },
  ];

  // 汇总卡：蓝/绿/红，对齐 Dashboard
  const summaryStats = [
    { label: '总交易次数', value: '26',      color: '#1677ff' },
    { label: '总成交额',   value: '¥89,051', color: '#fff'    },
    { label: '累计盈亏',   value: '+¥3,700', color: '#52c41a' },
    { label: '月度胜率',   value: '61%',     color: '#1677ff' },
  ];

  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <div style={{ minHeight: '100vh', background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)' }}>
        <PageHeader title="历史交易归档" subtitle="完整交易流水 · 盈亏追踪" />

        <div style={{ padding: '28px 32px', maxWidth: 1400, margin: '0 auto' }}>
          {/* 汇总统计 */}
          <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
            {summaryStats.map(item => (
              <Col xs={12} sm={6} key={item.label}>
                <Card bordered={false} style={{ ...cardStyle, textAlign: 'center' }}>
                  <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12, display: 'block' }}>{item.label}</Text>
                  <div style={{ color: item.color, fontSize: 22, fontWeight: 700, marginTop: 6 }}>{item.value}</div>
                </Card>
              </Col>
            ))}
          </Row>

          {/* 筛选栏 */}
          <Card bordered={false} style={{ ...cardStyle, marginBottom: 20 }}>
            <Space wrap size="middle">
              <Input
                placeholder="搜索标的名称 / 代码"
                prefix={<SearchOutlined style={{ color: 'rgba(255,255,255,0.3)' }} />}
                style={{ width: 260, background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 10 }}
              />
              <RangePicker
                style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 10 }}
              />
              <Button type="primary" icon={<SearchOutlined />} style={{ borderRadius: 10 }}>查询</Button>
              <Button icon={<DownloadOutlined />}
                style={{ borderRadius: 10, background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)' }}
              >
                导出 CSV
              </Button>
            </Space>
          </Card>

          {/* 表格 */}
          <Card
            bordered={false}
            style={cardStyle}
            title={<Space><HistoryOutlined style={{ color: '#1677ff' }} /><Text style={{ color: 'rgba(255,255,255,0.85)' }}>交易明细流水</Text></Space>}
          >
            <Table
              columns={columns}
              dataSource={mockData}
              loading={loading}
              pagination={{ pageSize: 10 }}
              size="middle"
              style={{ background: 'transparent' }}
            />
          </Card>
        </div>
      </div>
    </ConfigProvider>
  );
}