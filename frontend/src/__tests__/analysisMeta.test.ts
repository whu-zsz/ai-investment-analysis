import { describe, it, expect } from 'vitest';
import { getMarketStatusMeta, marketStatusMap } from '../utils/analysisMeta';

describe('getMarketStatusMeta', () => {
  it('complete 状态返回正确配置', () => {
    const result = getMarketStatusMeta('complete');
    expect(result).toEqual(marketStatusMap.complete);
    expect(result.color).toBe('success');
    expect(result.text).toBe('市场数据完整');
  });

  it('fetched_live 状态返回正确配置', () => {
    const result = getMarketStatusMeta('fetched_live');
    expect(result).toEqual(marketStatusMap.fetched_live);
    expect(result.color).toBe('success');
    expect(result.text).toBe('市场数据已实时补全');
  });

  it('partial 状态返回正确配置', () => {
    const result = getMarketStatusMeta('partial');
    expect(result).toEqual(marketStatusMap.partial);
    expect(result.color).toBe('warning');
    expect(result.text).toBe('市场数据部分缺失');
  });

  it('unavailable 状态返回正确配置', () => {
    const result = getMarketStatusMeta('unavailable');
    expect(result).toEqual(marketStatusMap.unavailable);
    expect(result.color).toBe('error');
    expect(result.text).toBe('市场数据不可用');
  });

  it('undefined 返回默认配置', () => {
    const result = getMarketStatusMeta(undefined);
    expect(result.color).toBe('default');
    expect(result.text).toBe('未知状态');
  });

  it('未知状态返回默认配置', () => {
    const result = getMarketStatusMeta('unknown_status');
    expect(result.color).toBe('default');
    expect(result.text).toBe('unknown_status');
  });
});
