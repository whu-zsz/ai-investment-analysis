import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Empty, Row, Space, Spin, Statistic, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, BulbOutlined, ReloadOutlined, SafetyCertificateOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { analysisApi } from '../api';
import type { AnalysisCandidatesResponse, AnalysisRecommendationsResponse, RecommendationItemResponse } from '../api/types';

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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [candidateInfo, setCandidateInfo] = useState<AnalysisCandidatesResponse | null>(null);
  const [recommendation, setRecommendation] = useState<AnalysisRecommendationsResponse | null>(null);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [candidateResult, recommendationResult] = await Promise.allSettled([
        analysisApi.getCandidates(),
        analysisApi.getRecommendations(),
      ]);

      const nextCandidateInfo = candidateResult.status === 'fulfilled' ? candidateResult.value : null;
      const nextRecommendation = recommendationResult.status === 'fulfilled' ? recommendationResult.value : null;

      setCandidateInfo(nextCandidateInfo);
      setRecommendation(nextRecommendation);

      if (!nextCandidateInfo && !nextRecommendation) {
        const candidateError = candidateResult.status === 'rejected' ? candidateResult.reason : null;
        const recommendationError = recommendationResult.status === 'rejected' ? recommendationResult.reason : null;
        setError(
          recommendationError?.message
            ?? recommendationError?.data?.message
            ?? candidateError?.message
            ?? candidateError?.data?.message
            ?? 'AI 推荐加载失败'
        );
      } else if (!nextRecommendation && nextCandidateInfo) {
        setError('AI 解释生成较慢，当前先展示候选股票池。');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const topCandidate = useMemo<RecommendationItemResponse | null>(() => recommendation?.candidates?.[0] ?? null, [recommendation]);
  const profileSummary = recommendation?.profile_summary;

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
              根据你的投资偏好、风险承受能力、当前持仓和历史交易，实时生成更适合优先关注的股票建议。
            </Paragraph>
          </div>
          <Button ghost icon={<ReloadOutlined />} onClick={() => void load()} loading={loading} style={{ borderRadius: 10 }}>
            重新生成
          </Button>
        </div>
      </Card>

      <Spin spinning={loading}>
        {error && recommendation ? (
          <Card bordered={false} style={{ ...cardStyle, marginBottom: 16 }}>
            <Alert type="warning" showIcon message={error} />
          </Card>
        ) : null}

        {error && !recommendation && !candidateInfo ? (
          <Card bordered={false} style={cardStyle}><Alert type="error" showIcon message={error} /></Card>
        ) : !recommendation && !candidateInfo ? (
          <Card bordered={false} style={cardStyle}><Empty description="暂无推荐结果" /></Card>
        ) : recommendation && !recommendation.candidates.length ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description={recommendation.summary_text || '暂无可用推荐，请先补充交易数据'} />
            <div style={{ marginTop: 16 }}>
              <Button type="primary" onClick={() => navigate('/app/upload')}>去导入交易记录</Button>
            </div>
          </Card>
        ) : recommendation ? (
          <>
            <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="候选池数量" value={profileSummary?.candidate_count ?? candidateInfo?.candidates.length ?? 0} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="当前持仓数" value={profileSummary?.held_positions ?? 0} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="累计盈亏" value={toNumber(profileSummary?.total_profit)} precision={2} prefix="¥" valueStyle={{ color: '#52c41a' }} /></Card></Col>
              <Col xs={12} lg={6}><Card bordered={false} style={cardStyle}><Statistic title="首选评分" value={toNumber(topCandidate?.score)} precision={2} valueStyle={{ color: '#1677ff' }} /></Card></Col>
            </Row>

            <Card bordered={false} style={cardStyle} title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />AI 总结</span>}>
              <Paragraph style={{ marginBottom: 12 }}>{recommendation.summary_text}</Paragraph>
              <Space wrap>
                <Tag color="processing" icon={<SafetyCertificateOutlined />}>风险承受：{profileSummary?.risk_tolerance || '—'}</Tag>
                <Tag color="blue" icon={<ThunderboltOutlined />}>投资偏好：{preferenceMap[profileSummary?.investment_preference || ''] ?? profileSummary?.investment_preference ?? '—'}</Tag>
                <Tag>生成时间：{recommendation.generated_at}</Tag>
              </Space>
            </Card>

            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
              {recommendation.candidates.map((item) => {
                const meta = actionMeta(item.action);
                return (
                  <Col span={24} lg={12} key={item.symbol}>
                    <Card bordered={false} style={cardStyle}>
                      <Space direction="vertical" size={10} style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'flex-start' }}>
                          <div>
                            <Title level={4} style={{ margin: 0 }}>{item.asset_name || item.symbol}</Title>
                            <Text type="secondary">{item.symbol}</Text>
                          </div>
                          <Space wrap>
                            <Tag color={meta.color}>{meta.text}</Tag>
                            <Tag color={item.is_held ? 'success' : 'default'}>{item.is_held ? '已持仓' : '候选观察'}</Tag>
                          </Space>
                        </div>

                        <Row gutter={[12, 12]}>
                          <Col span={8}><Statistic title="评分" value={toNumber(item.score)} precision={2} valueStyle={{ color: '#1677ff', fontSize: 22 }} /></Col>
                          <Col span={8}><Statistic title="最新价" value={toNumber(item.latest_price)} precision={2} prefix="¥" valueStyle={{ fontSize: 22 }} /></Col>
                          <Col span={8}><Statistic title="涨跌幅" value={toNumber(item.change_percent)} precision={2} suffix="%" valueStyle={{ color: toNumber(item.change_percent) >= 0 ? '#52c41a' : '#ff4d4f', fontSize: 22 }} /></Col>
                        </Row>

                        <div>
                          <Text strong>推荐理由</Text>
                          <Paragraph style={{ margin: '6px 0 0' }}>{item.match_reason}</Paragraph>
                        </div>
                        <div>
                          <Text strong>风险提示</Text>
                          <Paragraph style={{ margin: '6px 0 0' }}>{item.risk_note}</Paragraph>
                        </div>
                        <Space wrap>
                          <Tag>数据状态：{item.data_status}</Tag>
                          <Tag>历史交易关注：{item.trade_count} 次</Tag>
                          <Button type="link" onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(item.symbol)}`)} style={{ paddingInline: 0 }}>
                            查看市场趋势
                          </Button>
                        </Space>
                      </Space>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          </>
        ) : (
          <Card bordered={false} style={cardStyle}>
            <Empty description="AI 解释生成较慢，当前先展示候选股票池。" />
            <div style={{ marginTop: 16 }}>
              <Space wrap>
                <Tag color="processing">候选池数量 {candidateInfo?.candidates.length ?? 0}</Tag>
                {candidateInfo?.default_symbol ? <Tag color="blue">默认关注 {candidateInfo.default_symbol}</Tag> : null}
              </Space>
            </div>
          </Card>
        )}
      </Spin>
    </div>
  );
}
