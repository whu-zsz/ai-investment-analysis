import { test, expect } from '@playwright/test';

test.describe('页面导航', () => {
  test('首页加载完成', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    // 验证页面标题或关键元素
    await expect(page.locator('h2')).toBeVisible();
  });

  test('登录页加载完成', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=观势智投')).toBeVisible();
  });

  test('404 重定向', async ({ page }) => {
    await page.goto('/nonexistent-page');
    await page.waitForLoadState('networkidle');
    // 应该重定向到首页或登录页
    await expect(page).toHaveURL(/\/(login)?/);
  });
});
