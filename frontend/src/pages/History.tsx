import React from 'react';
import { Table, Card, Tag } from 'antd';

export default function History() {
  const columns = [
    { title: '日期', dataIndex: 'date', key: 'date' },
    { title: '类型', dataIndex: 'type', key: 'type' },
    { title: '资产', dataIndex: 'asset', key: 'asset' },
    { title: '盈亏', dataIndex: 'profit', key: 'profit', 
      render: (val: string) => <span style={{ color: val.startsWith('+') ? 'green' : 'red' }}>{val}</span> 
    },
  ];

  const data = [
    { key: '1', date: '2026-03-15', type: '买入', asset: '腾讯控股', profit: '+¥890' },
    { key: '2', date: '2026-03-10', type: '卖出', asset: '贵州茅台', profit: '-¥120' },
  ];

  return (
    <Card title="📜 历史交易记录">
      <Table columns={columns} dataSource={data} pagination={{ pageSize: 5 }} />
    </Card>
  );
}