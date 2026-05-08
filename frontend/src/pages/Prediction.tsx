import { useEffect, useMemo, useState } from 'react';
import {
  Card, Row, Col, Typography, Statistic, Space,
  Tag, Alert, Button, Spin, Empty,
} from 'antd';
import {
  LineChartOutlined, ThunderboltOutlined, InfoCircleOutlined,
  StockOutlined, ArrowLeftOutlined, RiseOutlined, ReloadOutlined, FundOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi } from '../api';
import type { AnalysisReportDetailResponse, AnalysisReportResponse } from '../api/types';
import {
  buildChangePercentRankingData,
  buildProfitBySymbolViewData,
  buildProfitCompositionData,
  formatProfitValue,
  summarizeProfitBySymbolData,
  toNumericProfitValue,
} from '../utils/analysisChart';
import { getMarketStatusMeta } from '../utils/analysisMeta';

const { Title, Text, Paragraph } = Typography;

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

function formatValue(value?: string) {
  const text = value?.trim();
  return text ? text : '—';
}

function isUsableSummary(report: AnalysisReportResponse) {
  return Boolean(report.prediction_text?.trim()) || Boolean(report.summary_text?.trim()) || Boolean(report.chart_data?.trim());
}

function pickPredictionReport(reports: AnalysisReportResponse[]) {
  return reports.find(isUsableSummary) ?? reports[0] ?? null;
}

