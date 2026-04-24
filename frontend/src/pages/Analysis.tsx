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
import type {
  AnalysisReportDetailResponse,
  AnalysisTaskDetailResponse,
} from '../api/types';

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

const taskStageMap: Record<string, string> = {
  pending: '任务已创建',
  collecting_transactions: '正在收集交易记录',
  preparing_metrics: '正在准备指标数据',
  generating_stock_reports: '正在生成个股分析',
  generating_summary: '正在生成总结报告',
  persisting_report: '正在保存分析结果',
  completed: '分析已完成',
};

function getMarketStatusMeta(status?: string) {
  return marketStatusMap[status ?? ''] ?? { color: 'default', text: status || '未知状态' };
}

type ParsedChartData =
  | { type: 'radar'; labels: string[]; values: number[] }
  | { type: 'bar'; labels: string[]; values: number[] }
  | null;

function parseChartData(chartData?: string): ParsedChartData {
  if (!chartData) return null;

  try {
    const parsed = JSON.parse(chartData) as unknown;

    if (Array.isArray(parsed)) {
      const points = parsed
        .map((item) => {
          if (typeof item !== 'object' || item === null) return null;
          const point = item as { symbol?: unknown; value?: unknown };
          if (typeof point.symbol !== 'string') return null;
          const value = Number(point.value);
          if (Number.isNaN(value)) return null;
          return { label: point.symbol, value };
        })
        .filter((item): item is { label: string; value: number } => Boolean(item));

      if (points.length) {
        return {
          type: 'bar',
          labels: points.map((item) => item.label),
          values: points.map((item) => item.value),
        };
      }
    }

    if (typeof parsed === 'object' && parsed !== null) {
      const data = parsed as { labels?: unknown; radar?: unknown };
      if (Array.isArray(data.labels) && Array.isArray(data.radar)) {
        const labels = data.labels.filter((item): item is string => typeof item === 'string');
        const values = data.radar
          .map((item) => Number(item))
          .filter((item) => !Number.isNaN(item));

        if (labels.length && labels.length === values.length) {
          return { type: 'radar', labels, values };
        }
      }
    }
  } catch {
    return null;
  }

  return null;
}

function formatValue(value?: string) {
  const text = value?.trim();
  return text ? text : '—';
}

function formatDateTime(value?: string) {
  if (!value?.trim()) return '—';
  return value.replace('T', ' ').replace('Z', '').slice(0, 19);
}

function buildQuickMetrics(report: AnalysisReportDetailResponse | null) {
  if (!report) return [];

  return [
    { label: '累计收益率', value: `${report.profit_rate}%`, color: '#1677ff', bg: '#e6f4ff' },
    { label: '风险等级', value: report.risk_level, color: '#ff4d4f', bg: '#fff1f0' },
    { label: '投资风格', value: styleMap[report.investment_style] ?? report.investment_style, color: '#722ed1', bg: '#f9f0ff' },
    { label: '覆盖标的', value: `${report.symbols_count}`, color: '#13c2c2', bg: '#e6fffb' },
  ];
}

function normalizeDate(value?: string) {
  if (!value) return '';
  return value.slice(0, 10);
}

