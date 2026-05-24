import { useMemo, useState, useEffect } from 'react';
import {
  Row, Col, Card, Statistic, Typography, Tag,
  Descriptions, Space, Skeleton, Alert, Button, Spin, Empty, List,
} from 'antd';
import {
  BarChartOutlined, SafetyCertificateOutlined, BulbOutlined,
  InfoCircleOutlined, ArrowLeftOutlined, ThunderboltOutlined, ReloadOutlined,
  DownloadOutlined, PieChartOutlined, FundOutlined, AlertOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi, transactionApi } from '../api/index';
import type {
  AnalysisReportDetailResponse,
  AnalysisTaskDetailResponse,
  AnalysisReportItemResponse,
  RiskAlertItemResponse,
  RiskSymbolResponse,
} from '../api/types';
import {
  buildOutcomeDistributionData,
  buildProfitBySymbolViewData,
  buildProfitCompositionData,
  formatProfitValue,
  summarizeProfitBySymbolData,
} from '../utils/analysisChart';
import { getMarketStatusMeta } from '../utils/analysisMeta';

const { Title, Paragraph, Text } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const styleMap: Record<string, string> = {
  aggressive: '激进型成长',
  balanced: '稳健均衡',
  conservative: '保守防御',
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

function formatValue(value?: string) {
  const text = value?.trim();
  return text ? text : '—';
}

function formatDateTime(value?: string) {
  if (!value?.trim()) return '—';
  return value.replace('T', ' ').replace('Z', '').slice(0, 19);
}

function getRiskTagColor(level?: string) {
  if (level === 'low') return 'success';
  if (level === 'medium') return 'warning';
  return 'error';
}

function getAlertTypeColor(level?: string) {
  if (level === 'low') return 'processing';
  if (level === 'medium') return 'warning';
  return 'error';
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

function renderKeyPoints(keyPoints: string[]) {
  if (!keyPoints.length) {
    return '—';
  }

  return (
    <Space direction="vertical" size={4}>
      {keyPoints.map((item, index) => (
        <Text key={`${index}-${item}`}>{index + 1}. {item}</Text>
      ))}
    </Space>
  );
}

function renderStockCards(items: AnalysisReportItemResponse[], topRiskSymbols: RiskSymbolResponse[]) {
  if (!items.length) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无个股分析数据" />;
  }

  const highRiskMap = new Map(topRiskSymbols.map((item) => [item.symbol, item]));

  return (
    <Row gutter={[12, 12]}>
      {items.slice(0, 5).map((item) => {
        const riskInfo = highRiskMap.get(item.symbol);
        return (
          <Col span={24} key={item.id}>
            <Card size="small" bordered style={{ borderRadius: 12, borderColor: riskInfo ? '#ffd8bf' : undefined }}>
              <Space direction="vertical" style={{ width: '100%' }} size={8}>
                <Space wrap>
                  <Text strong>{item.symbol}</Text>
                  <Tag color="blue">{item.asset_name}</Tag>
                  <Tag color={item.recommendation === 'buy' ? 'success' : item.recommendation === 'sell' ? 'red' : 'processing'}>
                    {item.recommendation}
                  </Tag>
                  {riskInfo && <Tag color={getRiskTagColor(riskInfo.risk_level)}>风险分 {riskInfo.risk_score}</Tag>}
                </Space>
                <Space wrap>
                  <Text type="secondary">总盈亏 {formatValue(item.total_profit)}</Text>
                  <Text type="secondary">风险 {formatValue(item.risk_level)}</Text>
                  <Text type="secondary">风格 {styleMap[item.investment_style] ?? formatValue(item.investment_style)}</Text>
                </Space>
                {riskInfo?.trigger_reasons?.length ? (
                  <Alert
                    type="warning"
                    showIcon
                    message="触发预警"
                    description={riskInfo.trigger_reasons.join('；')}
                    style={{ borderRadius: 12 }}
                  />
                ) : null}
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {formatValue(item.analysis_text)}
                </Paragraph>
                <Text type="secondary" style={{ fontSize: 12 }}>要点</Text>
                {renderKeyPoints(item.key_points)}
              </Space>
            </Card>
          </Col>
        );
      })}
    </Row>
  );
}

function renderRiskAlerts(alerts: RiskAlertItemResponse[]) {
  if (!alerts.length) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告未触发结构化风险预警" />;
  }

  return (
    <List
      dataSource={alerts}
      renderItem={(item) => (
        <List.Item style={{ paddingInline: 0 }}>
          <Card size="small" bordered style={{ width: '100%', borderRadius: 12 }}>
            <Space direction="vertical" style={{ width: '100%' }} size={8}>
              <Space wrap>
                <Tag color={getAlertTypeColor(item.level)}>{item.level.toUpperCase()}</Tag>
                <Text strong>{item.title}</Text>
              </Space>
              <Text type="secondary">{item.description}</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                涉及标的：{item.symbols.length ? item.symbols.join('、') : '—'}
              </Text>
            </Space>
          </Card>
        </List.Item>
      )}
    />
  );
}

