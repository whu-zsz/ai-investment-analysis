import { Card, Typography, Alert } from 'antd';
import ReactECharts from 'echarts-for-react';

export default function Prediction() {
  const option = {
    title: { text: '未来 30 天收益模拟 (AI)', left: 'center' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['W1', 'W2', 'W3', 'W4'] },
    yAxis: { type: 'value' },
    series: [{
      name: '预测区间',
      type: 'line',
      smooth: true,
      data: [310, 330, 325, 350],
      lineStyle: { type: 'dashed', color: '#1890ff' },
      areaStyle: { color: 'rgba(24, 144, 255, 0.1)' }
    }]
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card bordered={false}>
        <Typography.Title level={2}>AI 趋势模拟</Typography.Title>
        <Alert message="基于历史数据生成的模拟趋势，不构成具体投资建议。" type="info" showIcon style={{ marginBottom: 24 }} />
        <ReactECharts option={option} style={{ height: '400px' }} />
      </Card>
    </div>
  );
}