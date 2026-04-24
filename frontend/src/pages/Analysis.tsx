import { useState, useEffect } from 'react';
import {
  Row, Col, Card, Statistic, Typography, Tag,
  Descriptions, Space, Skeleton, Progress, Alert, Button, Spin, Empty,
} from 'antd';
import {
  RadarChartOutlined, SafetyCertificateOutlined, BulbOutlined,
  InfoCircleOutlined, ArrowLeftOutlined, ThunderboltOutlined, ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi, transactionApi } from '../api/index';
import type { AnalysisReportResponse, TransactionResponse } from '../api/types';

const { Title, Paragraph, Text } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const riskScoreMap: Record<string, number> = {
  low: 90,
  medium: 74,
  high: 52,
  very_high: 35,
};

const styleMap: Record<string, string> = {
  aggressive: '激进型成长',
  balanced: '稳健均衡',
  conservative: '保守防御',
};

const marketStatusMap: Record<string, { color: string; text: string }> = {
  complete: { color: 'success', text: '市场数据完整' },
  fetched_live: { color: 'processing', text: '市场数据实时拉取' },
  partial: { color: 'warning', text: '市场数据部分缺失' },
  unavailable: { color: 'error', text: '市场数据不可用' },
};

function getMarketStatusMeta(status?: string) {
  return marketStatusMap[status ?? ''] ?? { color: 'default', text: status || '未知状态' };
}

function parseChartData(chartData?: string): { radar?: number[]; labels?: string[] } | null {
  if (!chartData) return null;
  try {
    return JSON.parse(chartData);
  } catch {
    return null;
  }
}

function formatValue(value?: string) {
  const text = value?.trim();
  return text ? text : '—';
}

function formatDateTime(value?: string) {
  if (!value?.trim()) return '—';
  return value.replace('T', ' ').replace('Z', '').slice(0, 19);
}

function hasStructuredAnalysis(report: AnalysisReportResponse | null) {
  return Boolean(
    report && (
      report.risk_analysis?.trim()
      || report.pattern_insights?.trim()
      || report.recommendations?.trim()
      || report.chart_data?.trim()
    ),
  );
}

function buildQuickMetrics(report: AnalysisReportResponse | null) {
  if (!report) return [];

  return [
    { label: '累计收益率', value: `${report.profit_rate}%`, color: '#1677ff', bg: '#e6f4ff' },
    { label: '风险等级', value: report.risk_level, color: '#ff4d4f', bg: '#fff1f0' },
    { label: '投资风格', value: styleMap[report.investment_style] ?? report.investment_style, color: '#722ed1', bg: '#f9f0ff' },
    { label: '数据状态', value: getMarketStatusMeta(report.market_data_status).text, color: '#13c2c2', bg: '#e6fffb' },
  ];
}

function normalizeDate(value?: string) {
  if (!value) return '';
  return value.slice(0, 10);
}

function compareDateAsc(a: string, b: string) {
  return a.localeCompare(b);
}

function getDateRangeFromTransactions(transactions: TransactionResponse[]) {
  const dates = transactions
    .map((item) => normalizeDate(item.transaction_date))
    .filter(Boolean)
    .sort(compareDateAsc);

  if (!dates.length) {
    return null;
  }

  return {
    startDate: dates[0],
    endDate: dates[dates.length - 1],
  };
}

