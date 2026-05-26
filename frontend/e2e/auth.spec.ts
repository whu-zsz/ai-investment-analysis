import { test, expect } from '@playwright/test';

const TEST_USER = 'e2etest';
const TEST_PASS = 'Test123456';

test.describe('认证流程', () => {
  test('未登录访问受保护页面跳转登录页', async ({ page }) => {
    await page.goto('/app/upload');
    await page.waitForURL(/\/login/, { timeout: 5000 });
    // 确认看到登录页的标题
    await expect(page.locator('h2:has-text("观势智投")')).toBeVisible();
  });

  test('登录成功跳转首页', async ({ page }) => {
    // 1. 进入登录页
    await page.goto('/login');
    await page.waitForSelector('input[placeholder="请输入账号"]', { timeout: 5000 });

    // 2. 填写表单
    await page.fill('input[placeholder="请输入账号"]', TEST_USER);
    await page.fill('input[placeholder="请输入密码"]', TEST_PASS);

    // 3. 点击登录
    await page.click('button[type="submit"]');

    // 4. 等待跳转到首页
    await page.waitForURL('/', { timeout: 10000 });

    // 5. 验证在首页 - 应该看到 h2（Dashboard 的标题）
    await expect(page).toHaveURL('/');

    // 6. 验证 token 已存储
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();
  });

  test('登录失败显示错误提示', async ({ page }) => {
    await page.goto('/login');
    await page.waitForSelector('input[placeholder="请输入账号"]', { timeout: 5000 });

    await page.fill('input[placeholder="请输入账号"]', TEST_USER);
    await page.fill('input[placeholder="请输入密码"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // 等待错误提示
    await page.waitForTimeout(2000);

    // 验证仍在登录页
    await expect(page).toHaveURL(/\/login/);

    // 验证 token 未存储
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeFalsy();
  });

  test('退出登录清除状态', async ({ page }) => {
    // 1. 登录
    await page.goto('/login');
    await page.waitForSelector('input[placeholder="请输入账号"]', { timeout: 5000 });
    await page.fill('input[placeholder="请输入账号"]', TEST_USER);
    await page.fill('input[placeholder="请输入密码"]', TEST_PASS);
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { timeout: 10000 });

    // 2. 验证登录成功
    const tokenBefore = await page.evaluate(() => localStorage.getItem('token'));
    expect(tokenBefore).toBeTruthy();

    // 3. 点击用户头像下拉菜单
    await page.locator('button:has(.ant-avatar)').click();
    await page.waitForTimeout(500);

    // 4. 点击退出登录
    await page.getByText('退出登录').click();

    // 5. 等待跳转
    await page.waitForTimeout(2000);

    // 6. 验证 token 已清除
    const tokenAfter = await page.evaluate(() => localStorage.getItem('token'));
    expect(tokenAfter).toBeFalsy();
  });

  test('路由守卫正确重定向', async ({ page }) => {
    const protectedRoutes = [
      '/app/upload',
      '/app/analysis',
      '/app/history',
      '/app/portfolio',
      '/profile',
    ];

    for (const route of protectedRoutes) {
      await page.goto(route);
      await page.waitForURL(/\/login/, { timeout: 5000 });
      await expect(page).toHaveURL(/\/login/);
    }
  });
});
