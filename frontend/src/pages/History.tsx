import { useState } from 'react';
import {
  Table, Card, Typography, Tag, Input, Space, Button,
  DatePicker, Row, Col, Statistic, Alert
} from 'antd';
import {
  SearchOutlined, DownloadOutlined, HistoryOutlined,
  RiseOutlined, FallOutlined, ArrowLeftOutlined, BulbOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';

const { Text, Title, Paragraph } = Typography;
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

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const mockData: TradeRecord[] = [
  { key: '1', date: '2024-03-22 14:30', asset: '腾讯控股 (0700.HK)',   type: '买入', price: 290.5,  amount: 100,  total: 29050, status: '已成交', pnl:  1240 },
  { key: '2', date: '2024-03-21 10:15', asset: '贵州茅台 (600519)',    type: '卖出', price: 1720.0, amount: 10,   total: 17200, status: '已成交', pnl:  -380 },
  { key: '3', date: '2024-03-20 09:45', asset: '纳指100ETF (513100)', type: '买入', price: 1.25,   amount: 5000, total:  6250, status: '已成交', pnl:   860 },
  { key: '4', date: '2024-03-19 15:00', asset: '英伟达 (NVDA.US)',     type: '买入', price: 890.2,  amount: 5,    total:  4451, status: '已成交', pnl:  2100 },
  { key: '5', date: '2024-03-18 11:20', asset: '招商银行 (600036)',    type: '卖出', price: 32.1,   amount: 1000, total: 32100, status: '已成交', pnl:  -120 },
];

const summaryStats = [
  { label: '总交易次数', value: 26,      suffix: '次',  color: '#1677ff', bg: '#e6f4ff' },
  { label: '总成交额',   value: 89051,   suffix: '元',  color: '#262626', bg: '#f8fafc' },
  { label: '累计盈亏',   value: 3700,    suffix: '元',  color: '#52c41a', bg: '#f6ffed' },
  { label: '月度胜率',   value: 61,      suffix: '%',   color: '#1677ff', bg: '#e6f4ff' },
];

export default function HistoryPage() {
  const navigate = useNavigate();
  const [loading] = useState(false);

  const columns: ColumnsType<TradeRecord> = [
    {
      title: '交易时间', dataIndex: 'date',
      sorter: (a, b) => a.date.localeCompare(b.date),
      render: (text) => <Text type="secondary" style={{ fontSize: 13 }}>{text}</Text>,
    },
    {
      title: '标的名称', dataIndex: 'asset',
      render: (text) => <Text strong>{text}</Text>,
    },
    {
      title: '操作类型', dataIndex: 'type',
      render: (type: '买入' | '卖出') => (
        <Tag
          icon={type === '买入' ? <RiseOutlined /> : <FallOutlined />}
          color={type === '买入' ? 'processing' : 'default'}
          style={{ borderRadius: 20, padding: '2px 10px' }}
        >
          {type}
        </Tag>
      ),
    },
    {
      title: '成交均价', dataIndex: 'price',
      render: (val) => `¥${val.toLocaleString()}`,
    },
    {
      title: '成交数量', dataIndex: 'amount',
      render: (val) => val.toLocaleString(),
    },
    {
      title: '成交额', dataIndex: 'total',
      render: (val) => <Text strong>¥{val.toLocaleString()}</Text>,
    },
    {
      title: '浮动盈亏', dataIndex: 'pnl',
      render: (val: number) => (
        <Text strong style={{ color: val >= 0 ? '#52c41a' : '#ff4d4f' }}>
          {val >= 0 ? '+' : ''}¥{val.toLocaleString()}
        </Text>
      ),
    },
    {
      title: '状态', dataIndex: 'status',
      render: (status) => <Tag style={{ borderRadius: 20 }}>{status}</Tag>,
    },
  ];

  return (
    <div style={{ padding: '24px' }}>

      {/* 返回按钮 */}
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}
      >
        返回首页
      </Button>

      {/* Hero Banner */}
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
              <Tag color="blue">盈亏追踪</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>历史交易归档</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              完整的交易流水记录，自动计算每笔浮动盈亏，支持按标的、日期灵活筛选与导出。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              累计盈亏 +¥3,700
            </Tag>
            <Tag color="processing" icon={<HistoryOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              近 30 日 26 笔
            </Tag>
          </Space>
        </div>
      </Card>

      {/* 汇总统计 */}
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

      {/* 筛选栏 */}
      <Card bordered={false} style={{ ...cardStyle, marginBottom: 16 }}>
        <Space wrap size="middle">
          <Input
            placeholder="搜索标的名称 / 代码"
            prefix={<SearchOutlined />}
            style={{ width: 260, borderRadius: 10 }}
            allowClear
          />
          <RangePicker style={{ borderRadius: 10 }} />
          <Button type="primary" icon={<SearchOutlined />} style={{ borderRadius: 10 }}>查询</Button>
          <Button icon={<DownloadOutlined />} style={{ borderRadius: 10 }}>导出 CSV</Button>
        </Space>
      </Card>

      {/* 交易明细表格 */}
      <Card
        bordered={false}
        style={cardStyle}
        title={<span><HistoryOutlined style={{ color: '#1677ff', marginRight: 8 }} />交易明细流水</span>}
      >
        <Table
          columns={columns}
          dataSource={mockData}
          loading={loading}
          pagination={{ pageSize: 10 }}
          size="middle"
        />
      </Card>

      {/* 底部说明 Alert */}
      <Card bordered={false} style={{ ...cardStyle, marginTop: 16 }}>
        <Alert
          type="info"
          showIcon
          icon={<BulbOutlined />}
          message="AI 一句话结论：近 30 日胜率 61%，盈亏比 1.47，策略整体有效但高频换手侵蚀了部分收益。"
          description={
            <Space direction="vertical" size={4}>
              <Text type="secondary">平均持仓 11 天，短线切换偏多，建议拉长持股周期。</Text>
              <Text type="secondary">本月摩擦成本约 0.8%，可通过减少无效换手来提升净收益。</Text>
            </Space>
          }
        />
      </Card>
    </div>
  );
}