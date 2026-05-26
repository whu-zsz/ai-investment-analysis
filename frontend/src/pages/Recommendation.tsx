import { useEffect, useState } from 'react';
import { Alert, Button, Card, Col, Empty, List, Row, Space, Spin, Statistic, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, BulbOutlined, MessageOutlined, ReloadOutlined, SafetyCertificateOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { analysisApi } from '../api';
import type { AnalysisCandidatesResponse, AnalysisReportDetailResponse, AnalysisReportResponse } from '../api/types';

const { Title, Text, Paragraph } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

function actionMeta(action: string) {
  switch (action) {
    case 'buy':
      return { color: 'success', text: '优先关注/可加仓' };
    case 'hold':
      return { color: 'processing', text: '继续持有' };
    case 'reduce':
      return { color: 'warning', text: '适度减仓' };
    case 'sell':
      return { color: 'error', text: '考虑卖出' };
    default:
      return { color: 'default', text: '继续观察' };
  }
}

const preferenceMap: Record<string, string> = {
  conservative: '保守型',
  balanced: '稳健型',
  aggressive: '激进型',
};

export default function RecommendationPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [candidateInfo, setCandidateInfo] = useState<AnalysisCandidatesResponse | null>(null);
  const [latestReport, setLatestReport] = useState<AnalysisReportDetailResponse | null>(null);
  const [historyReports, setHistoryReports] = useState<AnalysisReportResponse[]>([]);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [candidateResult, historyResult] = await Promise.allSettled([
        analysisApi.getCandidates(),
        analysisApi.getReports({ report_type: 'recommendation', limit: 8 }),
      ]);

      const nextCandidateInfo = candidateResult.status === 'fulfilled' ? candidateResult.value : null;
      const nextHistory = historyResult.status === 'fulfilled' ? historyResult.value : [];

      setCandidateInfo(nextCandidateInfo);
      setHistoryReports(nextHistory);

      const reportId = Number(searchParams.get('reportId') || nextHistory[0]?.id || 0);
      if (reportId > 0) {
        try {
          const detail = await analysisApi.getReportDetail(reportId);
          setLatestReport(detail);
        } catch {
          setLatestReport(null);
        }
      } else {
        setLatestReport(null);
      }

      if (!nextCandidateInfo && nextHistory.length === 0) {
        const candidateError = candidateResult.status === 'rejected' ? candidateResult.reason : null;
        setError(
          candidateError?.message
            ?? candidateError?.data?.message
            ?? 'AI 推荐加载失败'
        );
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [searchParams]);

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')} style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
        返回首页
      </Button>

      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">AI 推荐</Tag>
              <Tag color="blue">基于用户相关股票池实时生成</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>适合你的股票推荐</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 620 }}>
              通过几轮推荐对话，结合你的偏好、已有持仓、历史关注和近期新闻热点，生成可追溯的推荐报告。
            </Paragraph>
          </div>
          <Space>
            <Button ghost icon={<MessageOutlined />} onClick={() => navigate('/app/chat?kind=recommendation')} style={{ borderRadius: 10 }}>
              发起推荐对话
            </Button>
            <Button ghost icon={<ReloadOutlined />} onClick={() => void load()} loading={loading} style={{ borderRadius: 10 }}>
              刷新报告
            </Button>
          </Space>
        </div>
      </Card>

      <Spin spinning={loading}>
        {error && latestReport ? (
          <Card bordered={false} style={{ ...cardStyle, marginBottom: 16 }}>
            <Alert type="warning" showIcon message={error} />
          </Card>
        ) : null}

        {error && !latestReport && !candidateInfo ? (
          <Card bordered={false} style={cardStyle}><Alert type="error" showIcon message={error} /></Card>
        ) : !latestReport && !candidateInfo ? (
          <Card bordered={false} style={cardStyle}><Empty description="暂无推荐结果" /></Card>
        ) : latestReport ? (
          <>
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="候选池数量" value={candidateInfo?.candidates.length ?? latestReport.items.length} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="当前持仓数" value={candidateInfo?.candidates.filter((item) => item.is_held).length ?? 0} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="推荐标的数" value={latestReport.items.length} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="报告时间" value={latestReport.created_at.slice(11, 16)} /></Card></Col>
            </Row>

            <Card bordered={false} style={cardStyle} title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />最近推荐报告</span>} extra={<Tag color="processing">{latestReport.report_title}</Tag>}>
              <Paragraph style={{ marginBottom: 12 }}>{latestReport.summary_text}</Paragraph>
              <Space wrap>
                <Tag color="processing" icon={<SafetyCertificateOutlined />}>风险等级：{latestReport.risk_level || '—'}</Tag>
                <Tag color="blue" icon={<ThunderboltOutlined />}>投资偏好：{preferenceMap[latestReport.investment_style || ''] ?? latestReport.investment_style ?? '—'}</Tag>
                <Tag>生成时间：{latestReport.created_at}</Tag>
              </Space>
            </Card>

            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
              {latestReport.items.map((item) => {
                const meta = actionMeta(item.recommendation);
                return (
                  <Col span={24} lg={12} key={`${latestReport.id}-${item.symbol}`}>
                    <Card bordered={false} style={cardStyle}>
                      <Space direction="vertical" size={10} style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'flex-start' }}>
                          <div>
                            <Title level={4} style={{ margin: 0 }}>{item.asset_name || item.symbol}</Title>
                            <Text type="secondary">{item.symbol}</Text>
                          </div>
                          <Tag color={meta.color}>{meta.text}</Tag>
                        </div>
                        <Row gutter={[12, 12]}>
                          <Col span={8}><Statistic title="最新价" value={toNumber(item.latest_price)} precision={2} prefix="¥" valueStyle={{ fontSize: 22 }} /></Col>
                          <Col span={8}><Statistic title="阶段变化" value={toNumber(item.period_price_change_pct)} precision={2} suffix="%" valueStyle={{ color: toNumber(item.period_price_change_pct) >= 0 ? '#52c41a' : '#ff4d4f', fontSize: 22 }} /></Col>
                          <Col span={8}><Statistic title="历史关注" value={item.trade_count} valueStyle={{ fontSize: 22 }} /></Col>
                        </Row>
                        <div>
                          <Text strong>推荐理由</Text>
                          <Paragraph style={{ margin: '6px 0 0' }}>{item.analysis_text}</Paragraph>
                        </div>
                        <div>
                          <Text strong>重点</Text>
                          <Paragraph style={{ margin: '6px 0 0' }}>{item.key_points?.join('；') || '—'}</Paragraph>
                        </div>
                        <Space wrap>
                          <Tag>数据状态：{item.market_data_status}</Tag>
                          <Button type="link" onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(item.symbol)}`)} style={{ paddingInline: 0 }}>
                            查看市场趋势
                          </Button>
                          <Button type="link" onClick={() => navigate(`/app/chat?kind=recommendation&reportId=${latestReport.id}`)} style={{ paddingInline: 0 }}>
                            继续追问本报告
                          </Button>
                        </Space>
                      </Space>
                    </Card>
                  </Col>
                );
              })}
            </Row>

            <Card bordered={false} style={{ ...cardStyle, marginTop: 16 }} title="历史推荐报告">
              {historyReports.length ? (
                <List
                  dataSource={historyReports}
                  renderItem={(item) => (
                    <List.Item
                      actions={[
                        <Button key="open" type="link" onClick={() => navigate(`/app/recommendation?reportId=${item.id}`)}>查看</Button>,
                        <Button key="chat" type="link" onClick={() => navigate(`/app/chat?kind=recommendation&reportId=${item.id}`)}>继续对话</Button>,
                      ]}
                    >
                      <List.Item.Meta
                        title={<Space wrap><span>{item.report_title}</span>{item.id === latestReport.id ? <Tag color="success">当前</Tag> : null}</Space>}
                        description={`${item.created_at} · ${item.risk_level || '—'} 风险 · ${item.id === latestReport.id ? latestReport.items.length : '查看详情'} `}
                      />
                    </List.Item>
                  )}
                />
              ) : (
                <Empty description="暂无历史推荐报告" />
              )}
            </Card>
          </>
        ) : (
          <Card bordered={false} style={cardStyle}>
            <Empty description="暂无推荐报告，先发起一轮推荐对话。" />
            <div style={{ marginTop: 16 }}>
              <Space wrap>
                <Tag color="processing">候选池数量 {candidateInfo?.candidates.length ?? 0}</Tag>
                {candidateInfo?.default_symbol ? <Tag color="blue">默认关注 {candidateInfo.default_symbol}</Tag> : null}
                <Button type="primary" icon={<MessageOutlined />} onClick={() => navigate('/app/chat?kind=recommendation')}>
                  发起推荐对话
                </Button>
              </Space>
            </div>
          </Card>
        )}
      </Spin>
    </div>
  );
}
