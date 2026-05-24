import { useEffect, useMemo, useState } from 'react';
import type { CSSProperties } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Empty,
  Input,
  List,
  Row,
  Space,
  Spin,
  Statistic,
  Steps,
  Tag,
  Typography,
  message,
} from 'antd';
import {
  ArrowLeftOutlined,
  ClockCircleOutlined,
  MessageOutlined,
  ReloadOutlined,
  RobotOutlined,
  SendOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { analysisApi, marketApi } from '../api';
import type {
  AnalysisCandidateResponse,
  StockChatMessageResponse,
  StockChatNewsItemResponse,
  StockChatResponse,
  StockChatStreamEvent,
} from '../api/types';

const { Title, Paragraph, Text, Link } = Typography;
const { TextArea } = Input;

type StepStatus = 'wait' | 'process' | 'finish' | 'error';

const panelStyle: CSSProperties = {
  borderRadius: 14,
  boxShadow: '0 14px 34px rgba(15, 23, 42, 0.08)',
};

const stageOrder = ['market', 'news', 'prompt', 'ai', 'done'];
const stageLabels: Record<string, string> = {
  market: '行情',
  news: '新闻',
  prompt: '上下文',
  ai: '生成',
  done: '完成',
};

const starterQuestions = [
  '结合最近新闻和走势，这只股票当前更适合继续持有还是等待回调？',
  '请指出现在最关键的利好、利空和风险点。',
  '如果我已经持仓，接下来三到五个交易日重点看哪些信号？',
];

function toNumber(value?: string) {
  const parsed = Number.parseFloat(value ?? '0');
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatPrice(value?: string) {
  return `¥${toNumber(value).toFixed(2)}`;
}

function formatPercent(value?: string) {
  const number = toNumber(value);
  const sign = number > 0 ? '+' : '';
  return `${sign}${number.toFixed(2)}%`;
}

function getChangeColor(value?: string) {
  const number = toNumber(value);
  if (number > 0) return '#c2410c';
  if (number < 0) return '#15803d';
  return '#31566f';
}

function marketLabel(market?: string) {
  switch (market) {
    case 'cn_index':
      return 'A 股指数';
    case 'cn_stock':
      return 'A 股个股';
    default:
      return market || '市场';
  }
}

function getNewsStatusMeta(status?: string) {
  if (status === 'complete') return { color: 'success' as const, text: '新闻覆盖完整' };
  if (status === 'partial') return { color: 'warning' as const, text: '新闻部分覆盖' };
  return { color: 'default' as const, text: '等待新闻上下文' };
}

function extractHighlights(text: string) {
  return text
    .split('\n')
    .map((line) => line.replace(/^[-*\d.\s]+/, '').trim())
    .filter((line) => line.startsWith('重点：') || line.startsWith('重点:'))
    .map((line) => line.replace(/^重点[：:]\s*/, '').trim())
    .filter(Boolean)
    .slice(0, 4);
}

function stepItems(activeStage: string, status: StepStatus) {
  const activeIndex = Math.max(0, stageOrder.indexOf(activeStage));
  return stageOrder.map((stage, index) => ({
    title: stageLabels[stage],
    status:
      status === 'error' && index === activeIndex
        ? ('error' as const)
        : index < activeIndex
          ? ('finish' as const)
          : index === activeIndex
            ? status
            : ('wait' as const),
  }));
}

function chatDataFromStreamData(data: StockChatStreamEvent['data']) {
  if (data && !Array.isArray(data) && 'messages' in data) {
    return data as StockChatResponse;
  }
  return null;
}

export default function StockChatPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [question, setQuestion] = useState('');
  const [candidates, setCandidates] = useState<AnalysisCandidateResponse[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState(searchParams.get('symbol') || '');
  const [chat, setChat] = useState<StockChatResponse | null>(null);
  const [detailName, setDetailName] = useState('');
  const [streamText, setStreamText] = useState('');
  const [currentStage, setCurrentStage] = useState('market');
  const [stageMessage, setStageMessage] = useState('等待提问');
  const [stepStatus, setStepStatus] = useState<StepStatus>('wait');
  const [latestNews, setLatestNews] = useState<StockChatNewsItemResponse[]>([]);

  const loadCandidates = async () => {
    setLoading(true);
    setError('');
    try {
      const candidateRes = await analysisApi.getCandidates();
      const candidateList = candidateRes.candidates ?? [];
      const querySymbol = searchParams.get('symbol') || '';
      const nextSymbol = querySymbol || candidateRes.default_symbol || candidateList[0]?.symbol || '';
      setCandidates(candidateList);
      setSelectedSymbol(nextSymbol);
      if (nextSymbol && searchParams.get('symbol') !== nextSymbol) {
        setSearchParams({ symbol: nextSymbol });
      }
      if (nextSymbol) {
        const fallbackName = candidateList.find((item) => item.symbol === nextSymbol)?.asset_name ?? '';
        try {
          const detail = await marketApi.getStockDetail(nextSymbol);
          setDetailName(detail.name || fallbackName);
        } catch {
          setDetailName(fallbackName);
        }
      }
    } catch (err: any) {
      setError(err?.message ?? err?.data?.message ?? 'AI 对话页加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadCandidates();
  }, []);

  const candidateList = useMemo(
    () => [...candidates].sort((a, b) => Number(b.is_held) - Number(a.is_held) || (b.trade_count ?? 0) - (a.trade_count ?? 0)),
    [candidates],
  );

  const selectedCandidate = useMemo(
    () => candidateList.find((item) => item.symbol === selectedSymbol) ?? null,
    [candidateList, selectedSymbol],
  );

  const messageList: StockChatMessageResponse[] = chat?.messages ?? [];
  const visibleMessages: StockChatMessageResponse[] = sending || refreshing
    ? [...messageList, { role: 'assistant', content: streamText || '正在准备回答...' }]
    : messageList;
  const newsList = latestNews.length ? latestNews : chat?.news_items ?? [];
  const newsStatusMeta = getNewsStatusMeta(chat?.news_status);
  const highlights = extractHighlights(streamText || chat?.reply || '');
  const snapshot = chat?.snapshot;
  const canAsk = Boolean(selectedSymbol);

  const submitQuestion = async (text: string, refreshMarket = false) => {
    const trimmed = text.trim();
    if (!selectedSymbol) {
      message.warning('请先选择或在 URL 中指定股票代码');
      return;
    }
    if (!trimmed) {
      message.warning('请输入问题');
      return;
    }

    const previousMessages = messageList.map((item) => ({ role: item.role, content: item.content }));
    const optimisticMessages: StockChatMessageResponse[] = [...messageList, { role: 'user', content: trimmed }];
    setChat((prev) => prev ? { ...prev, messages: optimisticMessages, reply: prev.reply } : {
      symbol: selectedSymbol,
      asset_name: detailName || selectedCandidate?.asset_name || selectedSymbol,
      market: selectedCandidate?.market || 'cn_stock',
      reply: '',
      ai_model: '',
      generated_at: '',
      news_status: '',
      news_summary: '',
      news_coverage: '',
      news_items: [],
      snapshot: {
        last_price: selectedCandidate?.last_price || '0',
        change_percent: selectedCandidate?.change_percent || '0',
        high_price: '',
        low_price: '',
        volume: '',
        turnover: '',
        source: '',
        fetched_at: '',
        period: '',
        trend_summary: '',
      },
      messages: optimisticMessages,
    });
    setStreamText('');
    setLatestNews([]);
    setSending(!refreshMarket);
    setRefreshing(refreshMarket);
    setError('');
    setCurrentStage('market');
    setStageMessage('正在获取最新行情和新闻');
    setStepStatus('process');

    try {
      await analysisApi.stockChatStream(
        {
          symbol: selectedSymbol,
          question: trimmed,
          messages: previousMessages,
          refresh_market: refreshMarket,
        },
        (event) => {
          if (event.type === 'step') {
            setCurrentStage(event.stage || 'market');
            setStageMessage(event.message || '正在处理');
            setStepStatus('process');
          }
          if (event.type === 'context' && Array.isArray(event.data)) {
            setLatestNews(event.data);
            setStageMessage(event.message || '已获取新闻上下文');
          }
          if (event.type === 'token' && event.token) {
            setCurrentStage('ai');
            setStreamText((prev) => prev + event.token);
          }
          if (event.type === 'done') {
            const finalChat = chatDataFromStreamData(event.data);
            if (finalChat) {
              setChat(finalChat);
              setDetailName(finalChat.asset_name || detailName);
              setLatestNews(finalChat.news_items ?? []);
            }
            setCurrentStage('done');
            setStageMessage(event.message || '回答生成完成');
            setStepStatus('finish');
            setQuestion('');
          }
          if (event.type === 'error') {
            throw new Error(event.message || 'AI 分析失败');
          }
        },
      );
    } catch (err: any) {
      const msg = err?.message ?? err?.data?.message ?? 'AI 分析失败';
      setError(msg);
      setStepStatus('error');
      setStageMessage(msg);
      message.error(msg);
      setChat((prev) => prev ? { ...prev, messages: messageList } : prev);
    } finally {
      setSending(false);
      setRefreshing(false);
    }
  };

  const handleSelectSymbol = async (symbol: string) => {
    setSelectedSymbol(symbol);
    setSearchParams({ symbol });
    setChat(null);
    setLatestNews([]);
    setStreamText('');
    setQuestion('');
    setError('');
    setStepStatus('wait');
    setStageMessage('等待提问');
    try {
      const detail = await marketApi.getStockDetail(symbol);
      setDetailName(detail.name || '');
    } catch {
      setDetailName(candidateList.find((item) => item.symbol === symbol)?.asset_name ?? '');
    }
  };

  const displayName = detailName || selectedCandidate?.asset_name || selectedSymbol || '个股 AI 助手';
  const displayMarket = chat?.market || selectedCandidate?.market || 'cn_stock';
  const displayPrice = snapshot?.last_price || selectedCandidate?.last_price;
  const displayChange = snapshot?.change_percent || selectedCandidate?.change_percent;

  return (
    <div style={{ minHeight: '100vh', padding: 24, background: 'linear-gradient(135deg, #f5f7ef 0%, #eef5f2 48%, #f7f0e6 100%)' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate(`/app/market-trend?symbol=${encodeURIComponent(selectedSymbol || searchParams.get('symbol') || '')}`)}
        style={{ marginBottom: 16, color: '#31566f', paddingLeft: 0 }}
      >
        返回行情页
      </Button>

      <div style={{ marginBottom: 18, padding: '22px 24px', borderRadius: 18, background: '#123a4a', color: '#fff', boxShadow: '0 18px 44px rgba(18, 58, 74, 0.22)' }}>
        <Row gutter={[18, 18]} align="middle">
          <Col span={24} lg={14}>
            <Space size={8} wrap style={{ marginBottom: 10 }}>
              <Tag color="gold">实时问答</Tag>
              <Tag color="cyan">{marketLabel(displayMarket)}</Tag>
              <Tag color={selectedCandidate?.is_held ? 'green' : 'default'}>{selectedCandidate?.is_held ? '当前持仓' : '候选标的'}</Tag>
              <Tag color={newsStatusMeta.color}>{newsStatusMeta.text}</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>{displayName}</Title>
            <Text style={{ color: 'rgba(255,255,255,0.72)', fontSize: 16 }}>{selectedSymbol || '未选择标的'}</Text>
            {snapshot?.trend_summary ? <Paragraph style={{ color: 'rgba(255,255,255,0.78)', margin: '10px 0 0' }}>{snapshot.trend_summary}</Paragraph> : null}
          </Col>
          <Col span={24} lg={10}>
            <Row gutter={[10, 10]}>
              <Col span={12}>
                <Statistic title={<span style={{ color: 'rgba(255,255,255,0.7)' }}>最新价</span>} value={toNumber(displayPrice)} precision={2} prefix="¥" valueStyle={{ color: '#fff' }} />
              </Col>
              <Col span={12}>
                <Statistic title={<span style={{ color: 'rgba(255,255,255,0.7)' }}>涨跌幅</span>} value={toNumber(displayChange)} precision={2} suffix="%" valueStyle={{ color: getChangeColor(displayChange) }} />
              </Col>
            </Row>
          </Col>
        </Row>
      </div>

      <Spin spinning={loading}>
        {error ? <Alert type="error" showIcon message={error} style={{ marginBottom: 16 }} /> : null}
        {!canAsk ? (
          <Card bordered={false} style={panelStyle}>
            <Empty description="暂无可分析标的，请在 URL 中指定 symbol，或先导入交易记录 / 生成持仓" />
          </Card>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={24} lg={7}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Card bordered={false} style={panelStyle} title="候选标的" extra={<Text type="secondary">持仓优先</Text>}>
                  {candidateList.length ? (
                    <List
                      dataSource={candidateList}
                      renderItem={(item) => (
                        <List.Item onClick={() => void handleSelectSymbol(item.symbol)} style={{ cursor: 'pointer', paddingInline: 0 }}>
                          <div style={{ width: '100%', padding: 12, borderRadius: 12, background: item.symbol === selectedSymbol ? '#e7f0ed' : '#fff', border: item.symbol === selectedSymbol ? '1px solid #6f9b8f' : '1px solid #edf0ed' }}>
                            <Space direction="vertical" size={4} style={{ width: '100%' }}>
                              <Space wrap>
                                <Text strong>{item.asset_name || item.symbol}</Text>
                                <Tag color={item.is_held ? 'success' : 'default'}>{item.is_held ? '已持仓' : '历史关注'}</Tag>
                              </Space>
                              <Text type="secondary">{item.symbol}</Text>
                              <Text style={{ color: getChangeColor(item.change_percent) }}>涨跌幅 {formatPercent(item.change_percent)}</Text>
                            </Space>
                          </div>
                        </List.Item>
                      )}
                    />
                  ) : (
                    <Alert type="info" showIcon message={`当前按 URL 标的 ${selectedSymbol} 直接分析`} />
                  )}
                </Card>

                <Card bordered={false} style={panelStyle} title={<span><ThunderboltOutlined /> 重点提示</span>}>
                  {highlights.length ? (
                    <Space direction="vertical" size={8} style={{ width: '100%' }}>
                      {highlights.map((item) => <Alert key={item} type="warning" showIcon message={item} />)}
                    </Space>
                  ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="提问后这里会提取 AI 的重点提示" />
                  )}
                </Card>

                <Card bordered={false} style={panelStyle} title="新闻上下文" extra={chat?.generated_at ? <Text type="secondary">{chat.generated_at}</Text> : null}>
                  {newsList.length ? (
                    <List
                      dataSource={newsList}
                      renderItem={(item) => (
                        <List.Item style={{ paddingInline: 0 }}>
                          <Space direction="vertical" size={4} style={{ width: '100%' }}>
                            <Text strong>{item.title}</Text>
                            <Space wrap size={6}>
                              <Tag>{item.source || item.provider}</Tag>
                              {item.published_at ? <Tag icon={<ClockCircleOutlined />} color="blue">{item.published_at}</Tag> : null}
                            </Space>
                            <Paragraph type="secondary" ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>{item.summary || '暂无摘要'}</Paragraph>
                            {item.url ? <Link href={item.url} target="_blank">查看原文</Link> : null}
                          </Space>
                        </List.Item>
                      )}
                    />
                  ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本轮分析使用的新闻会实时出现在这里" />
                  )}
                </Card>
              </Space>
            </Col>

            <Col span={24} lg={17}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Card
                  bordered={false}
                  style={panelStyle}
                  title={<span><RobotOutlined style={{ color: '#31566f', marginRight: 8 }} />AI 对话</span>}
                  extra={(
                    <Button icon={<ReloadOutlined />} loading={refreshing} disabled={sending} onClick={() => void submitQuestion(question || '请基于最新新闻和走势，更新一次当前判断。', true)}>
                      刷新后再问
                    </Button>
                  )}
                >
                  <Space direction="vertical" size={12} style={{ width: '100%' }}>
                    <Steps size="small" current={Math.max(0, stageOrder.indexOf(currentStage))} items={stepItems(currentStage, stepStatus)} />
                    <Alert type={stepStatus === 'error' ? 'error' : stepStatus === 'finish' ? 'success' : 'info'} showIcon message={stageMessage} />
                    <Space wrap>
                      {starterQuestions.map((item) => (
                        <Button key={item} size="small" onClick={() => setQuestion(item)}>{item}</Button>
                      ))}
                    </Space>

                    <div style={{ minHeight: 360, maxHeight: 560, overflowY: 'auto', padding: 16, borderRadius: 14, background: '#f7f8f4', border: '1px solid #e7ece5' }}>
                      {visibleMessages.length ? (
                        <Space direction="vertical" size={14} style={{ width: '100%' }}>
                          {visibleMessages.map((item, index) => {
                            const isAssistant = item.role === 'assistant';
                            return (
                              <div key={`${item.role}-${index}`} style={{ display: 'flex', justifyContent: isAssistant ? 'flex-start' : 'flex-end' }}>
                                <div style={{ maxWidth: '78%', display: 'flex', gap: 8, flexDirection: isAssistant ? 'row' : 'row-reverse', alignItems: 'flex-start' }}>
                                  <div style={{ width: 30, height: 30, borderRadius: 10, display: 'grid', placeItems: 'center', background: isAssistant ? '#dfe9e4' : '#31566f', color: isAssistant ? '#31566f' : '#fff', flex: '0 0 auto' }}>
                                    {isAssistant ? <RobotOutlined /> : <MessageOutlined />}
                                  </div>
                                  <div style={{ padding: '12px 14px', borderRadius: 14, background: isAssistant ? '#fff' : '#31566f', color: isAssistant ? '#21313a' : '#fff', boxShadow: '0 6px 18px rgba(15,23,42,0.08)' }}>
                                    <Paragraph style={{ marginBottom: 0, whiteSpace: 'pre-wrap', color: 'inherit' }}>{item.content}</Paragraph>
                                  </div>
                                </div>
                              </div>
                            );
                          })}
                        </Space>
                      ) : (
                        <Empty description="选择一个问题开始对话" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                      )}
                    </div>

                    <TextArea
                      value={question}
                      onChange={(event) => setQuestion(event.target.value)}
                      autoSize={{ minRows: 3, maxRows: 6 }}
                      disabled={sending || refreshing}
                      placeholder="例如：结合最近新闻和走势，这只股票当前更适合继续持有还是等待回调？"
                    />
                    <Space>
                      <Button type="primary" icon={<SendOutlined />} loading={sending} disabled={refreshing} onClick={() => void submitQuestion(question)}>
                        发送问题
                      </Button>
                    </Space>
                  </Space>
                </Card>

                <Card bordered={false} style={panelStyle} title="行情摘要">
                  {snapshot ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12} xl={6}><Statistic title="最新价" value={formatPrice(snapshot.last_price)} /></Col>
                      <Col span={12} xl={6}><Statistic title="涨跌幅" value={formatPercent(snapshot.change_percent)} valueStyle={{ color: getChangeColor(snapshot.change_percent) }} /></Col>
                      <Col span={12} xl={6}><Statistic title="成交量" value={snapshot.volume || '-'} /></Col>
                      <Col span={12} xl={6}><Statistic title="成交额" value={snapshot.turnover || '-'} /></Col>
                      <Col span={24}>
                        <Alert type="info" showIcon message={`数据源：${snapshot.source || '-'} · ${snapshot.fetched_at || '-'}`} description={snapshot.trend_summary} />
                      </Col>
                    </Row>
                  ) : (
                    <Empty description="对话后展示本轮使用的行情摘要" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                  )}
                </Card>
              </Space>
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
}
