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
import { api } from '../types';

const { Title, Text, Paragraph } = Typography;

interface AnalysisReport {
  id: number;
  report_type: string;
  report_title: string;
  analysis_period_start: string;
  analysis_period_end: string;
  prediction_text: string;
  profit_rate: string;
  ai_model: string;
  created_at: string;
}

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const scenarios = [
  { label: '乐观情景 (P90)', value: '+34.2%', color: '#52c41a' },
  { label: '基准情景 (P50)', value: '+12.8%', color: '#1677ff' },
  { label: '悲观情景 (P10)', value: '-5.1%', color: '#ff4d4f' },
];

export default function PredictionPage() {
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

  const getOption = (): EChartsOption => {
    const historyData = [1.0, 1.02, 1.01, 1.05, 1.08, 1.07, 1.12];
    const predictMid = [1.12, 1.15, 1.18, 1.22];
    const predictUpper = [1.12, 1.18, 1.25, 1.35];
    const predictLower = [1.12, 1.11, 1.08, 1.05];

    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.96)',
        borderColor: '#d9e6ff',
        borderWidth: 1,
      },
      legend: {
        data: ['历史净值', '趋势路径', '乐观上限', '悲观下限'],
        bottom: 0,
        textStyle: { color: '#595959' },
      },
      grid: { top: '8%', left: '3%', right: '4%', bottom: '14%', containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '预测Q3', '预测Q4', '预测Y1', '预测Y2'],
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
        { name: '历史净值', type: 'line', data: historyData, smooth: true, lineStyle: { width: 3, color: '#1677ff' }, symbol: 'none' },
        { name: '趋势路径', type: 'line', data: [...Array(6).fill(null), ...predictMid], smooth: true, lineStyle: { type: 'dashed', width: 2, color: '#52c41a' }, symbol: 'circle', symbolSize: 6 },
        { name: '乐观上限', type: 'line', data: [...Array(6).fill(null), ...predictUpper], smooth: true, lineStyle: { width: 1, type: 'dotted', color: '#1677ff' }, symbol: 'none' },
        { name: '悲观下限', type: 'line', data: [...Array(6).fill(null), ...predictLower], smooth: true, lineStyle: { width: 1, type: 'dotted', color: '#ff4d4f' }, symbol: 'none' },
      ],
    };
  };

  const annualRate = useMemo(() => Number(report?.profit_rate ?? 12.8), [report]);

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
              当前后端暂无独立 `/analysis/prediction`，本页改为展示最近分析报告中的趋势结论。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<RiseOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              参考收益率 {annualRate}%
            </Tag>
            <Tag color="processing" icon={<StockOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
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
                  valueStyle={{ color: '#52c41a', fontSize: 34 }}
                  prefix={<StockOutlined />}
                />
                <Tag color="success" style={{ marginTop: 10, borderRadius: 20, padding: '2px 12px' }}>
                  最近分析报告推导
                </Tag>
              </Card>

              <Card bordered={false} style={cardStyle} title={<span><InfoCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />情景模拟区间</span>}>
                {scenarios.map((item, i) => (
                  <div key={item.label} style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '10px 0',
                    borderBottom: i < scenarios.length - 1 ? '1px solid #f0f0f0' : 'none',
                  }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                    <Tag color={item.color === '#52c41a' ? 'success' : item.color === '#1677ff' ? 'processing' : 'error'} style={{ borderRadius: 20, fontWeight: 700 }}>
                      {item.value}
                    </Tag>
                  </div>
                ))}
              </Card>

              <Alert
                message="风险提示"
                description="该页面当前展示的是基于最近分析报告的趋势结论，并非独立预测接口结果。"
                type="warning"
                showIcon
                icon={<InfoCircleOutlined />}
                style={{ borderRadius: 12 }}
              />
            </Space>
          </Col>

          <Col span={24} lg={18}>
            <Space direction="vertical" style={{ width: '100%' }} size={16}>
              <Card bordered={false} style={cardStyle} title={<span><LineChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />资产净值演变模拟（趋势展示）</span>} extra={<Text type="secondary" style={{ fontSize: 12 }}>报告时间: {report.created_at}</Text>}>
                <ReactECharts option={getOption()} style={{ height: '460px' }} />
              </Card>

              <Card bordered={false} style={cardStyle}>
                <Alert
                  type="info"
                  showIcon
                  icon={<InfoCircleOutlined />}
                  message={report.report_title || '最近一份分析报告'}
                  description="预测页已不再请求不存在的 /analysis/prediction，而是复用最近分析报告的趋势文本。"
                />
              </Card>
            </Space>
          </Col>
        </Row>
      )}
    </div>
  );
}
