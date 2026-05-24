import { test, expect } from '@playwright/test';

test.describe('认证流程', () => {
  test('未登录访问受保护页面跳转登录页', async ({ page }) => {
    await page.goto('/app/upload');
    await expect(page).toHaveURL(/\/login/);
  });

  test('登录成功跳转首页', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="请输入账号"]', 'testuser');
    await page.fill('input[placeholder="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
  });

  test('登录失败显示错误提示', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="请输入账号"]', 'testuser');
    await page.fill('input[placeholder="请输入密码"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    await expect(page.locator('.ant-message-error')).toBeVisible();
  });

  test('退出登录清除状态', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.fill('input[placeholder="请输入账号"]', 'testuser');
    await page.fill('input[placeholder="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');

    // 点击退出
    await page.click('.user-dropdown');
    await page.click('text=退出登录');
    await expect(page).toHaveURL('/');
  });

  test('已登录访问登录页跳转首页', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.fill('input[placeholder="请输入账号"]', 'testuser');
    await page.fill('input[placeholder="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');

    // 再次访问登录页
    await page.goto('/login');
    await expect(page).toHaveURL('/');
  });

  test('路由守卫正确重定向', async ({ page }) => {
    const protectedRoutes = ['/app/upload', '/app/analysis', '/app/history', '/app/portfolio', '/profile'];

    for (const route of protectedRoutes) {
      await page.goto(route);
      await expect(page).toHaveURL(/\/login/);
    }
  });
});
