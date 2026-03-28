import { useState, useEffect } from 'react';
import { Card, Row, Col, Typography, Statistic, Slider, Space, Alert, Button, Spin, ConfigProvider, theme } from 'antd';
import {
  LineChartOutlined,
  ThunderboltOutlined,
  InfoCircleOutlined,
  PlayCircleOutlined,
  StockOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import PageHeader from '../components/PageHeader';

const { Text } = Typography;

const cardStyle = {
  background: 'rgba(15, 23, 42, 0.6)',
  border: '1px solid rgba(255,255,255,0.08)',
  borderRadius: 16,
  backdropFilter: 'blur(10px)',
};

export default function PredictionPage() {
  const [loading, setLoading] = useState(true);
  const [riskLevel, setRiskLevel] = useState(50);

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 1200);
    return () => clearTimeout(timer);
  }, []);

  const getOption = (): EChartsOption => {
    const historyData = [1.0, 1.02, 1.01, 1.05, 1.08, 1.07, 1.12];
    const offset = (riskLevel - 50) / 500;
    const predictMid   = [1.12, 1.15 + offset,       1.18 + offset * 2,       1.22 + offset * 3];
    const predictUpper = [1.12, 1.18 + offset,       1.25 + offset * 2,       1.35 + offset * 4];
    const predictLower = [1.12, 1.11 + offset,       1.08 + offset,            1.05 + offset];

    return {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(15,23,42,0.92)',
        borderColor: 'rgba(22,119,255,0.3)',
        textStyle: { color: '#fff' },
      },
      legend: {
        data: ['历史净值', 'AI 预测路径', '乐观上限', '悲观下限'],
        bottom: 0,
        textStyle: { color: 'rgba(255,255,255,0.5)' },
      },
      grid: { top: '8%', left: '3%', right: '4%', bottom: '14%', containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '预测Q3', '预测Q4', '预测Y1', '预测Y2'],
        axisLine: { lineStyle: { color: 'rgba(255,255,255,0.12)' } },
        axisLabel: { color: 'rgba(255,255,255,0.4)' },
      },
      yAxis: {
        type: 'value',
        scale: true,
        splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)', type: 'dashed' } },
        axisLabel: { color: 'rgba(255,255,255,0.4)', formatter: '{value}x' },
      },
      series: [
        {
          name: '历史净值',
          type: 'line',
          data: historyData,
          smooth: true,
          lineStyle: { width: 3, color: '#1677ff' },
          areaStyle: {
            color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [{ offset: 0, color: 'rgba(22,119,255,0.25)' }, { offset: 1, color: 'rgba(22,119,255,0.02)' }] }
          },
          symbol: 'none',
        },
        {
          name: 'AI 预测路径',
          type: 'line',
          data: [...Array(6).fill(null), ...predictMid],
          smooth: true,
          lineStyle: { type: 'dashed', width: 2, color: '#52c41a' },
          symbol: 'circle',
          symbolSize: 6,
        },
        {
          name: '乐观上限',
          type: 'line',
          data: [...Array(6).fill(null), ...predictUpper],
          smooth: true,
          lineStyle: { width: 1, type: 'dotted', color: '#4096ff' },
          areaStyle: { color: 'rgba(64,150,255,0.06)' },
          symbol: 'none',
        },
        {
          name: '悲观下限',
          type: 'line',
          data: [...Array(6).fill(null), ...predictLower],
          smooth: true,
          lineStyle: { width: 1, type: 'dotted', color: '#ff4d4f' },
          areaStyle: { color: 'rgba(255,77,79,0.05)' },
          symbol: 'none',
        },
      ],
    };
  };

  // 情景区间：绿/蓝/红，对齐 Dashboard
  const scenarios = [
    { label: '乐观情景 (P90)', value: '+34.2%', color: '#52c41a' },
    { label: '基准情景 (P50)', value: '+12.8%', color: '#1677ff' },
    { label: '悲观情景 (P10)', value: '-5.1%',  color: '#ff4d4f' },
  ];

  // 底部统计：全用蓝色系
  const modelStats = [
    { label: '模拟路径数',  value: '10,000', unit: '条',  color: '#1677ff' },
    { label: '置信区间',    value: '95',      unit: '%',   color: '#4096ff' },
    { label: '波动率假设',  value: '18.4',    unit: '%/年',color: '#69b1ff' },
    { label: '相关性系数',  value: '0.72',    unit: '',    color: '#1677ff' },
  ];

  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <div style={{ minHeight: '100vh', background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)' }}>
        <PageHeader title="AI 收益趋势预测" subtitle="Monte Carlo 模拟算法 · 未来 24 个月资产走势预演" />

        <div style={{ padding: '28px 32px', maxWidth: 1400, margin: '0 auto' }}>
          <Row gutter={[20, 20]}>
            {/* ── 左侧控制面板 ── */}
            <Col span={24} lg={6}>
              <Space direction="vertical" style={{ width: '100%' }} size={20}>

                <Card
                  title={<Space><ThunderboltOutlined style={{ color: '#1677ff' }} /><Text style={{ color: 'rgba(255,255,255,0.85)' }}>预测参数配置</Text></Space>}
                  bordered={false}
                  style={cardStyle}
                >
                  <div style={{ marginBottom: 24 }}>
                    <Text style={{ color: 'rgba(255,255,255,0.55)', fontSize: 13 }}>预期风险因子权重</Text>
                    <Slider
                      value={riskLevel}
                      onChange={setRiskLevel}
                      marks={{
                        0:   <span style={{ color: 'rgba(255,255,255,0.35)', fontSize: 11 }}>保守</span>,
                        50:  <span style={{ color: 'rgba(255,255,255,0.35)', fontSize: 11 }}>平衡</span>,
                        100: <span style={{ color: 'rgba(255,255,255,0.35)', fontSize: 11 }}>激进</span>,
                      }}
                      style={{ marginTop: 8 }}
                    />
                  </div>
                  <Button
                    type="primary"
                    block
                    icon={<PlayCircleOutlined />}
                    onClick={() => { setLoading(true); setTimeout(() => setLoading(false), 800); }}
                    style={{ borderRadius: 10, height: 42 }}
                  >
                    重新跑数
                  </Button>
                </Card>

                <Card bordered={false} style={cardStyle}>
                  <Statistic
                    title={<Text style={{ color: 'rgba(255,255,255,0.5)' }}>预期年化回报率</Text>}
                    value={12.8}
                    suffix="%"
                    precision={2}
                    valueStyle={{ color: '#52c41a', fontSize: 36 }}
                    prefix={<StockOutlined />}
                  />
                  <div style={{
                    marginTop: 10,
                    display: 'inline-block',
                    background: 'rgba(82,196,26,0.1)',
                    border: '1px solid rgba(82,196,26,0.25)',
                    borderRadius: 20,
                    padding: '2px 14px',
                    fontSize: 12,
                    color: '#52c41a',
                  }}>
                    优于 92% 的同类组合
                  </div>
                </Card>

                <Alert
                  message={<Text style={{ color: 'rgba(255,255,255,0.8)', fontWeight: 500 }}>风险提示</Text>}
                  description={<Text style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12, lineHeight: 1.6 }}>预测结果仅供参考，不构成投资建议。模拟路径基于历史波动率计算。</Text>}
                  type="warning"
                  showIcon
                  icon={<InfoCircleOutlined />}
                  style={{
                    background: 'rgba(255,77,79,0.07)',
                    border: '1px solid rgba(255,77,79,0.2)',
                    borderRadius: 12,
                  }}
                />

                {/* 情景区间 */}
                <Card bordered={false} style={cardStyle}>
                  <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12, display: 'block', marginBottom: 12 }}>情景模拟区间</Text>
                  {scenarios.map(item => (
                    <div key={item.label} style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      padding: '10px 0',
                      borderBottom: '1px solid rgba(255,255,255,0.06)',
                    }}>
                      <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>{item.label}</Text>
                      <Text style={{ color: item.color, fontWeight: 700 }}>{item.value}</Text>
                    </div>
                  ))}
                </Card>
              </Space>
            </Col>

            {/* ── 右侧图表 ── */}
            <Col span={24} lg={18}>
              <Card
                title={<Space><LineChartOutlined style={{ color: '#1677ff' }} /><Text style={{ color: 'rgba(255,255,255,0.85)' }}>资产净值演变模拟 (未来24个月)</Text></Space>}
                bordered={false}
                style={cardStyle}
                extra={<Text style={{ color: 'rgba(255,255,255,0.25)', fontSize: 12 }}>模型版本: V2.4-DeepSeek-Enhanced</Text>}
              >
                <Spin spinning={loading} tip="AI 模型深度运算中...">
                  <ReactECharts option={getOption()} style={{ height: '480px' }} />
                </Spin>
              </Card>

              {/* 底部模型参数 */}
              <Row gutter={[16, 16]} style={{ marginTop: 20 }}>
                {modelStats.map(item => (
                  <Col span={6} key={item.label}>
                    <Card bordered={false} style={{ ...cardStyle, textAlign: 'center' }}>
                      <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12, display: 'block' }}>{item.label}</Text>
                      <div style={{ marginTop: 6 }}>
                        <span style={{ color: item.color, fontSize: 22, fontWeight: 700 }}>{item.value}</span>
                        {item.unit && <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: 12, marginLeft: 4 }}>{item.unit}</span>}
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            </Col>
          </Row>
        </div>
      </div>
    </ConfigProvider>
  );
}