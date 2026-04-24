import { useEffect, useMemo, useState } from 'react';
import {
  Card, Row, Col, Typography, Statistic, Space,
  Tag, Alert, Button, Spin, Empty
} from 'antd';
import {
  LineChartOutlined, ThunderboltOutlined, InfoCircleOutlined,
  StockOutlined, ArrowLeftOutlined, RiseOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { analysisApi } from '../api';
import type { AnalysisReportResponse } from '../api/types';
import { mockAnalysisReport } from '../mockData';

const { Title, Text, Paragraph } = Typography;

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

interface ProfitChartPoint {
  symbol: string;
  value: string;
}

const marketStatusMap: Record<string, { color: string; text: string }> = {
  complete: { color: 'success', text: '市场数据完整' },
  fetched_live: { color: 'processing', text: '市场数据实时拉取' },
  partial: { color: 'warning', text: '市场数据部分缺失' },
  unavailable: { color: 'error', text: '市场数据不可用' },
};

function getMarketStatusMeta(status?: string) {
  return marketStatusMap[status ?? ''] ?? { color: 'default', text: status || '未知状态' };
}

function parseProfitChartData(chartData?: string): ProfitChartPoint[] {
  if (!chartData) {
    return [];
  }

  try {
    const parsed = JSON.parse(chartData) as ProfitChartPoint[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function toNumber(value?: string): number {
  if (!value) return 0;
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

export default function PredictionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [report, setReport] = useState<AnalysisReportResponse | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const reports = await analysisApi.getReports({ report_type: 'summary', limit: 1 });
        setReport(reports[0] ?? null);
      } catch {
        setReport(mockAnalysisReport);
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, []);

  const annualRate = useMemo(() => toNumber(report?.profit_rate), [report]);
  const marketStatus = useMemo(() => getMarketStatusMeta(report?.market_data_status), [report]);
  const profitChartData = useMemo(() => parseProfitChartData(report?.chart_data), [report]);
  const topProfitPoint = useMemo(() => {
    if (!profitChartData.length) {
      return null;
    }

    return [...profitChartData].sort((a, b) => toNumber(b.value) - toNumber(a.value))[0];
  }, [profitChartData]);

  const getOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => {
        const list = params as Array<{ axisValueLabel?: string; value: number }>;
        const data = list[0];
        return `<div style="padding: 4px 6px;">
                  <div style="color: #888; margin-bottom: 4px;">${data.axisValueLabel ?? ''}</div>
                  <div style="font-weight: bold; color: #1677ff; font-size: 16px;">${data.value.toFixed(2)}</div>
                </div>`;
      },
    },
    grid: { top: '8%', left: '3%', right: '4%', bottom: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: profitChartData.map(item => item.symbol),
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
        data: profitChartData.map(item => toNumber(item.value)),
        itemStyle: {
          color: (params: { value?: unknown }) => (Number(params.value ?? 0) >= 0 ? '#52c41a' : '#ff4d4f'),
          borderRadius: [8, 8, 0, 0],
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
          marginBottom: 24, borderRadius: 20,
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
              当前后端暂无独立预测接口，本页展示最近一份 summary 报告中的趋势结论与收益分布。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              参考收益率 {annualRate}%
            </Tag>
            <Tag color={marketStatus.color} icon={<StockOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              {marketStatus.text}
            </Tag>
            <Tag color="processing" icon={<ThunderboltOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              模型 {report?.ai_model || '—'}
            </Tag>
          </Space>
        </div>
      </Card>

      {loading ? (
        <Spin spinning />
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
                  {report.prediction_text || '暂无趋势预测文本，当前以后端分析报告中的趋势字段为准。'}
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
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>最高收益标的</Text>
                  <Text strong>{topProfitPoint ? `${topProfitPoint.symbol} (${topProfitPoint.value})` : '暂无数据'}</Text>
                </div>
              </Card>

              <Alert
                message="说明"
                description="该页面当前展示的是最近分析报告中的预测文本与收益分布，不再伪造独立趋势路径。"
                type="info"
                showIcon
                icon={<InfoCircleOutlined />}
                style={{ borderRadius: 12 }}
              />
            </Space>
          </Col>

          <Col span={24} lg={18}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle} title={<span><LineChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />个股累计收益分布</span>} extra={<Text type="secondary" style={{ fontSize: 12 }}>报告时间: {report.created_at}</Text>}>
                {profitChartData.length ? (
                  <ReactECharts option={getOption()} style={{ height: '460px' }} />
                ) : (
                  <Empty description="当前报告没有可用的图表数据" />
                )}
              </Card>

              <Card bordered={false} style={cardStyle}>
                <Alert
                  type="info"
                  showIcon
                  icon={<InfoCircleOutlined />}
                  message={report.report_title || '最近一份分析报告'}
                  description={report.summary_text || '当前预测页复用最近分析报告的数据。'}
                />
              </Card>
            </Space>
          </Col>
        </Row>
      )}
    </div>
  );
}