export default function AnalysisPage() {
  const navigate = useNavigate();
  const [report, setReport] = useState<AnalysisReportDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [exporting, setExporting] = useState(false);
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
      const errorObj = err as { message?: string; data?: { message?: string } };
      setReport(null);
      setError(errorObj.message ?? errorObj.data?.message ?? '分析报告加载失败');
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
      const errorObj = err as { message?: string; data?: { message?: string } };
      setError(errorObj.message ?? errorObj.data?.message ?? '分析生成失败');
      setTaskStage('');
    } finally {
      setGenerating(false);
    }
  };

  const handleExportPDF = async () => {
    if (!report?.id) return;

    setExporting(true);
    setError('');
    try {
      const blob = await analysisApi.downloadReportPDF(report.id);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `analysis-report-${report.id}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err: unknown) {
      const errorObj = err as { message?: string; data?: { message?: string } };
      setError(errorObj.message ?? errorObj.data?.message ?? 'PDF 导出失败');
    } finally {
      setExporting(false);
    }
  };

  const marketStatus = getMarketStatusMeta(report?.market_data_status);
  const chartData = useMemo(() => buildProfitBySymbolViewData(report?.chart_data, report?.items), [report?.chart_data, report?.items]);
  const chartSummary = useMemo(() => summarizeProfitBySymbolData(chartData), [chartData]);
  const outcomeDistribution = useMemo(() => buildOutcomeDistributionData(report), [report]);
  const profitComposition = useMemo(() => buildProfitCompositionData(report?.items), [report?.items]);
  const quickMetrics = buildQuickMetrics(report);

  const getProfitChartOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => {
        const list = params as Array<{ dataIndex?: number }>;
        const point = chartData[list[0]?.dataIndex ?? -1];
        if (!point) {
          return '';
        }
        return `<div style="padding: 4px 6px;">
                  <div style="color: #888; margin-bottom: 4px;">${point.symbol}</div>
                  <div style="font-weight: bold; color: ${point.color}; font-size: 16px;">${formatProfitValue(point.numericValue)}</div>
                  <div style="margin-top: 4px; color: #666;">${point.semanticLabel}</div>
                </div>`;
      },
    },
    grid: { left: 48, right: 24, top: 24, bottom: 48 },
    xAxis: {
      type: 'category',
      data: chartData.map((item) => item.symbol),
      axisLabel: { color: '#8c8c8c', interval: 0, rotate: 20 },
      axisLine: { lineStyle: { color: '#d9d9d9' } },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#8c8c8c' },
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
    },
    series: [{
      type: 'bar',
      name: '累计收益',
      data: chartData.map((item) => ({ value: item.numericValue, itemStyle: { color: item.color } })),
      itemStyle: { borderRadius: [6, 6, 0, 0] },
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { color: '#bfbfbf', type: 'dashed' },
        data: [{ yAxis: 0 }],
      },
    }],
  });

  const getOutcomeDistributionOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      bottom: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: '#8c8c8c' },
    },
    series: [
      {
        type: 'pie',
        radius: ['52%', '74%'],
        center: ['50%', '42%'],
        avoidLabelOverlap: false,
        label: {
          show: true,
          formatter: '{b}\n{d}%',
          color: '#595959',
          fontSize: 12,
        },
        labelLine: { length: 12, length2: 10 },
        data: outcomeDistribution.points.map((item) => ({
          name: item.label,
          value: item.value,
          itemStyle: { color: item.color },
        })),
      },
    ],
  });

  const getProfitCompositionOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const list = params as Array<{ seriesName?: string; value?: number; color?: string }>;
        const rowIndex = (params as Array<{ dataIndex?: number }>)[0]?.dataIndex ?? -1;
        const point = profitComposition.points[rowIndex];
        if (!point) {
          return '';
        }
        const rows = list.map((item) => `<div style="color:${item.color}">${item.seriesName}：${formatProfitValue(item.value ?? 0)}</div>`).join('');
        return `<div style="padding: 4px 6px;"><div style="font-weight:700;margin-bottom:4px;">${point.symbol}</div>${rows}</div>`;
      },
    },
    grid: { left: 48, right: 20, top: 24, bottom: 24, containLabel: true },
    xAxis: {
      type: 'value',
      axisLabel: { color: '#8c8c8c' },
      splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
    },
    yAxis: {
      type: 'category',
      data: profitComposition.points.map((item) => item.symbol),
      axisLabel: { color: '#8c8c8c' },
      axisLine: { show: false },
    },
    legend: {
      top: 0,
      textStyle: { color: '#8c8c8c' },
    },
    series: [
      {
        name: '已实现收益',
        type: 'bar',
        stack: 'profit',
        itemStyle: { color: '#1677ff', borderRadius: [0, 8, 8, 0] },
        data: profitComposition.points.map((item) => item.realizedProfit),
      },
      {
        name: '浮动收益',
        type: 'bar',
        stack: 'profit',
        itemStyle: { color: '#73d13d', borderRadius: [0, 8, 8, 0] },
        data: profitComposition.points.map((item) => item.unrealizedProfit),
      },
    ],
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
              <Tag color="processing">结构化分析</Tag>
              <Tag color="blue">智能报告生成</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 投资分析</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              生成投资分析报告，集中查看风险等级、预警信息、图表洞察和 AI 结论。
            </Paragraph>
          </div>
          <Space wrap>
            {report?.risk_level && (
              <Tag color={getRiskTagColor(report.risk_level)} icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                风险等级 {report.risk_level}
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
            <Button
              ghost
              icon={<DownloadOutlined />}
              loading={exporting}
              onClick={handleExportPDF}
              disabled={!report}
              style={{ borderRadius: 10 }}
            >
              导出 PDF
            </Button>
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
              <>
                <Row gutter={[16, 16]}>
                  <Col span={24} lg={8}>
                    <Card bordered={false} style={cardStyle}>
                      <Statistic
                        title="累计收益率"
                        value={report.profit_rate || '—'}
                        suffix={report.profit_rate ? '%' : undefined}
                        prefix={<SafetyCertificateOutlined />}
                        valueStyle={{ color: '#52c41a', fontSize: 34 }}
                      />
                      <Text type="secondary" style={{ fontSize: 12, marginTop: 6, display: 'block' }}>
                        风险等级 {report.risk_level} · 数据状态 {marketStatus.text}
                      </Text>
                    </Card>
                  </Col>
                  <Col span={24} lg={16}>
                    <Card bordered={false} style={cardStyle}>
                      <Row gutter={[12, 12]}>
                        {quickMetrics.map((item) => (
                          <Col span={12} md={6} key={item.label}>
                            <div style={{ background: item.bg, borderRadius: 12, padding: '14px 16px', height: '100%' }}>
                              <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                              <div style={{ color: item.color, fontSize: 22, fontWeight: 700, marginTop: 4 }}>{item.value}</div>
                            </div>
                          </Col>
                        ))}
                      </Row>
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={8}>
                    <Card bordered={false} style={cardStyle} title={<span><AlertOutlined style={{ color: '#1677ff', marginRight: 8 }} />风险预警总览</span>}>
                      <Row gutter={[12, 12]}>
                        <Col span={12}>
                          <div style={{ background: '#fff7e6', borderRadius: 12, padding: '12px 14px' }}>
                            <Text type="secondary" style={{ fontSize: 12 }}>风险等级</Text>
                            <div style={{ marginTop: 6 }}><Tag color={getRiskTagColor(report.risk_overview?.risk_level)}>{formatValue(report.risk_overview?.risk_level)}</Tag></div>
                          </div>
                        </Col>
                        <Col span={12}>
                          <div style={{ background: '#fff1f0', borderRadius: 12, padding: '12px 14px' }}>
                            <Text type="secondary" style={{ fontSize: 12 }}>风险分数</Text>
                            <div style={{ color: '#ff4d4f', fontSize: 24, fontWeight: 700, marginTop: 4 }}>{report.risk_overview?.risk_score ?? 0}</div>
                          </div>
                        </Col>
                        <Col span={24}>
                          <div style={{ background: '#f6ffed', borderRadius: 12, padding: '12px 14px' }}>
                            <Text type="secondary" style={{ fontSize: 12 }}>主要风险因子</Text>
                            <div style={{ marginTop: 8 }}>
                              {report.risk_overview?.risk_factors?.length ? report.risk_overview.risk_factors.map((item) => (
                                <Tag key={item} color="warning" style={{ marginBottom: 8 }}>{item}</Tag>
                              )) : <Text type="secondary">当前未识别到结构化风险因子</Text>}
                            </div>
                          </div>
                        </Col>
                      </Row>
                    </Card>
                  </Col>
                  <Col span={24} xl={9}>
                    <Card bordered={false} style={cardStyle} title={<span><SafetyCertificateOutlined style={{ color: '#1677ff', marginRight: 8 }} />预警列表</span>}>
                      {renderRiskAlerts(report.risk_alerts ?? [])}
                    </Card>
                  </Col>
                  <Col span={24} xl={7}>
                    <Card bordered={false} style={cardStyle} title={<span><FundOutlined style={{ color: '#1677ff', marginRight: 8 }} />高风险标的排行</span>}>
                      {report.top_risk_symbols?.length ? (
                        <List
                          dataSource={report.top_risk_symbols}
                          renderItem={(item) => (
                            <List.Item style={{ paddingInline: 0 }}>
                              <Card size="small" bordered style={{ width: '100%', borderRadius: 12 }}>
                                <Space direction="vertical" style={{ width: '100%' }} size={6}>
                                  <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                                    <Space wrap>
                                      <Text strong>{item.symbol}</Text>
                                      <Tag color="blue">{item.asset_name}</Tag>
                                    </Space>
                                    <Tag color={getRiskTagColor(item.risk_level)}>风险分 {item.risk_score}</Tag>
                                  </Space>
                                  <Space wrap>
                                    {item.trigger_reasons.map((reason) => (
                                      <Tag key={`${item.symbol}-${reason}`} color="warning">{reason}</Tag>
                                    ))}
                                  </Space>
                                </Space>
                              </Card>
                            </List.Item>
                          )}
                        />
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前暂无高风险标的" />
                      )}
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={14}>
                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><BarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />个股累计收益分布</span>}
                    >
                      {chartData.length ? (
                        <>
                          <ReactECharts option={getProfitChartOption()} style={{ height: 320 }} />
                          <Row gutter={[8, 8]} style={{ marginTop: 12 }}>
                            <Col span={12}>
                              <div style={{ background: '#f6ffed', borderRadius: 12, padding: '10px 12px' }}>
                                <Text type="secondary" style={{ fontSize: 12 }}>最高收益</Text>
                                <div style={{ color: '#52c41a', fontSize: 18, fontWeight: 700, marginTop: 4 }}>
                                  {chartSummary.topPoint ? `${chartSummary.topPoint.symbol} ${formatProfitValue(chartSummary.topPoint.numericValue)}` : '—'}
                                </div>
                              </div>
                            </Col>
                            <Col span={12}>
                              <div style={{ background: '#fff1f0', borderRadius: 12, padding: '10px 12px' }}>
                                <Text type="secondary" style={{ fontSize: 12 }}>最低收益</Text>
                                <div style={{ color: '#ff4d4f', fontSize: 18, fontWeight: 700, marginTop: 4 }}>
                                  {chartSummary.bottomPoint ? `${chartSummary.bottomPoint.symbol} ${formatProfitValue(chartSummary.bottomPoint.numericValue)}` : '—'}
                                </div>
                              </div>
                            </Col>
                          </Row>
                        </>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无可用收益分布数据" />
                      )}
                    </Card>
                  </Col>

                  <Col span={24} xl={10}>
                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><PieChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />胜负占比</span>}
                    >
                      {outcomeDistribution.total ? (
                        <>
                          <ReactECharts option={getOutcomeDistributionOption()} style={{ height: 320 }} />
                          <Row gutter={[8, 8]} style={{ marginTop: 12 }}>
                            {outcomeDistribution.points.map((item) => (
                              <Col span={8} key={item.key}>
                                <div style={{ borderRadius: 12, padding: '10px 12px', background: `${item.color}12` }}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                                  <div style={{ color: item.color, fontSize: 18, fontWeight: 700, marginTop: 4 }}>{item.value}</div>
                                  <div style={{ color: '#8c8c8c', fontSize: 12 }}>{item.percent}%</div>
                                </div>
                              </Col>
                            ))}
                          </Row>
                        </>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无可用占比数据" />
                      )}
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]}>
                  <Col span={24} xl={12}>
                    <Card
                      bordered={false}
                      style={cardStyle}
                      title={<span><FundOutlined style={{ color: '#1677ff', marginRight: 8 }} />已实现与浮动收益构成</span>}
                    >
                      {profitComposition.points.length ? (
                        <>
                          <ReactECharts option={getProfitCompositionOption()} style={{ height: 300 }} />
                          <Row gutter={[8, 8]} style={{ marginTop: 12 }}>
                            <Col span={12}>
                              <div style={{ background: '#e6f4ff', borderRadius: 12, padding: '10px 12px' }}>
                                <Text type="secondary" style={{ fontSize: 12 }}>已实现收益汇总</Text>
                                <div style={{ color: '#1677ff', fontSize: 18, fontWeight: 700, marginTop: 4 }}>
                                  {formatProfitValue(profitComposition.realizedTotal)}
                                </div>
                              </div>
                            </Col>
                            <Col span={12}>
                              <div style={{ background: '#f6ffed', borderRadius: 12, padding: '10px 12px' }}>
                                <Text type="secondary" style={{ fontSize: 12 }}>浮动收益汇总</Text>
                                <div style={{ color: '#52c41a', fontSize: 18, fontWeight: 700, marginTop: 4 }}>
                                  {formatProfitValue(profitComposition.unrealizedTotal)}
                                </div>
                              </div>
                            </Col>
                          </Row>
                        </>
                      ) : (
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前报告暂无可用收益构成数据" />
                      )}
                    </Card>
                  </Col>

                  <Col span={24} xl={12}>
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
                  </Col>
                </Row>

                <Row gutter={[16, 16]}>
                  <Col span={24} lg={10}>
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
                  </Col>

                  <Col span={24} lg={14}>
                    <Card bordered={false} style={cardStyle} title={<span><BarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />个股分析</span>}>
                      {renderStockCards(report.items, report.top_risk_symbols ?? [])}
                    </Card>
                  </Col>
                </Row>

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
              </>
            )}
          </Space>
        </Spin>
      )}
    </div>
  );
}
