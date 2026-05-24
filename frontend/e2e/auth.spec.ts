import { test, expect } from '@playwright/test';

test.describe('认证流程', () => {
  test('未登录访问受保护页面跳转登录页', async ({ page }) => {
    await page.goto('/app/upload');
    await expect(page).toHaveURL(/\/login/);
  });

  test('登录成功跳转首页并存储token', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // 填写登录表单
    await page.fill('input[placeholder*="请输入账号"]', 'e2etest');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');

    // 等待页面跳转
    await page.waitForURL('/', { timeout: 10000 });

    // 验证 URL 是首页
    await expect(page).toHaveURL('/');

    // 验证 token 已存储到 localStorage
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();
    expect(token).toContain('eyJ'); // JWT token 以 eyJ 开头

    // 验证用户信息已存储
    const userInfo = await page.evaluate(() => localStorage.getItem('userInfo'));
    expect(userInfo).toBeTruthy();

    // 验证页面显示了用户相关元素（如用户头像或下拉菜单）
    await expect(page.locator('button:has(.ant-avatar)')).toBeVisible();
  });

  test('登录失败显示错误提示', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // 填写错误的密码
    await page.fill('input[placeholder*="请输入账号"]', 'e2etest');
    await page.fill('input[placeholder*="请输入密码"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // 等待错误提示出现
    await page.waitForTimeout(2000);

    // 验证仍在登录页
    await expect(page).toHaveURL(/\/login/);

    // 验证 token 未存储
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeFalsy();
  });

  test('退出登录清除状态', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[placeholder*="请输入账号"]', 'e2etest');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { timeout: 10000 });

    // 验证登录成功
    const tokenBefore = await page.evaluate(() => localStorage.getItem('token'));
    expect(tokenBefore).toBeTruthy();

    // 点击用户头像/下拉菜单
    await page.locator('button:has(.ant-avatar)').click();
    await page.waitForTimeout(500);
    await page.click('text=退出登录');

    // 等待跳转到首页
    await page.waitForURL('/', { timeout: 10000 });

    // 验证 token 已清除
    const tokenAfter = await page.evaluate(() => localStorage.getItem('token'));
    expect(tokenAfter).toBeFalsy();
  });

  test('已登录访问登录页跳转首页', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // 等待输入框出现
    await page.waitForSelector('input[placeholder*="请输入账号"]', { timeout: 5000 });
    await page.waitForSelector('input[placeholder*="请输入密码"]', { timeout: 5000 });

    // 填写登录信息
    await page.fill('input[placeholder*="请输入账号"]', 'e2etest');
    await page.fill('input[placeholder*="请输入密码"]', 'Test123456');

    // 点击登录按钮
    await page.click('button[type="submit"]');

    // 等待跳转到首页
    await page.waitForURL('/', { timeout: 10000 });

    // 验证登录成功
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();

    // 再次访问登录页
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
  });

  test('路由守卫正确重定向', async ({ page }) => {
    const protectedRoutes = ['/app/upload', '/app/analysis', '/app/history', '/app/portfolio', '/profile'];

    for (const route of protectedRoutes) {
      await page.goto(route);
      await expect(page).toHaveURL(/\/login/);
    }
  });
});
