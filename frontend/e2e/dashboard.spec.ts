import { test, expect } from '@playwright/test';

test.describe('数据展示', () => {
  test('市场数据加载完成', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    // 验证页面加载完成
    await expect(page.locator('h2')).toBeVisible();
  });

  test('未登录状态显示', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    // 验证页面加载完成
    await expect(page.locator('h2')).toBeVisible();
  });
});
