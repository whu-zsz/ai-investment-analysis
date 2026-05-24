import { test, expect } from '@playwright/test';

test.describe('文件上传', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('input[placeholder*="请输入账号"]', 'testuser');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
  });

  test('上传页面加载完成', async ({ page }) => {
    await page.goto('/app/upload');
    await expect(page.locator('text=上传')).toBeVisible();
  });

  test('查看上传历史', async ({ page }) => {
    await page.goto('/app/upload');
    await expect(page.locator('text=历史')).toBeVisible();
  });
});
