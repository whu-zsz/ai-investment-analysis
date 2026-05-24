import { describe, it, expect } from 'vitest';
import type { AnalysisReportItemResponse } from '../api/types';
import {
  toNumericProfitValue,
  getProfitSemantic,
  getProfitSemanticLabel,
  getProfitSemanticColor,
  formatProfitValue,
  parseProfitBySymbolChartData,
  buildOutcomeDistributionData,
  buildProfitCompositionData,
  buildChangePercentRankingData,
} from '../utils/analysisChart';

describe('toNumericProfitValue', () => {
  it('正数解析', () => {
    expect(toNumericProfitValue('123.45')).toBe(123.45);
  });

  it('负数解析', () => {
    expect(toNumericProfitValue('-50.00')).toBe(-50);
  });

  it('零值解析', () => {
    expect(toNumericProfitValue('0')).toBe(0);
  });

  it('undefined 返回 0', () => {
    expect(toNumericProfitValue(undefined)).toBe(0);
  });

  it('空字符串返回 0', () => {
    expect(toNumericProfitValue('')).toBe(0);
  });

  it('非数字返回 0', () => {
    expect(toNumericProfitValue('abc')).toBe(0);
  });
});

describe('getProfitSemantic', () => {
  it('正数返回 positive', () => {
    expect(getProfitSemantic(100)).toBe('positive');
  });

  it('负数返回 negative', () => {
    expect(getProfitSemantic(-50)).toBe('negative');
  });

  it('零返回 zero', () => {
    expect(getProfitSemantic(0)).toBe('zero');
  });
});

describe('getProfitSemanticLabel', () => {
  it('positive 返回 盈利', () => {
    expect(getProfitSemanticLabel('positive')).toBe('盈利');
  });

  it('negative 返回 亏损', () => {
    expect(getProfitSemanticLabel('negative')).toBe('亏损');
  });

  it('zero 返回 持平', () => {
    expect(getProfitSemanticLabel('zero')).toBe('持平');
  });
});

describe('getProfitSemanticColor', () => {
  it('positive 返回绿色', () => {
    expect(getProfitSemanticColor('positive')).toBe('#52c41a');
  });

  it('negative 返回红色', () => {
    expect(getProfitSemanticColor('negative')).toBe('#ff4d4f');
  });

  it('zero 返回灰色', () => {
    expect(getProfitSemanticColor('zero')).toBe('#8c8c8c');
  });
});

describe('formatProfitValue', () => {
  it('正常数字格式化', () => {
    expect(formatProfitValue(123.456)).toBe('123.46');
  });

  it('负数格式化', () => {
    expect(formatProfitValue(-50)).toBe('-50.00');
  });

  it('NaN 返回 0.00', () => {
    expect(formatProfitValue(NaN)).toBe('0.00');
  });

  it('Infinity 返回 0.00', () => {
    expect(formatProfitValue(Infinity)).toBe('0.00');
  });
});

describe('parseProfitBySymbolChartData', () => {
  it('undefined 返回空数组', () => {
    expect(parseProfitBySymbolChartData(undefined)).toEqual([]);
  });

  it('有效 JSON 数组解析', () => {
    const data = JSON.stringify([
      { symbol: '600519', value: '100' },
      { symbol: '000858', value: '-50' },
    ]);
    const result = parseProfitBySymbolChartData(data);
    expect(result).toHaveLength(2);
    expect(result[0].symbol).toBe('600519');
    expect(result[0].value).toBe('100');
  });

  it('无效 JSON 返回空数组', () => {
    expect(parseProfitBySymbolChartData('invalid json')).toEqual([]);
  });

  it('envelope 格式解析', () => {
    const data = JSON.stringify({
      kind: 'profit_by_symbol',
      metric: 'realized_profit',
      points: [
        { symbol: '600519', value: '100' },
      ],
    });
    const result = parseProfitBySymbolChartData(data);
    expect(result).toHaveLength(1);
    expect(result[0].symbol).toBe('600519');
  });

  it('优先使用 realized_profit 作为 fallback', () => {
    const items = [
      { symbol: '600519', realized_profit: '100', total_profit: '150' },
      { symbol: '000858', realized_profit: '-20', total_profit: '10' },
    ] as unknown as AnalysisReportItemResponse[];
    const result = parseProfitBySymbolChartData(undefined, items);
    expect(result).toEqual([
      { symbol: '600519', value: '100' },
      { symbol: '000858', value: '-20' },
    ]);
  });

  it('兼容旧 total_profit fallback', () => {
    const items = [
      { symbol: '600519', realized_profit: '', total_profit: '150' },
    ] as unknown as AnalysisReportItemResponse[];
    const result = parseProfitBySymbolChartData(undefined, items);
    expect(result).toEqual([{ symbol: '600519', value: '150' }]);
  });
});

describe('buildOutcomeDistributionData', () => {
  it('undefined 返回空摘要', () => {
    const result = buildOutcomeDistributionData(undefined);
    expect(result.total).toBe(0);
    expect(result.points).toEqual([]);
  });

  it('有数据时正确计算分布', () => {
    const report = {
      symbols_count: 3,
      winning_trades: 2,
      losing_trades: 1,
      chart_data: '',
      items: [],
    };
    const result = buildOutcomeDistributionData(report);
    expect(result.total).toBe(3);
    expect(result.positiveCount).toBe(2);
    expect(result.negativeCount).toBe(1);
  });
});

describe('buildProfitCompositionData', () => {
  it('undefined 返回空摘要', () => {
    const result = buildProfitCompositionData(undefined);
    expect(result.points).toEqual([]);
    expect(result.realizedTotal).toBe(0);
  });

  it('有数据时正确计算构成', () => {
    const items = [
      { symbol: '600519', realized_profit: '100', unrealized_profit: '50', total_profit: '150' },
      { symbol: '000858', realized_profit: '-20', unrealized_profit: '30', total_profit: '10' },
    ] as unknown as AnalysisReportItemResponse[];
    const result = buildProfitCompositionData(items);
    expect(result.points).toHaveLength(2);
    expect(result.realizedTotal).toBe(80);
    expect(result.unrealizedTotal).toBe(80);
  });
});

describe('buildChangePercentRankingData', () => {
  it('undefined 返回空数组', () => {
    expect(buildChangePercentRankingData(undefined)).toEqual([]);
  });

  it('有数据时正确排序', () => {
    const items = [
      { symbol: '600519', change_percent_7d: '5.5' },
      { symbol: '000858', change_percent_7d: '-2.3' },
      { symbol: '300750', change_percent_7d: '10.1' },
    ] as unknown as AnalysisReportItemResponse[];
    const result = buildChangePercentRankingData(items);
    expect(result).toHaveLength(3);
    expect(result[0].symbol).toBe('300750'); // 最高
    expect(result[2].symbol).toBe('000858'); // 最低
  });
});
