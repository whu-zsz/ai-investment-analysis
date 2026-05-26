import { useEffect, useMemo, useState } from 'react';
import type { CSSProperties } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Collapse,
  Empty,
  Input,
  List,
  Row,
  Space,
  Spin,
  Statistic,
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
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { analysisApi, marketApi } from '../api';
import type {
  AnalysisCandidateResponse,
  ChatContextSnapshotResponse,
  ChatStepContextResponse,
  RecommendationChatResponse,
  StockChatMessageResponse,
  StockChatNewsItemResponse,
  StockChatResponse,
  BoardChatResponse,
  StockChatStreamEvent,
} from '../api/types';

const { Title, Paragraph, Text, Link } = Typography;
const { TextArea } = Input;

type StepStatus = 'wait' | 'process' | 'finish' | 'error';

const panelStyle: CSSProperties = {
  borderRadius: 18,
  boxShadow: '0 18px 40px rgba(15, 23, 42, 0.08)',
};

const markdownBodyStyle: CSSProperties = {
  fontSize: 14,
  lineHeight: 1.8,
};

const stageLabels: Record<string, string> = {
  planning: '规划',
  profile: '画像',
  holdings: '持仓',
  news: '新闻',
  quote: '资料',
  trend: '趋势',
  board: '板块',
  ranking: '排序',
  report: '生成',
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
    case 'board':
      return '板块';
    case 'recommendation':
      return '推荐对话';
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

function chatDataFromStreamData(data: StockChatStreamEvent['data']) {
  if (data && !Array.isArray(data) && 'messages' in data) {
    return data as StockChatResponse | BoardChatResponse | RecommendationChatResponse;
  }
  return null;
}

function contextDataFromStreamData(data: StockChatStreamEvent['data']) {
  if (data && !Array.isArray(data) && 'label' in data && 'stage' in data) {
    return data as ChatStepContextResponse;
  }
  return null;
}

function applyContextSnapshot(kind: 'stock' | 'board' | 'recommendation', snapshot: ChatContextSnapshotResponse, fallbackName: string, fallbackMarket: string): StockChatResponse | BoardChatResponse | RecommendationChatResponse {
  const base = {
    context_id: snapshot.context_id,
    asset_name: fallbackName,
    market: fallbackMarket,
    reply: snapshot.reply || '',
    ai_model: '',
    generated_at: snapshot.generated_at || '',
    news_status: snapshot.news_items?.length ? 'partial' : '',
    news_summary: '',
    news_coverage: '',
    news_items: snapshot.news_items ?? [],
    messages: snapshot.messages ?? [],
    tool_trace: snapshot.tool_trace ?? [],
    snapshot: {
      last_price: '0',
      change_percent: '0',
      high_price: '',
      low_price: '',
      volume: '',
      turnover: '',
      source: '',
      fetched_at: '',
      period: kind === 'board' ? 'board' : kind === 'recommendation' ? 'recommendation' : 'day',
      trend_summary: '',
    },
  };
  if (kind === 'board') {
    return {
      ...base,
      board_type: '',
      code: '',
    } as BoardChatResponse;
  }
  if (kind === 'recommendation') {
    return {
      ...base,
      context: {
        step_key: snapshot.report_id ? 'done' : 'clarify',
        step_label: snapshot.report_id ? '报告生成完成' : '继续补充偏好',
        profile_summary: {
          investment_preference: '',
          risk_tolerance: '',
          total_profit: '0.00',
          held_positions: 0,
          candidate_count: 0,
        },
        candidate_count: 0,
        discovery_count: snapshot.news_items?.length ?? 0,
        held_count: 0,
        focus_summary: snapshot.step_context?.summary || '',
      },
      report_id: snapshot.report_id || 0,
      report_title: snapshot.report_title || '',
      candidates: [],
      step_context: snapshot.step_context,
    } as RecommendationChatResponse;
  }
  return {
    ...base,
    symbol: '',
  } as StockChatResponse;
}

function renderChatMarkdown(content: string, isAssistant: boolean) {
  if (!isAssistant) {
    return <Paragraph style={{ marginBottom: 0, whiteSpace: 'pre-wrap', color: 'inherit' }}>{content}</Paragraph>;
  }

  return (
    <div style={markdownBodyStyle}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => <Title level={4} style={{ margin: '0 0 10px', color: 'inherit' }}>{children}</Title>,
          h2: ({ children }) => <Title level={5} style={{ margin: '14px 0 8px', color: '#31566f' }}>{children}</Title>,
          h3: ({ children }) => <Text strong style={{ display: 'block', margin: '12px 0 6px', color: '#31566f' }}>{children}</Text>,
          p: ({ children }) => <Paragraph style={{ margin: '0 0 10px', color: 'inherit' }}>{children}</Paragraph>,
          ul: ({ children }) => <ul style={{ margin: '0 0 10px', paddingLeft: 18 }}>{children}</ul>,
          ol: ({ children }) => <ol style={{ margin: '0 0 10px', paddingLeft: 18 }}>{children}</ol>,
          li: ({ children }) => <li style={{ marginBottom: 6 }}>{children}</li>,
          strong: ({ children }) => <strong style={{ color: '#173142' }}>{children}</strong>,
          blockquote: ({ children }) => (
            <div style={{ margin: '10px 0', padding: '10px 12px', borderRadius: 12, background: '#f4f8f5', borderLeft: '3px solid #6f9b8f' }}>
              {children}
            </div>
          ),
          a: ({ href, children }) => <Link href={href} target="_blank">{children}</Link>,
          code: ({ children }) => <code style={{ padding: '2px 6px', borderRadius: 6, background: '#edf4ef', color: '#204050' }}>{children}</code>,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

export default function StockChatPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const rawKind = searchParams.get('kind');
  const chatKind = rawKind === 'board' ? 'board' : rawKind === 'recommendation' ? 'recommendation' : 'stock';
  const boardType = searchParams.get('boardType') || '';
  const boardCode = searchParams.get('code') || '';
  const boardName = searchParams.get('name') || '';
  const recommendationReportId = Number(searchParams.get('reportId') || '0');
  const chatContextId = Number(searchParams.get('contextId') || '0');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [question, setQuestion] = useState('');
  const [candidates, setCandidates] = useState<AnalysisCandidateResponse[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState(searchParams.get('symbol') || '');
  const [chat, setChat] = useState<StockChatResponse | BoardChatResponse | RecommendationChatResponse | null>(null);
  const [detailName, setDetailName] = useState('');
  const [streamText, setStreamText] = useState('');
  const [currentStage, setCurrentStage] = useState('planning');
  const [stageMessage, setStageMessage] = useState('等待提问');
  const [stepStatus, setStepStatus] = useState<StepStatus>('wait');
  const [latestNews, setLatestNews] = useState<StockChatNewsItemResponse[]>([]);
  const [stepContext, setStepContext] = useState<ChatStepContextResponse | null>(null);

  const applyHeartbeat = (event: StockChatStreamEvent) => {
    setCurrentStage(event.stage || 'planning');
    setStageMessage(event.message || '正在持续处理当前步骤，请稍候');
    setStepContext((prev) => (prev ? { ...prev, summary: '' } : prev));
    setStepStatus('process');
  };

  const resetConversationState = () => {
    setChat(null);
    setLatestNews([]);
    setStreamText('');
    setQuestion('');
    setError('');
    setStepContext(null);
    setStepStatus('wait');
    setCurrentStage('planning');
    setStageMessage('等待提问');
  };

  const loadCandidates = async () => {
    setLoading(true);
    setError('');
    try {
      if (chatKind === 'board') {
        setCandidates([]);
        setSelectedSymbol('');
        setDetailName(boardName || boardCode);
        return;
      }
      if (chatKind === 'recommendation') {
        setCandidates([]);
        setSelectedSymbol('');
        setDetailName('AI 推荐对话');
        return;
      }
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
    resetConversationState();
    void loadCandidates();
  }, [chatKind, boardType, boardCode, boardName]);

  const candidateList = useMemo(
    () => [...candidates].sort((a, b) => Number(b.is_held) - Number(a.is_held) || (b.trade_count ?? 0) - (a.trade_count ?? 0)),
    [candidates],
  );

  const selectedCandidate = useMemo(
    () => candidateList.find((item) => item.symbol === selectedSymbol) ?? null,
    [candidateList, selectedSymbol],
  );

  useEffect(() => {
    const restore = async () => {
      if (chatKind === 'recommendation' && !chatContextId && recommendationReportId > 0) {
        try {
          const snapshot = await analysisApi.getRecommendationChatContextByReportId(recommendationReportId);
          if (snapshot.context_id) {
            setSearchParams((prev) => {
              const next = new URLSearchParams(prev);
              next.set('kind', 'recommendation');
              next.set('reportId', String(recommendationReportId));
              next.set('contextId', String(snapshot.context_id));
              return next;
            });
          }
        } catch (err: any) {
          setError(err?.message ?? err?.data?.message ?? '通过报告恢复推荐上下文失败');
        }
        return;
      }
      if (!chatContextId) return;
      try {
        const snapshot = await analysisApi.getChatContext(chatContextId);
        const fallbackName = chatKind === 'board' ? (boardName || boardCode || '板块 AI 助手') : chatKind === 'recommendation' ? 'AI 推荐对话' : (detailName || selectedCandidate?.asset_name || selectedSymbol || '个股 AI 助手');
        const fallbackMarket = chatKind === 'board' ? 'board' : chatKind === 'recommendation' ? 'recommendation' : (selectedCandidate?.market || 'cn_stock');
        setChat(applyContextSnapshot(chatKind, snapshot, fallbackName, fallbackMarket));
        setLatestNews(snapshot.news_items ?? []);
        setStepContext(snapshot.step_context ?? null);
        setStageMessage(snapshot.step_context?.summary || '已恢复历史上下文');
        setCurrentStage(snapshot.report_id ? 'done' : 'planning');
        setStepStatus('wait');
      } catch (err: any) {
        setError(err?.message ?? err?.data?.message ?? '恢复聊天上下文失败');
      }
    };
    void restore();
  }, [chatContextId, chatKind, boardCode, boardName, detailName, recommendationReportId, selectedCandidate?.asset_name, selectedCandidate?.market, selectedSymbol, setSearchParams]);

  const messageList: StockChatMessageResponse[] = chat?.messages ?? [];
  const visibleMessages: StockChatMessageResponse[] = sending || refreshing
    ? [...messageList, { role: 'assistant', content: streamText || '正在准备回答...' }]
    : messageList;
  const newsList = latestNews.length ? latestNews : chat?.news_items ?? [];
  const newsStatusMeta = getNewsStatusMeta(chat?.news_status);
  const highlights = extractHighlights(streamText || chat?.reply || '');
  const snapshot = chat?.snapshot;
  const canAsk = chatKind === 'board' ? Boolean(boardType && boardCode) : chatKind === 'recommendation' ? true : Boolean(selectedSymbol);

  const submitQuestion = async (text: string, refreshMarket = false) => {
    const trimmed = text.trim();
    if (chatKind === 'recommendation' && !trimmed) {
      message.warning('请输入你的推荐问题');
      return;
    }
    if (chatKind === 'board' && (!boardType || !boardCode)) {
      message.warning('请先指定板块');
      return;
    }
    if (chatKind === 'stock' && !selectedSymbol) {
      message.warning('请先选择或在 URL 中指定股票代码');
      return;
    }
    if (!trimmed) {
      message.warning('请输入问题');
      return;
    }

    const previousMessages = messageList.map((item) => ({ role: item.role, content: item.content }));
    const optimisticMessages: StockChatMessageResponse[] = [...messageList, { role: 'user', content: trimmed }];
    setChat((prev) => prev ? { ...prev, messages: optimisticMessages, reply: prev.reply } : (
      chatKind === 'board'
        ? {
            context_id: chatContextId || 0,
            board_type: boardType,
            code: boardCode,
            asset_name: detailName || boardName || boardCode,
            market: 'board',
            reply: '',
            ai_model: '',
            generated_at: '',
            news_status: '',
            news_summary: '',
            news_coverage: '',
            news_items: [],
            snapshot: {
              last_price: '0',
              change_percent: '0',
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
          } satisfies BoardChatResponse
        : {
            context_id: chatContextId || 0,
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
          } satisfies StockChatResponse
    ));
    setStreamText('');
    setLatestNews([]);
    setStepContext(null);
    setSending(!refreshMarket);
    setRefreshing(refreshMarket);
    setError('');
    setCurrentStage('planning');
    setStageMessage('正在规划工具步骤');
    setStepStatus('process');

    try {
      if (chatKind === 'board') {
        await analysisApi.boardChatStream(
          {
            board_type: boardType,
            code: boardCode,
            question: trimmed,
            messages: previousMessages,
            context_id: chatContextId || undefined,
          },
          (event) => {
            if (event.type === 'step') {
              setCurrentStage(event.stage || 'planning');
              setStageMessage(event.message || '正在处理');
              setStepStatus('process');
            }
            if (event.type === 'heartbeat') {
              applyHeartbeat(event);
            }
            if (event.type === 'context' && Array.isArray(event.data)) {
              setLatestNews(event.data);
              setStageMessage(event.message || '已获取新闻上下文');
            }
            if (event.type === 'token' && event.token) {
              setCurrentStage('report');
              setStreamText((prev) => prev + event.token);
            }
            if (event.type === 'done') {
              const finalChat = chatDataFromStreamData(event.data);
              if (finalChat) {
                setChat(finalChat);
                setDetailName(finalChat.asset_name || detailName);
                setLatestNews(finalChat.news_items ?? []);
                if ('context_id' in finalChat && finalChat.context_id) {
                  setSearchParams((prev) => {
                    const next = new URLSearchParams(prev);
                    next.set('kind', 'board');
                    next.set('boardType', boardType);
                    next.set('code', boardCode);
                    if (boardName) next.set('name', boardName);
                    next.set('contextId', String(finalChat.context_id));
                    return next;
                  });
                }
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
      } else if (chatKind === 'recommendation') {
        await analysisApi.recommendationChatStream(
          {
            question: trimmed,
            messages: previousMessages,
            report_id: recommendationReportId || undefined,
            context_id: chatContextId || undefined,
          },
          (event) => {
            if (event.type === 'step') {
              setCurrentStage(event.stage || 'planning');
              setStageMessage(event.message || '正在处理');
              setStepStatus('process');
            }
            if (event.type === 'heartbeat') {
              applyHeartbeat(event);
            }
            if (event.type === 'context') {
              if (Array.isArray(event.data)) {
                setLatestNews(event.data);
              } else {
                const contextData = contextDataFromStreamData(event.data);
                if (contextData) {
                  setStepContext(contextData);
                  if (contextData.news_items?.length) {
                    setLatestNews(contextData.news_items);
                  }
                  setStageMessage(event.message || contextData.summary || '已更新当前工具结果');
                }
              }
            }
            if (event.type === 'token' && event.token) {
              setCurrentStage('report');
              setStreamText((prev) => prev + event.token);
            }
            if (event.type === 'done') {
              const finalChat = chatDataFromStreamData(event.data);
              if (finalChat) {
                setChat(finalChat);
                setDetailName(finalChat.asset_name || 'AI 推荐对话');
                setLatestNews(finalChat.news_items ?? []);
                if ('context_id' in finalChat && finalChat.context_id) {
                  setSearchParams((prev) => {
                    const next = new URLSearchParams(prev);
                    next.set('kind', 'recommendation');
                    next.set('contextId', String(finalChat.context_id));
                    if ('report_id' in finalChat && finalChat.report_id) {
                      next.set('reportId', String(finalChat.report_id));
                    }
                    return next;
                  });
                }
              }
              setCurrentStage('done');
              setStageMessage(event.message || '推荐报告已生成');
              setStepStatus('finish');
              setQuestion('');
            }
            if (event.type === 'error') {
              throw new Error(event.message || 'AI 推荐失败');
            }
          },
        );
      } else {
        await analysisApi.stockChatStream(
          {
            symbol: selectedSymbol,
            question: trimmed,
            messages: previousMessages,
            context_id: chatContextId || undefined,
            refresh_market: refreshMarket,
          },
          (event) => {
            if (event.type === 'step') {
              setCurrentStage(event.stage || 'planning');
              setStageMessage(event.message || '正在处理');
              setStepStatus('process');
            }
            if (event.type === 'heartbeat') {
              applyHeartbeat(event);
            }
            if (event.type === 'context' && Array.isArray(event.data)) {
              setLatestNews(event.data);
              setStageMessage(event.message || '已获取新闻上下文');
            }
            if (event.type === 'token' && event.token) {
              setCurrentStage('report');
              setStreamText((prev) => prev + event.token);
            }
            if (event.type === 'done') {
              const finalChat = chatDataFromStreamData(event.data);
              if (finalChat) {
                setChat(finalChat);
                setDetailName(finalChat.asset_name || detailName);
                setLatestNews(finalChat.news_items ?? []);
                if ('context_id' in finalChat && finalChat.context_id) {
                  setSearchParams((prev) => {
                    const next = new URLSearchParams(prev);
                    if (selectedSymbol) next.set('symbol', selectedSymbol);
                    next.set('contextId', String(finalChat.context_id));
                    return next;
                  });
                }
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
      }
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
    resetConversationState();
    try {
      const detail = await marketApi.getStockDetail(symbol);
      setDetailName(detail.name || '');
    } catch {
      setDetailName(candidateList.find((item) => item.symbol === symbol)?.asset_name ?? '');
    }
  };

  const displayName = detailName || boardName || selectedCandidate?.asset_name || selectedSymbol || (chatKind === 'board' ? '板块 AI 助手' : chatKind === 'recommendation' ? '推荐 AI 助手' : '个股 AI 助手');
  const displayMarket = chat?.market || (chatKind === 'board' ? 'board' : chatKind === 'recommendation' ? 'recommendation' : (selectedCandidate?.market || 'cn_stock'));
  const displayPrice = snapshot?.last_price || selectedCandidate?.last_price;
  const displayChange = snapshot?.change_percent || selectedCandidate?.change_percent;
  const isStreaming = sending || refreshing;

  return (
    <div style={{ minHeight: '100vh', padding: 24, background: 'linear-gradient(135deg, #f5f7ef 0%, #eef5f2 48%, #f7f0e6 100%)' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate(chatKind === 'board'
          ? `/app/board?type=${encodeURIComponent(boardType)}&code=${encodeURIComponent(boardCode)}`
          : chatKind === 'recommendation'
            ? '/app/recommendation'
          : `/app/market-trend?symbol=${encodeURIComponent(selectedSymbol || searchParams.get('symbol') || '')}`)}
        style={{ marginBottom: 16, color: '#31566f', paddingLeft: 0 }}
      >
        {chatKind === 'board' ? '返回板块页' : chatKind === 'recommendation' ? '返回推荐页' : '返回行情页'}
      </Button>

      <div style={{ marginBottom: 18, padding: '22px 24px', borderRadius: 18, background: '#123a4a', color: '#fff', boxShadow: '0 18px 44px rgba(18, 58, 74, 0.22)' }}>
        <Row gutter={[18, 18]} align="middle">
          <Col span={24} lg={14}>
            <Space size={8} wrap style={{ marginBottom: 10 }}>
              <Tag color="gold">实时问答</Tag>
              <Tag color="cyan">{marketLabel(displayMarket)}</Tag>
              {chatKind === 'stock' ? <Tag color={selectedCandidate?.is_held ? 'green' : 'default'}>{selectedCandidate?.is_held ? '当前持仓' : '候选标的'}</Tag> : chatKind === 'board' ? <Tag color="purple">板块上下文</Tag> : <Tag color="magenta">推荐对话模式</Tag>}
              <Tag color={newsStatusMeta.color}>{newsStatusMeta.text}</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>{displayName}</Title>
            <Text style={{ color: 'rgba(255,255,255,0.72)', fontSize: 16 }}>{chatKind === 'board' ? `${boardType || 'board'} · ${boardCode}` : chatKind === 'recommendation' ? `recommendation${recommendationReportId ? ` · report ${recommendationReportId}` : ''}${chatContextId ? ` · context ${chatContextId}` : ''}` : (selectedSymbol || '未选择标的')}</Text>
            {snapshot?.trend_summary ? <Paragraph style={{ color: 'rgba(255,255,255,0.78)', margin: '10px 0 0' }}>{snapshot.trend_summary}</Paragraph> : null}
          </Col>
          <Col span={24} lg={10}>
            {chatKind === 'stock' ? (
              <Row gutter={[10, 10]}>
                <Col span={12}>
                  <Statistic title={<span style={{ color: 'rgba(255,255,255,0.7)' }}>最新价</span>} value={toNumber(displayPrice)} precision={2} prefix="¥" valueStyle={{ color: '#fff' }} />
                </Col>
                <Col span={12}>
                  <Statistic title={<span style={{ color: 'rgba(255,255,255,0.7)' }}>涨跌幅</span>} value={toNumber(displayChange)} precision={2} suffix="%" valueStyle={{ color: getChangeColor(displayChange) }} />
                </Col>
              </Row>
            ) : (
              <div style={{ padding: '12px 14px', borderRadius: 16, background: 'rgba(255,255,255,0.1)', border: '1px solid rgba(255,255,255,0.14)' }}>
                <Text style={{ color: 'rgba(255,255,255,0.82)', fontSize: 14, lineHeight: 1.7 }}>
                  {chatKind === 'board' ? '当前为板块对话，会结合成分股强弱、板块热度和相关新闻生成回答。' : '当前为推荐对话，会结合用户画像、已有关注和新闻潜力股生成推荐报告。'}
                </Text>
              </div>
            )}
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
            {chatKind === 'stock' ? <Col span={24} lg={6}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Card bordered={false} style={panelStyle} title="切换标的" extra={<Text type="secondary">持仓优先</Text>}>
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

                <Card bordered={false} style={panelStyle} title="新闻参考" extra={chat?.generated_at ? <Text type="secondary">{chat.generated_at}</Text> : null}>
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
            </Col> : null}

            <Col span={24} lg={chatKind === 'stock' ? 18 : 24}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Card
                  bordered={false}
                  style={panelStyle}
                  title={<span><RobotOutlined style={{ color: '#31566f', marginRight: 8 }} />{chatKind === 'board' ? 'AI 板块对话' : chatKind === 'recommendation' ? 'AI 推荐对话' : 'AI 个股对话'}</span>}
                  extra={(
                    chatKind === 'stock' ? <Button icon={<ReloadOutlined />} loading={refreshing} disabled={sending} onClick={() => void submitQuestion(question || '请基于最新新闻和走势，更新一次当前判断。', true)}>
                      刷新后再问
                    </Button> : null
                  )}
                >
                  <Space direction="vertical" size={14} style={{ width: '100%' }}>
                    <Space wrap>
                    {(chatKind === 'recommendation'
                      ? [
                        '结合我的风险偏好和最近热点，当前更值得优先关注哪些股票？',
                        '如果我想找未来 1 到 3 个月的潜力股，你会怎么筛？',
                        '请区分我已有关注标的和你新发现的潜力股。',
                      ]
                      : starterQuestions
                    ).map((item) => (
                        <Button key={item} size="small" onClick={() => setQuestion(item)}>{item}</Button>
                      ))}
                    </Space>

                    <div style={{ minHeight: 420, maxHeight: 620, overflowY: 'auto', padding: 18, borderRadius: 18, background: 'linear-gradient(180deg, #f8faf8 0%, #f3f6f3 100%)', border: '1px solid #e7ece5' }}>
                      {visibleMessages.length ? (
                        <Space direction="vertical" size={16} style={{ width: '100%' }}>
                          {visibleMessages.map((item, index) => {
                            const isAssistant = item.role === 'assistant';
                            const isLatestAssistant = isAssistant && index === visibleMessages.length - 1;
                            const showInlineProgress = isLatestAssistant && (isStreaming || stepStatus === 'error');
                            const showNewsReference = isLatestAssistant && isAssistant && !isStreaming && stepStatus !== 'error' && newsList.length > 0;
                            return (
                              <div key={`${item.role}-${index}`} style={{ display: 'flex', justifyContent: isAssistant ? 'flex-start' : 'flex-end' }}>
                                <div style={{ maxWidth: '82%', display: 'flex', gap: 10, flexDirection: isAssistant ? 'row' : 'row-reverse', alignItems: 'flex-start' }}>
                                  <div style={{ width: 34, height: 34, borderRadius: 12, display: 'grid', placeItems: 'center', background: isAssistant ? '#dfe9e4' : '#31566f', color: isAssistant ? '#31566f' : '#fff', flex: '0 0 auto' }}>
                                    {isAssistant ? <RobotOutlined /> : <MessageOutlined />}
                                  </div>
                                  <div style={{ padding: '14px 16px', borderRadius: 18, background: isAssistant ? '#fff' : '#31566f', color: isAssistant ? '#21313a' : '#fff', boxShadow: '0 8px 20px rgba(15,23,42,0.08)' }}>
                                    {showNewsReference ? (
                                      <div style={{ marginBottom: 12 }}>
                                        <Collapse
                                          size="small"
                                          ghost
                                          items={[
                                            {
                                              key: 'news-reference',
                                              label: <Text strong style={{ color: '#31566f' }}>参考新闻</Text>,
                                              children: (
                                                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                                                  {newsList.map((newsItem, newsIndex) => (
                                                    <div
                                                      key={`${newsItem.title}-${newsIndex}`}
                                                      style={{ padding: '10px 12px', borderRadius: 12, background: '#f6faf7', border: '1px solid #e1ebe4' }}
                                                    >
                                                      <Space direction="vertical" size={4} style={{ width: '100%' }}>
                                                        <Text strong>{newsItem.title}</Text>
                                                        <Space wrap size={6}>
                                                          <Tag>{newsItem.source || newsItem.provider || '新闻源'}</Tag>
                                                          {newsItem.published_at ? <Tag icon={<ClockCircleOutlined />} color="blue">{newsItem.published_at}</Tag> : null}
                                                        </Space>
                                                        {newsItem.summary ? <Text type="secondary">{newsItem.summary}</Text> : null}
                                                        {newsItem.url ? <Link href={newsItem.url} target="_blank">查看原文</Link> : null}
                                                      </Space>
                                                    </div>
                                                  ))}
                                                </Space>
                                              ),
                                            },
                                          ]}
                                        />
                                      </div>
                                    ) : null}
                                    {showInlineProgress ? (
                                      <div style={{ marginBottom: item.content ? 12 : 0, padding: '12px 12px 10px', borderRadius: 14, background: '#f4f8f5', border: '1px solid #dce8df' }}>
                                        <Space direction="vertical" size={8} style={{ width: '100%' }}>
                                          <Text strong style={{ color: '#31566f' }}>AI 正在使用工具</Text>
                                          <Text type={stepStatus === 'error' ? 'danger' : 'secondary'}>{stageLabels[currentStage] || currentStage}</Text>
                                          <Text type="secondary">{stageMessage}</Text>
                                          {stepContext?.summary ? <Text type="secondary">{stepContext.summary}</Text> : null}
                                          {stepContext?.reference_symbols?.length ? <Text type="secondary">关联股票：{stepContext.reference_symbols.join('、')}</Text> : null}
                                          {stepContext?.reference_boards?.length ? <Text type="secondary">关联板块：{stepContext.reference_boards.join('、')}</Text> : null}
                                        </Space>
                                      </div>
                                    ) : null}
                                    {renderChatMarkdown(item.content, isAssistant)}
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
                      placeholder={chatKind === 'board' ? '例如：这个板块今天走强的核心驱动是什么，持续性如何？' : chatKind === 'recommendation' ? '例如：我偏好成长风格，能接受中等波动，想找未来 1-3 个月值得关注的潜力股。' : '例如：结合最近新闻和走势，这只股票当前更适合继续持有还是等待回调？'}
                    />
                    <Space>
                      <Button type="primary" icon={<SendOutlined />} loading={sending} disabled={refreshing} onClick={() => void submitQuestion(question)}>
                        发送问题
                      </Button>
                    </Space>
                  </Space>
                </Card>
              </Space>
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
}
