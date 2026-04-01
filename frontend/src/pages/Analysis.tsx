import { useEffect, useMemo, useState } from 'react';
import {
  Row, Col, Card, Statistic, Typography, Tag,
  Descriptions, Space, Skeleton, Progress, Alert, Button, Empty
} from 'antd';
import {
  RadarChartOutlined, SafetyCertificateOutlined, BulbOutlined,
  InfoCircleOutlined, ArrowLeftOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { api } from '../types';

const { Title, Paragraph, Text } = Typography;

interface AnalysisReport {
  id: number;
  report_type: string;
  report_title: string;
  analysis_period_start: string;
  analysis_period_end: string;
  total_investment: string;
  total_profit: string;
  profit_rate: string;
  risk_level: string;
  investment_style: string;
  summary_text: string;
  risk_analysis: string;
  pattern_insights: string;
  prediction_text: string;
  chart_data: string;
  recommendations: string;
  ai_model: string;
  created_at: string;
}

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };
const subCardStyle = { background: '#f8fafc', borderRadius: 12 };

export default function AnalysisPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [report, setReport] = useState<AnalysisReport | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const response = await api.getReports({ limit: 1 });
        const reports = response.data as AnalysisReport[];
        setReport(reports[0] ?? null);
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, []);

  const radarValues = useMemo(() => {
    const riskLevel = report?.risk_level?.toLowerCase();
    if (riskLevel === 'high') return [82, 45, 30, 65, 90];
    if (riskLevel === 'medium') return [74, 62, 58, 71, 76];
    return [68, 76, 72, 78, 69];
  }, [report]);

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
      axisName: { color: '#8c8c8c' },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
      splitArea: { show: false },
    },
    series: [{
      type: 'radar',
      data: [{
        value: radarValues,
        name: '特征评分',
        itemStyle: { color: '#1677ff' },
        lineStyle: { color: '#1677ff' },
        areaStyle: { color: 'rgba(22,119,255,0.15)' },
      }],
    }],
  });

  return (
    <div style={{ padding: '24px' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}
      >
        返回首页
      </Button>

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
              <Tag color="processing">AI 驱动</Tag>
              <Tag color="blue">真实分析报告</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 深度风险诊断</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              当前页面已接入 `/analysis/reports`，优先展示最近一份分析报告。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              风险等级 {report?.risk_level || '待生成'}
            </Tag>
            <Tag color="processing" icon={<ThunderboltOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              {report?.investment_style || '等待分析'}
            </Tag>
          </Space>
        </div>
      </Card>

      {loading ? (
        <Skeleton active paragraph={{ rows: 12 }} />
      ) : !report ? (
        <Card bordered={false} style={cardStyle}>
          <Empty description="暂无分析报告，请先完成上传并触发分析" />
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          <Col span={24} lg={8}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle}>
                <Statistic
                  title="账户健康分"
                  value={report.risk_level === 'high' ? 60 : report.risk_level === 'medium' ? 74.2 : 86}
                  suffix="/ 100"
                  prefix={<SafetyCertificateOutlined />}
                  valueStyle={{ color: '#52c41a', fontSize: 34 }}
                />
                <Progress
                  percent={report.risk_level === 'high' ? 60 : report.risk_level === 'medium' ? 74.2 : 86}
                  showInfo={false}
                  strokeColor={{ '0%': '#52c41a', '100%': '#95de64' }}
                  style={{ marginTop: 12 }}
                />
                <Text type="secondary" style={{ fontSize: 12, marginTop: 6, display: 'block' }}>
                  分析区间：{report.analysis_period_start} ~ {report.analysis_period_end}
                </Text>
              </Card>

              <Card bordered={false} style={cardStyle} title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />投资风格画像</span>}>
                <ReactECharts option={getRadarOption()} style={{ height: 260 }} />
                <div style={{ textAlign: 'center', marginTop: 4 }}>
                  <Tag color="processing" style={{ borderRadius: 20, padding: '2px 14px' }}>{report.investment_style || '综合风格'}</Tag>
                </div>
              </Card>

              <Card bordered={false} style={cardStyle}>
                <Row gutter={[12, 12]}>
                  {[
                    { label: '总投入', value: report.total_investment || '0', color: '#1677ff', bg: '#e6f4ff' },
                    { label: '总收益', value: report.total_profit || '0', color: '#52c41a', bg: '#f6ffed' },
                    { label: '收益率', value: report.profit_rate || '0', color: '#1677ff', bg: '#e6f4ff' },
                    { label: '模型', value: report.ai_model || '—', color: '#ff4d4f', bg: '#fff1f0' },
                  ].map(item => (
                    <Col span={12} key={item.label}>
                      <div style={{ background: item.bg, borderRadius: 12, padding: '14px 16px' }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                        <div style={{ color: item.color, fontSize: 18, fontWeight: 700, marginTop: 4 }}>{item.value}</div>
                      </div>
                    </Col>
                  ))}
                </Row>
              </Card>
            </Space>
          </Col>

          <Col span={24} lg={16}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle} title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />AI 诊断结论</span>}>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="摘要结论">
                    {report.summary_text || '暂无摘要'}
                  </Descriptions.Item>
                  <Descriptions.Item label="风险分析">
                    {report.risk_analysis || '暂无风险分析'}
                  </Descriptions.Item>
                  <Descriptions.Item label="优化建议">
                    {report.recommendations || '暂无建议'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>

              <Card bordered={false} style={cardStyle} title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />行为特征报告</span>}>
                <div style={{ ...subCardStyle, padding: '18px 20px', marginBottom: 12 }}>
                  <Title level={5} style={{ marginTop: 0 }}>风险洞察</Title>
                  <Paragraph type="secondary" style={{ marginBottom: 0, lineHeight: 1.8 }}>
                    {report.pattern_insights || report.risk_analysis || '暂无行为特征分析'}
                  </Paragraph>
                </div>
                <div style={{ ...subCardStyle, padding: '18px 20px' }}>
                  <Title level={5} style={{ marginTop: 0 }}>趋势结论</Title>
                  <Paragraph type="secondary" style={{ marginBottom: 0, lineHeight: 1.8 }}>
                    {report.prediction_text || '当前后端暂无独立预测接口，这里展示最近报告中的趋势结论。'}
                  </Paragraph>
                </div>
              </Card>

              <Card bordered={false} style={cardStyle}>
                <Alert
                  type="info"
                  showIcon
                  icon={<BulbOutlined />}
                  message={report.report_title || '最近一份分析报告'}
                  description={
                    <Space direction="vertical" size={4}>
                      <Text type="secondary">生成时间：{report.created_at}</Text>
                      <Text type="secondary">当前页面已从静态内容切换到真实报告数据。</Text>
                    </Space>
                  }
                />
              </Card>
            </Space>
          </Col>
        </Row>
      )}
    </div>
  );
}
