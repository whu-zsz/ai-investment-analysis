import type { TagProps } from 'antd';
import type { MarketDataStatus } from '../api/types';

export const marketStatusMap: Record<MarketDataStatus, { color: TagProps['color']; text: string }> = {
  complete: { color: 'success', text: '市场数据完整' },
  fetched_live: { color: 'processing', text: '市场数据实时拉取' },
  partial: { color: 'warning', text: '市场数据部分缺失' },
  unavailable: { color: 'error', text: '市场数据不可用' },
};

export function getMarketStatusMeta(status?: string) {
  return status && status in marketStatusMap
    ? marketStatusMap[status as MarketDataStatus]
    : { color: 'default' as TagProps['color'], text: status || '未知状态' };
}
