import { test, expect } from '@playwright/test';

test.describe('页面导航', () => {
  test('首页加载完成', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('text=市场')).toBeVisible();
  });

  test('登录页加载完成', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('text=登录')).toBeVisible();
  });

  test('404 重定向', async ({ page }) => {
    await page.goto('/nonexistent-page');
    await expect(page).toHaveURL('/');
  });
});