export default function AnalysisPage() {
  const navigate = useNavigate();
  const [report, setReport] = useState<AnalysisReportResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    void fetchLatestReport();
  }, []);

  const fetchLatestReport = async () => {
    setLoading(true);
    setError('');
    try {
      const reports = await analysisApi.getReports();
      setReport(reports[0] ?? null);
    } catch (err: unknown) {
      const error = err as { message?: string; data?: { message?: string } };
      setReport(null);
      setError(error.message ?? error.data?.message ?? '分析报告加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleGenerate = async () => {
    setGenerating(true);
    setError('');
    try {
      const transactionRes = await transactionApi.getList({ page: 1, page_size: 1000 });
      const range = getDateRangeFromTransactions(transactionRes.transactions);

      if (!range) {
        setError('暂无可用于分析的交易记录');
        return;
      }

      const res = await analysisApi.generateSummary({
        start_date: range.startDate,
        end_date: range.endDate,
      });
      setReport(res);
    } catch (err: unknown) {
      const error = err as { message?: string; data?: { message?: string } };
      setError(error.message ?? error.data?.message ?? '分析生成失败');
    } finally {
      setGenerating(false);
    }
  };

  const structuredAnalysis = hasStructuredAnalysis(report);
  const score = report && structuredAnalysis ? (riskScoreMap[report.risk_level] ?? null) : null;
  const marketStatus = getMarketStatusMeta(report?.market_data_status);
  const chartData = parseChartData(report?.chart_data);
  const quickMetrics = structuredAnalysis ? buildQuickMetrics(report) : [];
  const profitRateText = report?.profit_rate?.trim() ? `${report.profit_rate}%` : '—';

  const getRadarOption = (): EChartsOption => ({
    radar: {
      indicator: (chartData?.labels ?? []).map((name: string) => ({ name, max: 100 })),
      shape: 'circle',
      splitNumber: 5,
      axisName: { color: '#8c8c8c' },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
      splitArea: { show: false },
    },
    series: [{
      type: 'radar',
      data: [{
        value: chartData?.radar ?? [],
        name: '特征评分',
        itemStyle: { color: '#1677ff' },
        lineStyle: { color: '#1677ff' },
        areaStyle: { color: 'rgba(22, 96, 255, 0.15)' },
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
          marginBottom: 24,
          borderRadius: 20,
          background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
          boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
        }}
        bodyStyle={{ padding: 28 }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">交易记录分析</Tag>
              <Tag color="blue">仅展示真实返回</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 投资分析</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              当前页面只展示后端真实返回的分析总结、风险字段和画像数据，不再使用预置结论或默认画像填充页面。
            </Paragraph>
          </div>
          <Space wrap>
            {score !== null && (
              <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                健康分 {score}
              </Tag>
            )}
            {report && (
              <Tag color={marketStatus.color} icon={<ThunderboltOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                {styleMap[report.investment_style] ?? formatValue(report.investment_style)}
              </Tag>
            )}
            {report?.market_data_status && (
              <Tag color={marketStatus.color} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                {marketStatus.text}
              </Tag>
            )}
            <Button ghost icon={<ReloadOutlined />} loading={generating} onClick={handleGenerate} style={{ borderRadius: 10 }}>
              重新生成
            </Button>
          </Space>
        </div>
      </Card>

      {loading ? (
        <Skeleton active paragraph={{ rows: 12 }} />
      ) : (
        <Spin spinning={generating} tip="AI 模型分析中...">
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            {error && <Alert type="error" showIcon message={error} />}

            {!report ? (
              <Card bordered={false} style={cardStyle}>
                <Empty description="暂无分析报告，请先导入交易记录后再生成分析。" />
              </Card>
            ) : (
              <Row gutter={[16, 16]}>
                <Col span={24} lg={8}>
                  <Space direction="vertical" style={{ width: '100%' }} size={16}>
                    <Card bordered={false} style={cardStyle}>
                      <Statistic
                        title="账户健康分"
                        value={score ?? '—'}
                        suffix={score !== null ? '/ 100' : undefined}
                        prefix={<SafetyCertificateOutlined />}
                        valueStyle={{ color: score !== null ? '#52c41a' : '#8c8c8c', fontSize: 34 }}
                      />
                      {score !== null ? (
                        <Progress
                          percent={score}
                          showInfo={false}
                          strokeColor={{ '0%': '#52c41a', '100%': '#95de64' }}
                          style={{ marginTop: 12 }}
                        />
                      ) : (
                        <Alert
                          type="info"
                          showIcon
                          message="当前报告仅返回文字总结，暂无可计算健康分的结构化字段。"
                          style={{ marginTop: 12 }}
                        />
                      )}
                      <Text type="secondary" style={{ fontSize: 12, marginTop: 6, display: 'block' }}>
                        盈利率 {profitRateText}
                      </Text>
                    </Card>

                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />投资风格画像</span>}
                    >
                      {chartData?.radar?.length && chartData?.labels?.length ? (
                        <>
                          <ReactECharts option={getRadarOption()} style={{ height: 260 }} />
                          <div style={{ textAlign: 'center', marginTop: 4 }}>
                            <Tag color="processing" style={{ borderRadius: 20, padding: '2px 14px' }}>
                              {styleMap[report.investment_style] ?? formatValue(report.investment_style)}
                            </Tag>
                          </div>
                        </>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无风格画像数据" />
                      )}
                    </Card>

                    <Card bordered={false} style={cardStyle}>
                      {quickMetrics.length ? (
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
                      ) : (
                        <Alert
                          type="info"
                          showIcon
                          message="当前报告暂无结构化指标卡片。"
                          description="当后端返回风险等级、风格画像和市场状态等完整字段后，这里才会展示对应指标。"
                        />
                      )}
                    </Card>
                  </Space>
                </Col>

                <Col span={24} lg={16}>
                  <Space direction="vertical" style={{ width: '100%' }} size={16}>
                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />AI 诊断结论</span>}
                    >
                      {!structuredAnalysis && (
                        <Alert
                          type="info"
                          showIcon
                          message="当前报告暂无结构化诊断字段。"
                          description="后端这次返回的是总结型报告，因此这里只显示真实返回的空值，不再补预置风险点或建议。"
                          style={{ marginBottom: 16 }}
                        />
                      )}
                      <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label="潜在风险点">
                          <Text type="danger">{formatValue(report.risk_analysis)}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label="行为特征">
                          {formatValue(report.pattern_insights)}
                        </Descriptions.Item>
                        <Descriptions.Item label="走势判断">
                          {formatValue(report.prediction_text)}
                        </Descriptions.Item>
                        <Descriptions.Item label="优化建议">
                          {formatValue(report.recommendations)}
                        </Descriptions.Item>
                      </Descriptions>
                    </Card>

                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />报告信息</span>}
                    >
                      <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label="报告标题">{formatValue(report.report_title)}</Descriptions.Item>
                        <Descriptions.Item label="报告类型">{formatValue(report.report_type)}</Descriptions.Item>
                        <Descriptions.Item label="分析周期">
                          {formatValue(report.analysis_period_start)} ~ {formatValue(report.analysis_period_end)}
                        </Descriptions.Item>
                        <Descriptions.Item label="累计投入">{formatValue(report.total_investment)}</Descriptions.Item>
                        <Descriptions.Item label="累计盈亏">{formatValue(report.total_profit)}</Descriptions.Item>
                        <Descriptions.Item label="生成时间">{formatDateTime(report.created_at)}</Descriptions.Item>
                      </Descriptions>
                    </Card>

                    <Card bordered={false} style={cardStyle}>
                      <Alert
                        type="info"
                        showIcon
                        icon={<BulbOutlined />}
                        message={`AI 总结：${formatValue(report.summary_text)}`}
                        description={
                          <Space direction="vertical" size={4}>
                            <Text type="secondary">分析周期：{formatValue(report.analysis_period_start)} ~ {formatValue(report.analysis_period_end)}</Text>
                            <Text type="secondary">模型版本：{formatValue(report.ai_model)}</Text>
                          </Space>
                        }
                      />
                    </Card>
                  </Space>
                </Col>
              </Row>
            )}
          </Space>
        </Spin>
      )}
    </div>
  );
}
