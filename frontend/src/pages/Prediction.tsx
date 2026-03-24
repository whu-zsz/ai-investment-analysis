import { useState, useEffect } from 'react';
import { Card, Row, Col, Typography, Statistic, Slider, Space, Tag, Alert, Button, Spin } from 'antd';
import { 
  LineChartOutlined, 
  ThunderboltOutlined, 
  InfoCircleOutlined,
  PlayCircleOutlined,
  StockOutlined
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

const { Title, Text, Paragraph } = Typography;

export default function PredictionPage() {
  const [loading, setLoading] = useState(true);
  const [riskLevel, setRiskLevel] = useState(50); // 预测风险偏好调节

  // 模拟 AI 模型运算
  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 1200);
    return () => clearTimeout(timer);
  }, []);

  // 生成预测曲线的数据
  const getOption = (): EChartsOption => {
    // 历史数据（实线）
    const historyData = [1.0, 1.02, 1.01, 1.05, 1.08, 1.07, 1.12];
    // 预测数据（虚线部分，根据 riskLevel 动态调整偏移）
    const predictionOffset = (riskLevel - 50) / 500;
    const predictMid = [1.12, 1.15 + predictionOffset, 1.18 + predictionOffset * 2, 1.22 + predictionOffset * 3];
    const predictUpper = [1.12, 1.18 + predictionOffset, 1.25 + predictionOffset * 2, 1.35 + predictionOffset * 4];
    const predictLower = [1.12, 1.11 + predictionOffset, 1.08 + predictionOffset, 1.05 + predictionOffset];

    return {
      tooltip: { trigger: 'axis' },
      legend: { data: ['历史净值', 'AI 预测路径', '乐观上限', '悲观下限'], bottom: 0 },
      grid: { top: '10%', left: '3%', right: '4%', bottom: '15%', containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '预测Q3', '预测Q4', '预测Y1', '预测Y2']
      },
      yAxis: { 
        type: 'value', 
        scale: true,
        axisLabel: { formatter: '{value} px' } 
      },
      series: [
        {
          name: '历史净值',
          type: 'line',
          data: historyData,
          smooth: true,
          lineStyle: { width: 4, color: '#1677ff' },
          symbol: 'none'
        },
        {
          name: 'AI 预测路径',
          type: 'line',
          data: [...Array(6).fill(null), ...predictMid],
          smooth: true,
          lineStyle: { type: 'dashed', width: 3, color: '#52c41a' },
          symbol: 'circle'
        },
        {
          name: '乐观上限',
          type: 'line',
          data: [...Array(6).fill(null), ...predictUpper],
          smooth: true,
          lineStyle: { width: 1, type: 'dotted', color: '#ff4d4f' },
          areaStyle: { color: 'rgba(255, 77, 79, 0.1)' },
          symbol: 'none'
        },
        {
          name: '悲观下限',
          type: 'line',
          data: [...Array(6).fill(null), ...predictLower],
          smooth: true,
          lineStyle: { width: 1, type: 'dotted', color: '#8c8c8c' },
          areaStyle: { color: 'rgba(140, 140, 140, 0.1)' },
          symbol: 'none'
        }
      ]
    };
  };

  return (
    <div style={{ padding: '4px' }}>
      <Title level={3}>AI 收益趋势预测</Title>
      <Paragraph type="secondary">
        基于 Monte Carlo 模拟算法，结合当前市场 Beta 系数，为您生成未来 24 个月的资产走势预演。
      </Paragraph>

      <Row gutter={[16, 16]}>
        {/* 控制面板 */}
        <Col span={24} lg={6}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card title={<span><ThunderboltOutlined /> 预测参数配置</span>} bordered={false}>
              <div style={{ marginBottom: 20 }}>
                <Text strong>预期风险因子权重</Text>
                <Slider 
                  value={riskLevel} 
                  onChange={(val) => setRiskLevel(val)} 
                  marks={{ 0: '保守', 50: '平衡', 100: '激进' }}
                />
              </div>
              <Button type="primary" block icon={<PlayCircleOutlined />} onClick={() => {
                setLoading(true);
                setTimeout(() => setLoading(false), 800);
              }}>
                重新跑数
              </Button>
            </Card>

            <Card bordered={false}>
              <Statistic 
                title="预期年化回报率" 
                value={12.8} 
                suffix="%" 
                precision={2}
                valueStyle={{ color: '#52c41a' }}
                prefix={<StockOutlined />}
              />
              <Tag color="green" style={{ marginTop: 8 }}>优于 92% 的同类组合</Tag>
            </Card>

            <Alert
              message="风险提示"
              description="预测结果仅供参考，不构成投资建议。模拟路径基于历史波动率计算。"
              type="warning"
              showIcon
              icon={<InfoCircleOutlined />}
            />
          </Space>
        </Col>

        {/* 图表展示区 */}
        <Col span={24} lg={18}>
          <Card 
            title={<span><LineChartOutlined /> 资产净值演变模拟 (未来24个月)</span>} 
            bordered={false}
            extra={<Text type="secondary">当前模型版本: V2.4-DeepSeek-Enhanced</Text>}
          >
            <Spin spinning={loading} tip="AI 模型深度运算中...">
              <ReactECharts option={getOption()} style={{ height: '500px' }} />
            </Spin>
          </Card>
        </Col>
      </Row>
    </div>
  );
}