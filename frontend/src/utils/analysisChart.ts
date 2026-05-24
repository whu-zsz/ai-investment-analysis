import type {
  AnalysisReportDetailResponse,
  AnalysisReportItemResponse,
} from '../api/types';

export interface ProfitBySymbolPoint {
  symbol: string;
  value: string;
}

export type ProfitSemantic = 'positive' | 'negative' | 'zero';

export interface ProfitBySymbolViewPoint extends ProfitBySymbolPoint {
  numericValue: number;
  semantic: ProfitSemantic;
  semanticLabel: string;
  color: string;
}

export interface ProfitBySymbolSummary {
  topPoint: ProfitBySymbolViewPoint | null;
  bottomPoint: ProfitBySymbolViewPoint | null;
  positiveCount: number;
  negativeCount: number;
  zeroCount: number;
  totalCount: number;
}

export interface OutcomeDistributionPoint {
  key: ProfitSemantic;
  label: string;
  value: number;
  percent: number;
  color: string;
}

export interface OutcomeDistributionSummary {
  total: number;
  positiveCount: number;
  negativeCount: number;
  zeroCount: number;
  points: OutcomeDistributionPoint[];
}

export interface ProfitCompositionPoint {
  symbol: string;
  realizedProfit: number;
  unrealizedProfit: number;
  totalProfit: number;
}

export interface ProfitCompositionSummary {
  realizedTotal: number;
  unrealizedTotal: number;
  netTotal: number;
  points: ProfitCompositionPoint[];
}

export interface ChangePercentRankingPoint {
  symbol: string;
  value: string;
  numericValue: number;
  semantic: ProfitSemantic;
  semanticLabel: string;
  color: string;
}

interface ProfitBySymbolChartEnvelope {
  version?: number;
  kind?: string;
  metric?: string;
  points?: unknown;
}

function isPoint(value: unknown): value is ProfitBySymbolPoint {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const point = value as { symbol?: unknown; value?: unknown };
  return typeof point.symbol === 'string' && (typeof point.value === 'string' || typeof point.value === 'number');
}

function normalizePoints(points: unknown[]): ProfitBySymbolPoint[] {
  return points
    .filter(isPoint)
    .map((point) => ({
      symbol: point.symbol,
      value: String(point.value),
    }));
}

function buildFallbackProfitPoints(items?: AnalysisReportItemResponse[]): ProfitBySymbolPoint[] {
  if (!items?.length) {
    return [];
  }

  return items
    .filter((item) => item.symbol?.trim() && (item.realized_profit?.trim() || item.total_profit?.trim()))
    .map((item) => ({
      symbol: item.symbol,
      value: item.realized_profit?.trim() || item.total_profit,
    }));
}

export function parseProfitBySymbolChartData(chartData?: string, items?: AnalysisReportItemResponse[]): ProfitBySymbolPoint[] {
  if (chartData) {
    try {
      const parsed = JSON.parse(chartData) as unknown;

      if (Array.isArray(parsed)) {
        return normalizePoints(parsed);
      }

      if (typeof parsed === 'object' && parsed !== null) {
        const envelope = parsed as ProfitBySymbolChartEnvelope;
        if (envelope.kind === 'profit_by_symbol' && Array.isArray(envelope.points)) {
          return normalizePoints(envelope.points);
        }
      }
    } catch {
      return buildFallbackProfitPoints(items);
    }
  }

  return buildFallbackProfitPoints(items);
}

