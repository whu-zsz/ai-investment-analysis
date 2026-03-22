// src/pages/Dashboard.tsx
import React from 'react';
import { Card, Row, Col, Statistic } from 'antd';

export default function Dashboard() {
  return (
    // 只能有一个普通的 div，绝对不能套 Layout
    <div style={{ padding: '0' }}> 
      <Row gutter={16}>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic title="组合总资产" value={428560} precision={2} />
          </Card>
        </Col>
        {/* 其他统计卡片... */}
      </Row>
      <Card title="收益趋势" style={{ marginTop: 24 }}>
        <p>这里放图表内容</p>
      </Card>
    </div>
  );
}