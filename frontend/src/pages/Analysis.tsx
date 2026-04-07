import { useState, useEffect } from 'react';
import {
  Row, Col, Card, Statistic, Typography, Divider, Tag,
  Descriptions, Space, Skeleton, Progress, Alert, Button, Spin
} from 'antd';
import {
  RadarChartOutlined, SafetyCertificateOutlined, BulbOutlined,
  InfoCircleOutlined, ArrowLeftOutlined, ThunderboltOutlined, ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi } from '../api/index';
import type { AnalysisReportResponse } from '../api/types';
import { mockAnalysisReport } from '../mockData';

const { Title, Paragraph, Text } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };
const subCardStyle = { background: '#f8fafc', borderRadius: 12 };

// 风险等级 → 健康分映射
const riskScoreMap: Record<string, number> = {
  low: 90, medium: 74, high: 52, very_high: 35,
};
const styleMap: Record<string, string> = {
  aggressive: '激进型成长', balanced: '稳健均衡', conservative: '保守防御',
};

export default function AnalysisPage() {
  const navigate = useNavigate();
  const [report, setReport]       = useState<AnalysisReportResponse | null>(null);
  const [loading, setLoading]     = useState(true);
  const [generating, setGenerating] = useState(false);

  useEffect(() => { fetchLatestReport(); }, []);

  const fetchLatestReport = async () => {
    setLoading(true);
    try {
      const reports = await analysisApi.getReports();
      setReport(reports[0] ?? null);
    } catch {
      setReport(mockAnalysisReport);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      // 计算默认日期范围（最近30天）
      const endDate = new Date().toISOString().split('T')[0];
      const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

      const res = await analysisApi.generateSummary({
        start_date: startDate,
        end_date: endDate
      });
      setReport(res);
    } catch {
      setReport(mockAnalysisReport);
    } finally {
      setGenerating(false);
    }
  };

  const score = report ? (riskScoreMap[report.risk_level] ?? 74) : 74;
  const chartData = (() => {
    try { return report ? JSON.parse(report.chart_data) : null; } catch { return null; }
  })();

  const getRadarOption = (): EChartsOption => ({
    radar: {
      indicator: (chartData?.labels ?? ['收益爆发力','回撤控制','资产分散度','交易纪律','风格稳定性'])
        .map((name: string) => ({ name, max: 100 })),
      shape: 'circle', splitNumber: 5,
      axisName: { color: '#8c8c8c' },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
      splitArea: { show: false },
    },
    series: [{
      type: 'radar',
      data: [{
        value: chartData?.radar ?? [82, 45, 30, 65, 90],
        name: '特征评分',
        itemStyle: { color: '#1677ff' },
        lineStyle: { color: '#1677ff' },
        areaStyle: { color: 'rgba(22, 96, 255, 0.15)' },
      }],
    }],
  });

  const quickMetrics = [
    { label: 'Beta 系数', value: '1.42', color: '#1677ff', bg: '#e6f4ff' },
    { label: '月换手率',  value: '120%', color: '#ff4d4f', bg: '#fff1f0' },
    { label: '夏普比率',  value: '0.87', color: '#1677ff', bg: '#e6f4ff' },
    { label: '最大回撤',  value: '4.9%', color: '#ff4d4f', bg: '#fff1f0' },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
        返回首页
      </Button>

      {/* Hero Banner */}
      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">AI 驱动</Tag>
              <Tag color="blue">多因子模型</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 深度风险诊断</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              基于多因子量化模型，深度穿透您的历史交易行为与仓位分布，识别潜在风险。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              健康分 {score}
            </Tag>
            <Tag color="processing" icon={<ThunderboltOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              {report ? styleMap[report.investment_style] ?? report.investment_style : '分析中'}
            </Tag>
            <Button ghost icon={<ReloadOutlined />} loading={generating} onClick={handleGenerate}
              style={{ borderRadius: 10 }}>
              重新生成
            </Button>
          </Space>
        </div>
      </Card>

      {loading ? <Skeleton active paragraph={{ rows: 12 }} /> : (
        <Spin spinning={generating} tip="AI 模型分析中...">
          <Row gutter={[16, 16]}>
            {/* 左栏 */}
            <Col span={24} lg={8}>
              <Space direction="vertical" style={{ width: '100%' }} size={16}>
                <Card bordered={false} style={cardStyle}>
                  <Statistic title="账户健康分" value={score} suffix="/ 100"
                    prefix={<SafetyCertificateOutlined />}
                    valueStyle={{ color: '#52c41a', fontSize: 34 }} />
                  <Progress percent={score} showInfo={false}
                    strokeColor={{ '0%': '#52c41a', '100%': '#95de64' }} style={{ marginTop: 12 }} />
                  <Text type="secondary" style={{ fontSize: 12, marginTop: 6, display: 'block' }}>
                    盈利率 {report?.profit_rate ?? '—'}%
                  </Text>
                </Card>

                <Card bordered={false} style={cardStyle}
                  title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />投资风格画像</span>}>
                  <ReactECharts option={getRadarOption()} style={{ height: 260 }} />
                  <div style={{ textAlign: 'center', marginTop: 4 }}>
                    <Tag color="processing" style={{ borderRadius: 20, padding: '2px 14px' }}>
                      {report ? (styleMap[report.investment_style] ?? report.investment_style) : '—'}
                    </Tag>
                  </div>
                </Card>

                <Card bordered={false} style={cardStyle}>
                  <Row gutter={[12, 12]}>
                    {quickMetrics.map(item => (
                      <Col span={12} key={item.label}>
                        <div style={{ background: item.bg, borderRadius: 12, padding: '14px 16px' }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                          <div style={{ color: item.color, fontSize: 22, fontWeight: 700, marginTop: 4 }}>{item.value}</div>
                        </div>
                      </Col>
                    ))}
                  </Row>
                </Card>
              </Space>
            </Col>

            {/* 右栏 */}
            <Col span={24} lg={16}>
              <Space direction="vertical" style={{ width: '100%' }} size={16}>
                <Card bordered={false} style={cardStyle}
                  title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />AI 诊断结论</span>}>
                  <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label="潜在风险点">
                      <Text type="danger">{report?.risk_analysis ?? '—'}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="行为特征">
                      {report?.pattern_insights ?? '—'}
                    </Descriptions.Item>
                    <Descriptions.Item label="优化建议">
                      {report?.recommendations ?? '—'}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>

                <Card bordered={false} style={cardStyle}
                  title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />行为特征报告</span>}>
                  <div style={{ ...subCardStyle, padding: '18px 20px', marginBottom: 12 }}>
                    <Title level={5} style={{ marginTop: 0 }}>⚡ 市场敏感度扫描</Title>
                    <Paragraph type="secondary" style={{ marginBottom: 0, lineHeight: 1.8 }}>
                      您的组合 Beta 系数为 1.42，意味着市场每波动 1%，您的账户预期波动 1.42%。这表明您处于杠杆化配置状态，在牛市表现优异，但在宽幅震荡期可能面临较大压力。
                    </Paragraph>
                  </div>
                  <div style={{ ...subCardStyle, padding: '18px 20px' }}>
                    <Title level={5} style={{ marginTop: 0 }}>🔍 换手率分析</Title>
                    <Paragraph type="secondary" style={{ marginBottom: 0, lineHeight: 1.8 }}>
                      近 30 天换手率达到 120%，高频交易产生的佣金损耗已侵蚀掉约 2.4% 的潜在收益，建议拉长持股周期以降低摩擦成本。
                    </Paragraph>
                  </div>
                </Card>

                <Card bordered={false} style={cardStyle}>
                  <Alert type="info" showIcon icon={<BulbOutlined />}
                    message={`AI 一句话结论：${report?.summary_text ?? '分析中...'}`}
                    description={
                      <Space direction="vertical" size={4}>
                        <Text type="secondary">分析周期：{report?.analysis_period_start} ~ {report?.analysis_period_end}</Text>
                        <Text type="secondary">模型版本：{report?.ai_model ?? '—'}</Text>
                      </Space>
                    }
                  />
                </Card>
              </Space>
            </Col>
          </Row>
        </Spin>
      )}
    </div>
  );
}