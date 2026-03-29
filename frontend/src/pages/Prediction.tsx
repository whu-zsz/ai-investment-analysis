import { useState, useEffect } from 'react';
import {
  Card, Row, Col, Typography, Statistic, Slider, Space,
  Tag, Alert, Button, Spin
} from 'antd';
import {
  LineChartOutlined, ThunderboltOutlined, InfoCircleOutlined,
  PlayCircleOutlined, StockOutlined, ArrowLeftOutlined, RiseOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

const { Title, Text, Paragraph } = Typography;

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const scenarios = [
  { label: '乐观情景 (P90)', value: '+34.2%', color: '#52c41a', bg: '#f6ffed' },
  { label: '基准情景 (P50)', value: '+12.8%', color: '#1677ff', bg: '#e6f4ff' },
  { label: '悲观情景 (P10)', value: '-5.1%',  color: '#ff4d4f', bg: '#fff1f0' },
];

const modelStats = [
  { label: '模拟路径数', value: '10,000', unit: '条',   color: '#1677ff', bg: '#e6f4ff' },
  { label: '置信区间',   value: '95',      unit: '%',    color: '#1677ff', bg: '#e6f4ff' },
  { label: '波动率假设', value: '18.4',    unit: '%/年', color: '#ff4d4f', bg: '#fff1f0' },
  { label: '相关性系数', value: '0.72',    unit: '',     color: '#52c41a', bg: '#f6ffed' },
];

export default function PredictionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [riskLevel, setRiskLevel] = useState(50);

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 1200);
    return () => clearTimeout(timer);
  }, []);

  const getOption = (): EChartsOption => {
    const historyData = [1.0, 1.02, 1.01, 1.05, 1.08, 1.07, 1.12];
    const offset = (riskLevel - 50) / 500;
    const predictMid   = [1.12, 1.15 + offset,      1.18 + offset * 2,      1.22 + offset * 3];
    const predictUpper = [1.12, 1.18 + offset,      1.25 + offset * 2,      1.35 + offset * 4];
    const predictLower = [1.12, 1.11 + offset,      1.08 + offset,           1.05 + offset];

    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.96)',
        borderColor: '#d9e6ff',
        borderWidth: 1,
      },
      legend: {
        data: ['历史净值', 'AI 预测路径', '乐观上限', '悲观下限'],
        bottom: 0,
        textStyle: { color: '#595959' },
      },
      grid: { top: '8%', left: '3%', right: '4%', bottom: '14%', containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['1月','2月','3月','4月','5月','6月','7月','预测Q3','预测Q4','预测Y1','预测Y2'],
        axisLine: { lineStyle: { color: '#d9d9d9' } },
        axisLabel: { color: '#8c8c8c' },
      },
      yAxis: {
        type: 'value',
        scale: true,
        splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } },
        axisLabel: { color: '#8c8c8c', formatter: '{value}x' },
      },
      series: [
        { name: '历史净值',   type: 'line', data: historyData, smooth: true, lineStyle: { width: 3, color: '#1677ff' }, areaStyle: { color: { type: 'linear', x:0,y:0,x2:0,y2:1, colorStops:[{offset:0,color:'rgba(22,119,255,0.3)'},{offset:1,color:'rgba(22,119,255,0.02)'}] } }, symbol: 'none' },
        { name: 'AI 预测路径', type: 'line', data: [...Array(6).fill(null), ...predictMid],   smooth: true, lineStyle: { type: 'dashed', width: 2, color: '#52c41a' }, symbol: 'circle', symbolSize: 6 },
        { name: '乐观上限',   type: 'line', data: [...Array(6).fill(null), ...predictUpper], smooth: true, lineStyle: { width: 1, type: 'dotted', color: '#1677ff' }, areaStyle: { color: 'rgba(22,119,255,0.06)' }, symbol: 'none' },
        { name: '悲观下限',   type: 'line', data: [...Array(6).fill(null), ...predictLower], smooth: true, lineStyle: { width: 1, type: 'dotted', color: '#ff4d4f' }, areaStyle: { color: 'rgba(255,77,79,0.05)' }, symbol: 'none' },
      ],
    };
  };

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
              <Tag color="processing">Monte Carlo 模拟</Tag>
              <Tag color="blue">24 个月预演</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 收益趋势预测</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              基于 Monte Carlo 模拟算法，结合当前市场 Beta 系数，为您生成未来 24 个月的资产走势预演。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              预期年化 +12.8%
            </Tag>
            <Tag color="processing" icon={<StockOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              优于 92% 同类组合
            </Tag>
          </Space>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        {/* ── 左侧控制面板 ── */}
        <Col span={24} lg={6}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>

            <Card
              bordered={false}
              style={cardStyle}
              title={<span><ThunderboltOutlined style={{ color: '#1677ff', marginRight: 8 }} />预测参数配置</span>}
            >
              <Text type="secondary" style={{ fontSize: 13 }}>预期风险因子权重</Text>
              <Slider
                value={riskLevel}
                onChange={setRiskLevel}
                marks={{
                  0:   <Text type="secondary" style={{ fontSize: 11 }}>保守</Text>,
                  50:  <Text type="secondary" style={{ fontSize: 11 }}>平衡</Text>,
                  100: <Text type="secondary" style={{ fontSize: 11 }}>激进</Text>,
                }}
                style={{ margin: '12px 0 24px' }}
              />
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
                title="预期年化回报率"
                value={12.8}
                suffix="%"
                precision={2}
                valueStyle={{ color: '#52c41a', fontSize: 34 }}
                prefix={<StockOutlined />}
              />
              <Tag color="success" style={{ marginTop: 10, borderRadius: 20, padding: '2px 12px' }}>
                优于 92% 的同类组合
              </Tag>
            </Card>

            {/* 情景区间 */}
            <Card
              bordered={false}
              style={cardStyle}
              title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />情景模拟区间</span>}
            >
              {scenarios.map((item, i) => (
                <div key={item.label} style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '10px 0',
                  borderBottom: i < scenarios.length - 1 ? '1px solid #f0f0f0' : 'none',
                }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                  <Tag color={item.color === '#52c41a' ? 'success' : item.color === '#1677ff' ? 'processing' : 'error'}
                    style={{ borderRadius: 20, fontWeight: 700 }}>
                    {item.value}
                  </Tag>
                </div>
              ))}
            </Card>

            <Alert
              message="风险提示"
              description="预测结果仅供参考，不构成投资建议。模拟路径基于历史波动率计算。"
              type="warning"
              showIcon
              icon={<InfoCircleOutlined />}
              style={{ borderRadius: 12 }}
            />
          </Space>
        </Col>

        {/* ── 右侧图表 ── */}
        <Col span={24} lg={18}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card
              bordered={false}
              style={cardStyle}
              title={<span><LineChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />资产净值演变模拟（未来 24 个月）</span>}
              extra={<Text type="secondary" style={{ fontSize: 12 }}>模型版本: V2.4-DeepSeek-Enhanced</Text>}
            >
              <Spin spinning={loading} tip="AI 模型深度运算中...">
                <ReactECharts option={getOption()} style={{ height: '460px' }} />
              </Spin>
            </Card>

            {/* 模型参数卡 */}
            <Row gutter={[12, 12]}>
              {modelStats.map(item => (
                <Col span={6} key={item.label}>
                  <Card bordered={false} style={{ ...cardStyle, textAlign: 'center' }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                    <div style={{ marginTop: 6 }}>
                      <span style={{ color: item.color, fontSize: 22, fontWeight: 700 }}>{item.value}</span>
                      {item.unit && <span style={{ color: '#bfbfbf', fontSize: 12, marginLeft: 3 }}>{item.unit}</span>}
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          </Space>
        </Col>
      </Row>
    </div>
  );
}