function delay(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function getTaskStageText(stage?: string) {
  return taskStageMap[stage ?? ''] ?? '分析任务处理中';
}

function renderRecommendations(recommendations: string[]) {
  if (!recommendations.length) {
    return '—';
  }

  return (
    <Space direction="vertical" size={4}>
      {recommendations.map((item, index) => (
        <Text key={`${index}-${item}`}>{index + 1}. {item}</Text>
      ))}
    </Space>
  );
}

export default function AnalysisPage() {
  const navigate = useNavigate();
  const [report, setReport] = useState<AnalysisReportDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');
  const [taskStage, setTaskStage] = useState('');

  useEffect(() => {
    void fetchLatestReport();
  }, []);

  const fetchLatestReport = async () => {
    setLoading(true);
    setError('');
    try {
      const tasks = await analysisApi.getTasks({ status: 'success', page: 1, page_size: 1 });
      const latestTask = tasks.items[0];

      if (!latestTask?.result_report_id) {
        setReport(null);
        return;
      }

      const detail = await analysisApi.getReportDetail(latestTask.result_report_id);
      setReport(detail);
    } catch (err: unknown) {
      const error = err as { message?: string; data?: { message?: string } };
      setReport(null);
      setError(error.message ?? error.data?.message ?? '分析报告加载失败');
    } finally {
      setLoading(false);
    }
  };

  const getTransactionDateRange = async () => {
    const latestRes = await transactionApi.getList({ page: 1, page_size: 1 });
    const latest = latestRes.transactions[0];

    if (!latestRes.total || !latest) {
      return null;
    }

    if (latestRes.total === 1) {
      const date = normalizeDate(latest.transaction_date);
      return { startDate: date, endDate: date };
    }

    const earliestRes = await transactionApi.getList({ page: latestRes.total, page_size: 1 });
    const earliest = earliestRes.transactions[0];

    if (!earliest) {
      return null;
    }

    return {
      startDate: normalizeDate(earliest.transaction_date),
      endDate: normalizeDate(latest.transaction_date),
    };
  };

  const waitForTask = async (taskId: number) => {
    for (let attempt = 0; attempt < 90; attempt += 1) {
      const task = await analysisApi.getTask(taskId);
      setTaskStage(getTaskStageText(task.progress_stage));

      if (task.status === 'success') {
        return task;
      }

      if (task.status === 'failed') {
        throw new Error(task.error_message || '分析任务失败');
      }

      await delay(2000);
    }

    throw new Error('分析任务仍在处理中，请稍后刷新页面查看结果');
  };

  const handleGenerate = async () => {
    setGenerating(true);
    setError('');
    setTaskStage('正在创建分析任务');

    try {
      const range = await getTransactionDateRange();

      if (!range?.startDate || !range?.endDate) {
        setError('暂无可用于分析的交易记录');
        return;
      }

      const task = await analysisApi.createTask({
        start_date: range.startDate,
        end_date: range.endDate,
      });

      const completedTask: AnalysisTaskDetailResponse = await waitForTask(task.id);

      if (!completedTask.result_report_id) {
        throw new Error('分析任务已完成，但未返回报告 ID');
      }

      const detail = await analysisApi.getReportDetail(completedTask.result_report_id);
      setReport(detail);
      setTaskStage('分析已完成');
    } catch (err: unknown) {
      const error = err as { message?: string; data?: { message?: string } };
      setError(error.message ?? error.data?.message ?? '分析生成失败');
      setTaskStage('');
    } finally {
      setGenerating(false);
    }
  };

  const score = report ? (riskScoreMap[report.risk_level] ?? null) : null;
  const marketStatus = getMarketStatusMeta(report?.market_data_status);
  const chartData = parseChartData(report?.chart_data);
  const quickMetrics = buildQuickMetrics(report);
  const profitRateText = report?.profit_rate?.trim() ? `${report.profit_rate}%` : '—';

  const getChartOption = (): EChartsOption => {
    if (chartData?.type === 'radar') {
      return {
        radar: {
          indicator: chartData.labels.map((name) => ({ name, max: 100 })),
          shape: 'circle',
          splitNumber: 5,
          axisName: { color: '#8c8c8c' },
          splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
          splitArea: { show: false },
        },
        series: [{
          type: 'radar',
          data: [{
            value: chartData.values,
            name: '特征评分',
            itemStyle: { color: '#1677ff' },
            lineStyle: { color: '#1677ff' },
            areaStyle: { color: 'rgba(22, 96, 255, 0.15)' },
          }],
        }],
      };
    }

    return {
      tooltip: { trigger: 'axis' },
      grid: { left: 48, right: 24, top: 24, bottom: 48 },
      xAxis: {
        type: 'category',
        data: chartData?.labels ?? [],
        axisLabel: { color: '#8c8c8c', interval: 0, rotate: 20 },
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#8c8c8c' },
        splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
      },
      series: [{
        type: 'bar',
        data: chartData?.values ?? [],
        itemStyle: {
          color: '#1677ff',
          borderRadius: [6, 6, 0, 0],
        },
      }],
    };
  };

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
              <Tag color="processing">结构化分析</Tag>
              <Tag color="blue">异步任务生成</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 投资分析</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              当前页面通过后端分析任务生成结构化报告，并展示真实返回的风险等级、投资风格、图表数据和 AI 结论。
            </Paragraph>
          </div>
          <Space wrap>
            {score !== null && (
              <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                健康分 {score}
              </Tag>
            )}
            {report?.investment_style && (
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
        <Spin spinning={generating} tip={taskStage || 'AI 结构化分析中...'}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            {taskStage && generating && <Alert type="info" showIcon message={taskStage} />}
            {error && <Alert type="error" showIcon message={error} />}

            {!report ? (
              <Card bordered={false} style={cardStyle}>
                <Empty description="暂无结构化分析报告，请先导入交易记录后再生成分析。" />
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
                      {score !== null && (
                        <Progress
                          percent={score}
                          showInfo={false}
                          strokeColor={{ '0%': '#52c41a', '100%': '#95de64' }}
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
                      title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />分析图表</span>}
                    >
                      {chartData ? (
                        <>
                          <ReactECharts option={getChartOption()} style={{ height: 260 }} />
                          <div style={{ textAlign: 'center', marginTop: 4 }}>
                            <Tag color="processing" style={{ borderRadius: 20, padding: '2px 14px' }}>
                              {styleMap[report.investment_style] ?? formatValue(report.investment_style)}
                            </Tag>
                          </div>
                        </>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无图表数据" />
                      )}
                    </Card>

                    <Card bordered={false} style={cardStyle}>
                      <Row gutter={[12, 12]}>
                        {quickMetrics.map((item) => (
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

                <Col span={24} lg={16}>
                  <Space direction="vertical" style={{ width: '100%' }} size={16}>
                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />AI 诊断结论</span>}
                    >
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
                          {renderRecommendations(report.recommendations)}
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
                        <Descriptions.Item label="覆盖标的数">{report.symbols_count}</Descriptions.Item>
                        <Descriptions.Item label="盈利标的数">{report.winning_trades}</Descriptions.Item>
                        <Descriptions.Item label="亏损标的数">{report.losing_trades}</Descriptions.Item>
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
