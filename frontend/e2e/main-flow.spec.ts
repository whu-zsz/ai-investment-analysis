import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const TEST_USER = 'e2etest';
const TEST_PASS = 'Test123456';
const AUTH_FILE = path.join(__dirname, '.auth.json');
const CSV_FILE = path.join(__dirname, 'fixtures', 'test.csv');

test.describe.serial('主要用户流程', () => {

  // 第一步：登录
  test('01-登录系统', async ({ page }) => {
    await page.goto('/login');
    await page.waitForSelector('input[placeholder="请输入账号"]', { timeout: 10000 });

    await page.fill('input[placeholder="请输入账号"]', TEST_USER);
    await page.fill('input[placeholder="请输入密码"]', TEST_PASS);
    await page.click('button[type="submit"]');

    await page.waitForURL('/', { timeout: 15000 });
    await page.waitForTimeout(2000);

    await expect(page).toHaveURL('/');
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();

    await page.context().storageState({ path: AUTH_FILE });
  });

  // 后续测试共享登录状态
  test.describe(() => {
    test.use({ storageState: AUTH_FILE });

    // 第二步：仪表盘
    test('02-仪表盘页面', async ({ page }) => {
      await page.goto('/');
      await page.waitForTimeout(10000);

      await expect(page).not.toHaveURL(/\/login/);
      const count = await page.locator('h2').count();
      expect(count).toBeGreaterThan(0);
    });

    // 第三步：上传 CSV 文件
    test('03-上传交易文件', async ({ page }) => {
      await page.goto('/app/upload');
      await page.waitForTimeout(5000);

      await expect(page).not.toHaveURL(/\/login/);

      // 找到文件 input（在 Ant Design Dragger 组件中）
      // 使用 setInputFiles 选择文件
      const fileChooserPromise = page.waitForEvent('filechooser');
      // 点击 Dragger 区域触发文件选择
      await page.locator('.ant-upload-drag').click();
      const fileChooser = await fileChooserPromise;
      await fileChooser.setFiles(CSV_FILE);

      // 等待本地预览出现（表格显示前 5 行数据）
      await page.waitForTimeout(3000);

      // 点击"确认提交"按钮上传
      const submitBtn = page.getByText('确认提交');
      await submitBtn.waitFor({ timeout: 10000 });
      await submitBtn.click();

      // 等待上传完成（进度条到 100%）
      await page.waitForTimeout(8000);

      // 截图记录上传结果
      await page.screenshot({ path: 'test-results/upload-result.png' });
    });

    // 第四步：历史交易页面
    test('04-历史交易页面', async ({ page }) => {
      await page.goto('/app/history');
      await page.waitForTimeout(10000);

      await expect(page).not.toHaveURL(/\/login/);
      const count = await page.locator('h2, h3, h4').count();
      expect(count).toBeGreaterThan(0);

      // 截图记录
      await page.screenshot({ path: 'test-results/history.png' });
    });

    // 第五步：AI 分析页面 - 生成报告（需要较长时间，设置 2 分钟超时）
    test('05-AI分析页面', async ({ page }) => {
      test.setTimeout(120000);
      await page.goto('/app/analysis');
      await page.waitForTimeout(15000);

      await expect(page).not.toHaveURL(/\/login/);
      const count = await page.locator('h2, h3, h4').count();
      expect(count).toBeGreaterThan(0);

      // 截图记录初始状态
      await page.screenshot({ path: 'test-results/analysis-before.png' });

      // 查找"重新生成"按钮
      const btn = page.getByRole('button', { name: /重新生成|重新分析|生成/i });
      if (await btn.isVisible({ timeout: 5000 })) {
        await btn.click();

        // AI 分析需要等待较长时间
        await page.waitForTimeout(30000);

        // 截图记录分析结果
        await page.screenshot({ path: 'test-results/analysis-result.png' });
      }
    });

    // 第六步：退出登录
    test('06-退出登录', async ({ page }) => {
      await page.goto('/');
      await page.waitForTimeout(5000);

      const avatarBtn = page.locator('button:has(.ant-avatar)');
      await avatarBtn.waitFor({ timeout: 10000 });
      await avatarBtn.click();
      await page.waitForTimeout(500);

      await page.getByText('退出登录').click();
      await page.waitForTimeout(3000);

      const token = await page.evaluate(() => localStorage.getItem('token'));
      expect(token).toBeFalsy();

      // 清理
      try { fs.unlinkSync(AUTH_FILE); } catch {}
    });
  });
});
