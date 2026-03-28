import { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Typography, Divider, Tag, Descriptions, Space, Skeleton, Progress, ConfigProvider, theme } from 'antd';
import {
  RadarChartOutlined,
  SafetyCertificateOutlined,
  BulbOutlined,
  InfoCircleOutlined
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import PageHeader from '../components/PageHeader';

const { Title, Paragraph, Text } = Typography;

const cardStyle = {
  background: 'rgba(15, 23, 42, 0.6)',
  border: '1px solid rgba(255,255,255,0.08)',
  borderRadius: 16,
  backdropFilter: 'blur(10px)',
};

export default function AnalysisPage() {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 1200);
    return () => clearTimeout(timer);
  }, []);

  const getRadarOption = (): EChartsOption => ({
    radar: {
      indicator: [
        { name: '收益爆发力', max: 100 },
        { name: '回撤控制', max: 100 },
        { name: '资产分散度', max: 100 },
        { name: '交易纪律', max: 100 },
        { name: '风格稳定性', max: 100 },
      ],
      shape: 'circle',
      splitNumber: 5,
      axisName: { color: 'rgba(255,255,255,0.45)' },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.07)' } },
      splitArea: { show: false },
    },
    backgroundColor: 'transparent',
    series: [{
      type: 'radar',
      data: [{
        value: [82, 45, 30, 65, 90],
        name: '特征评分',
        itemStyle: { color: '#1677ff' },
        lineStyle: { color: '#1677ff' },
        areaStyle: { color: 'rgba(22, 119, 255, 0.2)' },
      }],
    }],
  });

  // 指标卡：全用蓝色系
  const quickMetrics = [
    { label: 'Beta 系数',  value: '1.42', color: '#4096ff' },
    { label: '月换手率',   value: '120%', color: '#ff4d4f' },   // 红色代表风险，Dashboard 也用
    { label: '夏普比率',   value: '0.87', color: '#1677ff' },
    { label: '最大回撤',   value: '4.9%', color: '#ff4d4f' },
  ];

  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <div style={{ minHeight: '100vh', background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)' }}>
        <PageHeader title="AI 深度风险诊断" subtitle="基于多因子量化模型，深度穿透您的历史交易行为" />

        <div style={{ padding: '28px 32px', maxWidth: 1400, margin: '0 auto' }}>
          {loading ? (
            <Skeleton active paragraph={{ rows: 12 }} />
          ) : (
            <Row gutter={[20, 20]}>
              {/* ── 左栏 ── */}
              <Col span={24} lg={8}>
                <Space direction="vertical" style={{ width: '100%' }} size={20}>

                  {/* 健康分 */}
                  <Card bordered={false} style={cardStyle}>
                    <Statistic
                      title={<Text style={{ color: 'rgba(255,255,255,0.5)' }}>账户健康分</Text>}
                      value={74.2}
                      suffix="/ 100"
                      prefix={<SafetyCertificateOutlined />}
                      valueStyle={{ color: '#52c41a', fontSize: 36 }}
                    />
                    <Progress
                      percent={74.2}
                      showInfo={false}
                      strokeColor={{ '0%': '#52c41a', '100%': '#95de64' }}
                      trailColor="rgba(255,255,255,0.08)"
                      style={{ marginTop: 12 }}
                    />
                    <Text style={{ color: 'rgba(255,255,255,0.35)', fontSize: 12, marginTop: 6, display: 'block' }}>
                      高于 78% 的同类用户
                    </Text>
                  </Card>

                  {/* 雷达图 */}
                  <Card
                    title={<Space><RadarChartOutlined style={{ color: '#1677ff' }} /><Text style={{ color: 'rgba(255,255,255,0.85)' }}>投资风格画像</Text></Space>}
                    bordered={false}
                    style={cardStyle}
                  >
                    <ReactECharts option={getRadarOption()} style={{ height: 280 }} />
                    <div style={{ textAlign: 'center', marginTop: 8 }}>
                      <Tag color="blue" style={{ borderRadius: 20, padding: '2px 14px' }}>激进型成长风格</Tag>
                    </div>
                  </Card>

                  {/* 快速指标 */}
                  <Card bordered={false} style={cardStyle}>
                    <Row gutter={[12, 12]}>
                      {quickMetrics.map(item => (
                        <Col span={12} key={item.label}>
                          <div style={{
                            background: 'rgba(255,255,255,0.04)',
                            borderRadius: 10,
                            padding: '12px 14px',
                            border: '1px solid rgba(255,255,255,0.06)',
                          }}>
                            <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 11 }}>{item.label}</Text>
                            <div style={{ color: item.color, fontSize: 20, fontWeight: 700, marginTop: 4 }}>{item.value}</div>
                          </div>
                        </Col>
                      ))}
                    </Row>
                  </Card>
                </Space>
              </Col>

              {/* ── 右栏 ── */}
              <Col span={24} lg={16}>
                <Space direction="vertical" style={{ width: '100%' }} size={20}>

                  {/* AI 诊断 */}
                  <Card
                    bordered={false}
                    style={cardStyle}
                    title={<Space><BulbOutlined style={{ color: '#1677ff' }} /><Text style={{ color: 'rgba(255,255,255,0.85)' }}>AI 诊断结论</Text></Space>}
                  >
                    <Descriptions
                      column={1}
                      size="small"
                      labelStyle={{ color: 'rgba(255,255,255,0.4)', width: 110 }}
                      contentStyle={{ color: 'rgba(255,255,255,0.8)' }}
                    >
                      <Descriptions.Item label="潜在风险点">
                        <Text type="danger">持仓集中度过高。</Text>
                        <Text style={{ color: 'rgba(255,255,255,0.7)' }}> 您的前两大持仓占总资产 65%，极易受单一行业波动影响。</Text>
                      </Descriptions.Item>
                      <Descriptions.Item label="交易倾向">
                        <Text style={{ color: 'rgba(255,255,255,0.7)' }}>
                          检测到轻微的 <Text strong style={{ color: '#fff' }}>"处置效应"</Text>（倾向于过早卖出盈利股，而长期持有亏损股）。
                        </Text>
                      </Descriptions.Item>
                      <Descriptions.Item label="优化建议">
                        <Text style={{ color: 'rgba(255,255,255,0.7)' }}>
                          建议将科技板块仓位下调 15%，增配防御性资产如红利低波 ETF。
                        </Text>
                      </Descriptions.Item>
                    </Descriptions>
                  </Card>

                  {/* 行为特征报告 */}
                  <Card bordered={false} style={cardStyle}>
                    <Divider style={{ borderColor: 'rgba(255,255,255,0.08)', margin: '0 0 20px' }}>
                      <Space style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>
                        <InfoCircleOutlined />行为特征报告
                      </Space>
                    </Divider>

                    <div style={{
                      background: 'rgba(22,119,255,0.06)',
                      border: '1px solid rgba(22,119,255,0.18)',
                      padding: '20px 24px',
                      borderRadius: 12,
                      marginBottom: 16,
                    }}>
                      <Title level={5} style={{ color: '#fff', marginTop: 0 }}>⚡ 市场敏感度扫描</Title>
                      <Paragraph style={{ color: 'rgba(255,255,255,0.6)', marginBottom: 0, lineHeight: 1.8 }}>
                        您的组合 Beta 系数为 1.42，意味着市场每波动 1%，您的账户预期波动 1.42%。这表明您处于杠杆化配置状态，在牛市表现优异，但在宽幅震荡期可能面临较大压力。建议通过对冲工具锁定部分利润。
                      </Paragraph>
                    </div>

                    <div style={{
                      background: 'rgba(22,119,255,0.04)',
                      border: '1px solid rgba(22,119,255,0.12)',
                      padding: '20px 24px',
                      borderRadius: 12,
                    }}>
                      <Title level={5} style={{ color: '#fff', marginTop: 0 }}>🔍 换手率分析</Title>
                      <Paragraph style={{ color: 'rgba(255,255,255,0.6)', marginBottom: 0, lineHeight: 1.8 }}>
                        近 30 天换手率达到 120%，远高于基准水平。高频交易产生的佣金损耗已侵蚀掉约 2.4% 的潜在收益，建议拉长持股周期以降低摩擦成本。
                      </Paragraph>
                    </div>
                  </Card>
                </Space>
              </Col>
            </Row>
          )}
        </div>
      </div>
    </ConfigProvider>
  );
}