export default function PredictionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [report, setReport] = useState<AnalysisReportDetailResponse | null>(null);
  const [error, setError] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const reports = await analysisApi.getReports({ report_type: 'summary', limit: 5 });
      const summaryReport = pickPredictionReport(reports);
      if (!summaryReport) {
        setReport(null);
        return;
      }

      const detail = await analysisApi.getReportDetail(summaryReport.id);
      setReport(detail);
    } catch (err: unknown) {
      const apiError = err as { message?: string; data?: { message?: string } };
      setReport(null);
      setError(apiError.message ?? apiError.data?.message ?? '趋势结论加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const annualRate = useMemo(() => toNumericProfitValue(report?.profit_rate), [report?.profit_rate]);
  const marketStatus = useMemo(() => getMarketStatusMeta(report?.market_data_status), [report?.market_data_status]);
  const profitChartData = useMemo(() => buildProfitBySymbolViewData(report?.chart_data, report?.items), [report?.chart_data, report?.items]);
  const chartSummary = useMemo(() => summarizeProfitBySymbolData(profitChartData), [profitChartData]);
  const momentumData = useMemo(() => buildChangePercentRankingData(report?.items), [report?.items]);
  const compositionSummary = useMemo(() => buildProfitCompositionData(report?.items), [report?.items]);

  const getProfitOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => {
        const list = params as Array<{ dataIndex?: number }>;
        const point = profitChartData[list[0]?.dataIndex ?? -1];
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
    grid: { top: '8%', left: '3%', right: '4%', bottom: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: profitChartData.map((item) => item.symbol),
      axisLine: { lineStyle: { color: '#d9d9d9' } },
      axisLabel: { color: '#8c8c8c' },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } },
      axisLabel: { color: '#8c8c8c' },
    },
    series: [
      {
        name: '累计收益',
        type: 'bar',
        data: profitChartData.map((item) => ({ value: item.numericValue, itemStyle: { color: item.color } })),
        itemStyle: { borderRadius: [8, 8, 0, 0] },
        markLine: {
          silent: true,
          symbol: 'none',
          lineStyle: { color: '#bfbfbf', type: 'dashed' },
          data: [{ yAxis: 0 }],
        },
      },
    ],
  });

  const getMomentumOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const list = params as Array<{ dataIndex?: number }>;
        const point = momentumData[list[0]?.dataIndex ?? -1];
        if (!point) {
          return '';
        }
        return `<div style="padding: 4px 6px;">
                  <div style="color: #888; margin-bottom: 4px;">${point.symbol}</div>
                  <div style="font-weight: bold; color: ${point.color}; font-size: 16px;">${formatProfitValue(point.numericValue)}%</div>
                  <div style="margin-top: 4px; color: #666;">${point.semanticLabel}</div>
                </div>`;
      },
    },
    grid: { left: 56, right: 20, top: 24, bottom: 24, containLabel: true },
    xAxis: {
      type: 'value',
      axisLabel: { color: '#8c8c8c' },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } },
    },
    yAxis: {
      type: 'category',
      data: momentumData.map((item) => item.symbol),
      axisLabel: { color: '#8c8c8c' },
      axisLine: { show: false },
    },
    series: [
      {
        name: '近 7 日变化率',
        type: 'bar',
        data: momentumData.map((item) => ({ value: item.numericValue, itemStyle: { color: item.color } })),
        itemStyle: { borderRadius: [0, 8, 8, 0] },
        markLine: {
          silent: true,
          symbol: 'none',
          lineStyle: { color: '#bfbfbf', type: 'dashed' },
          data: [{ xAxis: 0 }],
        },
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
              <Tag color="processing">趋势结论</Tag>
              <Tag color="blue">复用分析报告</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>AI 收益趋势预测</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              当前后端暂无独立预测接口，本页展示最近一份 summary 报告中的趋势结论、收益分布与结构动量视角。
            </Paragraph>
          </div>
          {report && (
            <Space wrap>
              <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                参考收益率 {annualRate}%
              </Tag>
              <Tag color={marketStatus.color} icon={<StockOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                {marketStatus.text}
              </Tag>
              <Tag color="processing" icon={<ThunderboltOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
                模型 {formatValue(report.ai_model)}
              </Tag>
            </Space>
          )}
          <Button ghost icon={<ReloadOutlined />} onClick={() => void load()} loading={loading} style={{ borderRadius: 10 }}>
            重新加载
          </Button>
        </div>
      </Card>

      {loading ? (
        <Spin spinning />
      ) : error ? (
        <Card bordered={false} style={cardStyle}>
          <Alert
            type="error"
            showIcon
            message={error}
            action={<Button size="small" onClick={() => void load()}>重试</Button>}
          />
        </Card>
      ) : !report ? (
        <Card bordered={false} style={cardStyle}>
          <Empty description="暂无可用趋势结论，请先生成分析报告" />
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          <Col span={24} lg={6}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle} title={<span><ThunderboltOutlined style={{ color: '#1677ff', marginRight: 8 }} />趋势结论</span>}>
                <Paragraph type="secondary" style={{ marginBottom: 0, lineHeight: 1.8 }}>
                  {formatValue(report.prediction_text)}
                </Paragraph>
              </Card>

              <Card bordered={false} style={cardStyle}>
                <Statistic
                  title="参考收益率"
                  value={annualRate}
                  suffix="%"
                  precision={2}
                  valueStyle={{ color: annualRate >= 0 ? '#52c41a' : '#ff4d4f', fontSize: 34 }}
                  prefix={<StockOutlined />}
                />
                <Tag color="success" style={{ marginTop: 10, borderRadius: 20, padding: '2px 12px' }}>
                  最近分析报告推导
                </Tag>
              </Card>

              <Card bordered={false} style={cardStyle} title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />关键信息</span>}>
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>报告类型</Text>
                  <Tag color="processing" style={{ borderRadius: 20, fontWeight: 700 }}>{report.report_type}</Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>分析周期</Text>
                  <Text strong>{report.analysis_period_start} ~ {report.analysis_period_end}</Text>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid #f0f0f0' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>最高收益标的</Text>
                  <Text strong>{chartSummary.topPoint ? `${chartSummary.topPoint.symbol} (${formatProfitValue(chartSummary.topPoint.numericValue)})` : '—'}</Text>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>最低收益标的</Text>
                  <Text strong>{chartSummary.bottomPoint ? `${chartSummary.bottomPoint.symbol} (${formatProfitValue(chartSummary.bottomPoint.numericValue)})` : '—'}</Text>
                </div>
              </Card>

              <Alert
                message="说明"
                description="该页面当前展示的是最近分析报告中的预测文本、收益分布和结构摘要，不再伪造独立趋势路径。"
                type="info"
                showIcon
                icon={<InfoCircleOutlined />}
                style={{ borderRadius: 12 }}
              />
            </Space>
          </Col>

          <Col span={24} lg={18}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle} title={<span><LineChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />个股累计收益分布</span>} extra={<Text type="secondary" style={{ fontSize: 12 }}>报告时间: {formatValue(report.created_at)}</Text>}>
                {profitChartData.length ? (
                  <>
                    <ReactECharts option={getProfitOption()} style={{ height: '360px' }} />
                    <Row gutter={[8, 8]} style={{ marginTop: 12 }}>
                      <Col span={8}>
                        <Tag color="success" style={{ width: '100%', textAlign: 'center', padding: '6px 0', borderRadius: 20 }}>盈利 {chartSummary.positiveCount}</Tag>
                      </Col>
                      <Col span={8}>
                        <Tag color="error" style={{ width: '100%', textAlign: 'center', padding: '6px 0', borderRadius: 20 }}>亏损 {chartSummary.negativeCount}</Tag>
                      </Col>
                      <Col span={8}>
                        <Tag color="default" style={{ width: '100%', textAlign: 'center', padding: '6px 0', borderRadius: 20 }}>持平 {chartSummary.zeroCount}</Tag>
                      </Col>
                    </Row>
                  </>
                ) : (
                  <Empty description="当前报告没有可用的收益分布数据" />
                )}
              </Card>

              <Row gutter={[16, 16]}>
                <Col span={24} xl={14}>
                  <Card bordered={false} style={cardStyle} title={<span><FundOutlined style={{ color: '#1677ff', marginRight: 8 }} />近 7 日变化率排序</span>}>
                    {momentumData.length ? (
                      <ReactECharts option={getMomentumOption()} style={{ height: '320px' }} />
                    ) : (
                      <Empty description="当前报告暂无可用的 7 日变化率数据" />
                    )}
                  </Card>
                </Col>
                <Col span={24} xl={10}>
                  <Card bordered={false} style={cardStyle} title={<span><StockOutlined style={{ color: '#1677ff', marginRight: 8 }} />结构摘要</span>}>
                    <Space direction="vertical" size={14} style={{ width: '100%' }}>
                      <div style={{ background: '#e6f4ff', borderRadius: 12, padding: '14px 16px' }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>收益分布结论</Text>
                        <div style={{ color: '#1677ff', fontSize: 18, fontWeight: 700, marginTop: 4 }}>
                          {chartSummary.topPoint && chartSummary.bottomPoint ? `${chartSummary.topPoint.symbol} 领跑，${chartSummary.bottomPoint.symbol} 承压` : '暂无足够数据'}
                        </div>
                      </div>
                      <div style={{ background: '#f6ffed', borderRadius: 12, padding: '14px 16px' }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>已实现 / 浮动收益</Text>
                        <div style={{ color: '#52c41a', fontSize: 16, fontWeight: 700, marginTop: 4 }}>
                          {formatProfitValue(compositionSummary.realizedTotal)} / {formatProfitValue(compositionSummary.unrealizedTotal)}
                        </div>
                      </div>
                      <div style={{ background: '#fff7e6', borderRadius: 12, padding: '14px 16px' }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>结构数据状态</Text>
                        <div style={{ color: '#d48806', fontSize: 16, fontWeight: 700, marginTop: 4 }}>
                          {profitChartData.length ? '收益分布可用' : '收益分布不足'} · {momentumData.length ? '动量数据可用' : '动量数据不足'}
                        </div>
                      </div>
                    </Space>
                  </Card>
                </Col>
              </Row>

              <Card bordered={false} style={cardStyle}>
                <Alert
                  type="info"
                  showIcon
                  icon={<InfoCircleOutlined />}
                  message={formatValue(report.report_title)}
                  description={formatValue(report.summary_text)}
                />
              </Card>
            </Space>
          </Col>
        </Row>
      )}
    </div>
  );
}