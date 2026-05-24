import { test, expect } from '@playwright/test';

test.describe('数据展示', () => {
  test('市场数据加载完成', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('text=指数')).toBeVisible();
  });

  test('未登录状态显示', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('text=登录')).toBeVisible();
  });
});