export function toNumericProfitValue(value?: string): number {
  if (!value) {
    return 0;
  }

  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function getProfitSemantic(value: number): ProfitSemantic {
  if (value > 0) {
    return 'positive';
  }
  if (value < 0) {
    return 'negative';
  }
  return 'zero';
}

export function getProfitSemanticLabel(semantic: ProfitSemantic): string {
  if (semantic === 'positive') {
    return '盈利';
  }
  if (semantic === 'negative') {
    return '亏损';
  }
  return '持平';
}

export function getProfitSemanticColor(semantic: ProfitSemantic): string {
  if (semantic === 'positive') {
    return '#52c41a';
  }
  if (semantic === 'negative') {
    return '#ff4d4f';
  }
  return '#8c8c8c';
}

export function buildProfitBySymbolViewData(chartData?: string, items?: AnalysisReportItemResponse[]): ProfitBySymbolViewPoint[] {
  return parseProfitBySymbolChartData(chartData, items)
    .map((point) => {
      const numericValue = toNumericProfitValue(point.value);
      const semantic = getProfitSemantic(numericValue);
      return {
        ...point,
        numericValue,
        semantic,
        semanticLabel: getProfitSemanticLabel(semantic),
        color: getProfitSemanticColor(semantic),
      };
    })
    .sort((left, right) => right.numericValue - left.numericValue);
}

export function summarizeProfitBySymbolData(points: ProfitBySymbolViewPoint[]): ProfitBySymbolSummary {
  const positiveCount = points.filter((point) => point.semantic === 'positive').length;
  const negativeCount = points.filter((point) => point.semantic === 'negative').length;
  const zeroCount = points.filter((point) => point.semantic === 'zero').length;

  return {
    topPoint: points[0] ?? null,
    bottomPoint: points.length ? points[points.length - 1] : null,
    positiveCount,
    negativeCount,
    zeroCount,
    totalCount: points.length,
  };
}

export function buildOutcomeDistributionData(report?: Pick<AnalysisReportDetailResponse, 'symbols_count' | 'winning_trades' | 'losing_trades' | 'chart_data' | 'items'> | null): OutcomeDistributionSummary {
  const fallbackPoints = buildProfitBySymbolViewData(report?.chart_data, report?.items);
  const fallbackSummary = summarizeProfitBySymbolData(fallbackPoints);
  const positiveCount = report?.winning_trades || fallbackSummary.positiveCount;
  const negativeCount = report?.losing_trades || fallbackSummary.negativeCount;
  const inferredTotal = report?.symbols_count || fallbackSummary.totalCount;
  const zeroCount = Math.max(inferredTotal - positiveCount - negativeCount, fallbackSummary.zeroCount, 0);
  const total = positiveCount + negativeCount + zeroCount;

  if (!total) {
    return {
      total: 0,
      positiveCount: 0,
      negativeCount: 0,
      zeroCount: 0,
      points: [],
    };
  }

  const counts: Array<{ key: ProfitSemantic; value: number }> = [
    { key: 'positive', value: positiveCount },
    { key: 'negative', value: negativeCount },
    { key: 'zero', value: zeroCount },
  ];

  return {
    total,
    positiveCount,
    negativeCount,
    zeroCount,
    points: counts.map((item) => ({
      key: item.key,
      label: getProfitSemanticLabel(item.key),
      value: item.value,
      percent: total ? Number(((item.value / total) * 100).toFixed(1)) : 0,
      color: getProfitSemanticColor(item.key),
    })),
  };
}

export function buildProfitCompositionData(items?: AnalysisReportItemResponse[], limit = 6): ProfitCompositionSummary {
  if (!items?.length) {
    return {
      realizedTotal: 0,
      unrealizedTotal: 0,
      netTotal: 0,
      points: [],
    };
  }

  const normalized = items
    .map((item) => {
      const realizedProfit = toNumericProfitValue(item.realized_profit);
      const unrealizedProfit = toNumericProfitValue(item.unrealized_profit);
      const totalProfit = realizedProfit + unrealizedProfit;
      return {
        symbol: item.symbol,
        realizedProfit,
        unrealizedProfit,
        totalProfit,
      } satisfies ProfitCompositionPoint;
    })
    .filter((item) => item.symbol?.trim() && (item.realizedProfit !== 0 || item.unrealizedProfit !== 0))
    .sort((left, right) => Math.abs(right.totalProfit) - Math.abs(left.totalProfit));

  const realizedTotal = normalized.reduce((sum, item) => sum + item.realizedProfit, 0);
  const unrealizedTotal = normalized.reduce((sum, item) => sum + item.unrealizedProfit, 0);

  return {
    realizedTotal,
    unrealizedTotal,
    netTotal: realizedTotal + unrealizedTotal,
    points: normalized.slice(0, limit),
  };
}

export function buildChangePercentRankingData(items?: AnalysisReportItemResponse[], limit = 6): ChangePercentRankingPoint[] {
  if (!items?.length) {
    return [];
  }

  return items
    .map((item) => {
      const numericValue = toNumericProfitValue(item.change_percent_7d);
      const semantic = getProfitSemantic(numericValue);
      return {
        symbol: item.symbol,
        value: item.change_percent_7d,
        numericValue,
        semantic,
        semanticLabel: getProfitSemanticLabel(semantic),
        color: getProfitSemanticColor(semantic),
      } satisfies ChangePercentRankingPoint;
    })
    .filter((item) => item.symbol?.trim() && item.value?.trim())
    .sort((left, right) => right.numericValue - left.numericValue)
    .slice(0, limit);
}

export function formatProfitValue(value: number): string {
  if (!Number.isFinite(value)) {
    return '0.00';
  }
  return value.toFixed(2);
}
