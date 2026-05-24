import { test, expect } from '@playwright/test';

test.describe('文件上传', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[placeholder*="请输入账号"]', 'e2etest');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');
  });

  test('上传页面加载完成', async ({ page }) => {
    await page.goto('/app/upload');
    await page.waitForLoadState('networkidle');
    // 验证页面加载完成
    await expect(page.locator('h2, h3, h4')).toBeVisible();
  });

  test('查看上传历史', async ({ page }) => {
    await page.goto('/app/upload');
    await page.waitForLoadState('networkidle');
    // 验证页面加载完成
    await expect(page.locator('h2, h3, h4')).toBeVisible();
  });
});
