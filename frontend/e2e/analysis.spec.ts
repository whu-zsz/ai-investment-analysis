import { test, expect } from '@playwright/test';

test.describe('AI 分析', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('input[placeholder*="请输入账号"]', 'testuser');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
  });

  test('分析页面加载完成', async ({ page }) => {
    await page.goto('/app/analysis');
    await expect(page.locator('text=分析')).toBeVisible();
  });

  test('历史页面加载完成', async ({ page }) => {
    await page.goto('/app/history');
    await expect(page.locator('text=历史')).toBeVisible();
  });
});
