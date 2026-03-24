import { useState } from 'react';
import { Table, Card, Typography, Tag, Input, Space, Button, DatePicker } from 'antd';
import { SearchOutlined, DownloadOutlined, HistoryOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

// 定义交易记录类型
interface TradeRecord {
  key: string;
  date: string;
  asset: string;
  type: '买入' | '卖出';
  price: number;
  amount: number;
  total: number;
  status: '已成交' | '已撤单';
}

export default function HistoryPage() {
  const [loading] = useState(false);

  // 模拟历史数据
  const mockData: TradeRecord[] = [
    { key: '1', date: '2024-03-22 14:30', asset: '腾讯控股 (0700.HK)', type: '买入', price: 290.5, amount: 100, total: 29050, status: '已成交' },
    { key: '2', date: '2024-03-21 10:15', asset: '贵州茅台 (600519)', type: '卖出', price: 1720.0, amount: 10, total: 17200, status: '已成交' },
    { key: '3', date: '2024-03-20 09:45', asset: '纳指100ETF (513100)', type: '买入', price: 1.25, amount: 5000, total: 6250, status: '已成交' },
    { key: '4', date: '2024-03-19 15:00', asset: '英伟达 (NVDA.US)', type: '买入', price: 890.2, amount: 5, total: 4451, status: '已成交' },
    { key: '5', date: '2024-03-18 11:20', asset: '招商银行 (600036)', type: '卖出', price: 32.1, amount: 1000, total: 32100, status: '已成交' },
  ];

  const columns: ColumnsType<TradeRecord> = [
    {
      title: '交易时间',
      dataIndex: 'date',
      key: 'date',
      sorter: (a, b) => a.date.localeCompare(b.date),
    },
    {
      title: '标的名称',
      dataIndex: 'asset',
      key: 'asset',
      render: (text) => <Text strong>{text}</Text>,
    },
    {
      title: '操作类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag color={type === '买入' ? 'volcano' : 'cyan'}>{type}</Tag>
      ),
    },
    {
      title: '成交均价',
      dataIndex: 'price',
      key: 'price',
      render: (val) => `¥${val.toLocaleString()}`,
    },
    {
      title: '成交数量',
      dataIndex: 'amount',
      key: 'amount',
    },
    {
      title: '成交额',
      dataIndex: 'total',
      key: 'total',
      render: (val) => <Text strong>¥${val.toLocaleString()}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => <Tag color="default">{status}</Tag>,
    },
  ];

  return (
    <div style={{ padding: '8px' }}>
      <Title level={3}>历史交易归档</Title>
      
      <Card bordered={false} style={{ marginBottom: 16 }}>
        <Space wrap size="middle">
          <Input placeholder="搜索标的名称/代码" prefix={<SearchOutlined />} style={{ width: 250 }} />
          <RangePicker />
          <Button type="primary" icon={<SearchOutlined />}>查询</Button>
          <Button icon={<DownloadOutlined />}>导出 CSV</Button>
        </Space>
      </Card>

      <Card 
        bordered={false} 
        title={<span><HistoryOutlined /> 交易明细流水</span>}
      >
        <Table 
          columns={columns} 
          dataSource={mockData} 
          loading={loading}
          pagination={{ pageSize: 10 }}
          size="middle"
        />
      </Card>
    </div>
  );
}