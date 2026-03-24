import { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Typography, Divider, Tag, Descriptions, Space, Skeleton, Progress } from 'antd';
import { 
  RadarChartOutlined, 
  SafetyCertificateOutlined, 
  ThunderboltOutlined, 
  RiseOutlined,
  BulbOutlined,
  InfoCircleOutlined
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

const { Title, Paragraph, Text } = Typography;

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
      axisName: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.05)' } },
      splitArea: { show: false }
    },
    series: [{
      type: 'radar',
      data: [{
        value: [82, 45, 30, 65, 90],
        name: '特征评分',
        itemStyle: { color: '#1677ff' },
        areaStyle: { color: 'rgba(22, 119, 255, 0.2)' }
      }]
    }]
  });

  if (loading) return <div style={{ padding: 24 }}><Skeleton active paragraph={{ rows: 12 }} /></div>;

  return (
    <div style={{ padding: '8px' }}>
      <Title level={3}>AI 深度风险诊断</Title>
      <Paragraph type="secondary">基于多因子量化模型，深度穿透您的历史交易行为与仓位分布。</Paragraph>

      <Row gutter={[16, 16]}>
        <Col span={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false}>
              <Statistic 
                title="账户健康分" 
                value={74.2} 
                suffix="/ 100"
                prefix={<SafetyCertificateOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
              <Progress percent={74.2} showInfo={false} strokeColor="#52c41a" style={{ marginTop: 8 }} />
            </Card>
            <Card title={<span><RadarChartOutlined /> 投资风格画像</span>} bordered={false}>
              <ReactECharts option={getRadarOption()} style={{ height: 300 }} />
              <div style={{ textAlign: 'center' }}><Tag color="blue">激进型成长风格</Tag></div>
            </Card>
          </Space>
        </Col>

        <Col span={24} lg={16}>
          <Card 
            bordered={false} 
            title={<span><BulbOutlined style={{ color: '#1677ff' }} /> AI 诊断结论</span>}
          >
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="潜在风险点">
                <Text type="danger">持仓集中度过高。</Text> 您的前两大持仓占总资产 65%，极易受单一行业波动影响。
              </Descriptions.Item>
              <Descriptions.Item label="交易倾向">
                检测到轻微的 <Text strong>“处置效应”</Text>（倾向于过早卖出盈利股，而长期持有亏损股）。
              </Descriptions.Item>
              <Descriptions.Item label="优化建议">
                建议将科技板块仓位下调 15%，增配防御性资产如红利低波 ETF。
              </Descriptions.Item>
            </Descriptions>

            <Divider orientation="left"><InfoCircleOutlined /> 行为特征报告</Divider>
            <div style={{ background: '#f8fafc', padding: 20, borderRadius: 8 }}>
              <Title level={5}>⚡ 市场敏感度扫描</Title>
              <Paragraph>
                您的组合 Beta 系数为 1.42，意味着市场每波动 1%，您的账户预期波动 1.42%。这表明您处于杠杆化配置状态，在牛市表现优异，但在宽幅震荡期可能面临较大压力。建议通过对冲工具锁定部分利润。
              </Paragraph>
              <Title level={5}>🔍 换手率分析</Title>
              <Paragraph>
                近 30 天换手率达到 120%，远高于基准水平。高频交易产生的佣金损耗已侵蚀掉约 2.4% 的潜在收益，建议拉长持股周期以降低摩擦成本。
              </Paragraph>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}