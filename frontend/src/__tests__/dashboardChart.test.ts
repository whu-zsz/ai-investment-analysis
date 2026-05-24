import { describe, expect, it } from 'vitest';
import { dashboardRangeQueryMap, formatDashboardBarTime } from '../pages/Dashboard';

describe('dashboardRangeQueryMap', () => {
  it('3日使用小时线参数', () => {
    expect(dashboardRangeQueryMap['3d']).toEqual({ period: '60m', adjust: 'none', limit: 12 });
  });

  it('7日使用小时线参数', () => {
    expect(dashboardRangeQueryMap['7d']).toEqual({ period: '60m', adjust: 'none', limit: 28 });
  });

  it('30日保持日线参数', () => {
    expect(dashboardRangeQueryMap['30d']).toEqual({ period: 'day', adjust: 'none', limit: 30 });
  });
});

describe('formatDashboardBarTime', () => {
  it('60m 展示到小时分钟', () => {
    expect(formatDashboardBarTime('2026-05-24 14:30:00', '60m')).toBe('05-24 14:30');
  });

  it('day 仅展示月日', () => {
    expect(formatDashboardBarTime('2026-05-24 00:00:00', 'day')).toBe('05-24');
  });

  it('保留异常短字符串', () => {
    expect(formatDashboardBarTime('2026', 'day')).toBe('2026');
  });
});
