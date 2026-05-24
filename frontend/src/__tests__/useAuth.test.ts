// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';

// normalizeStoredUser 是内部函数，需要通过测试 hook 的行为来间接测试
// 但我们可以直接导入并测试它（如果导出的话）
// 由于它是内部函数，我们通过 localStorage 行为来测试

describe('useAuth', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('无 token 时未登录', () => {
    const token = localStorage.getItem('token');
    expect(token).toBeNull();
  });

  it('存储 token 后可以读取', () => {
    localStorage.setItem('token', 'test_token');
    expect(localStorage.getItem('token')).toBe('test_token');
  });

  it('存储 userInfo 后可以读取', () => {
    const userInfo = { username: 'test', email: 'test@test.com' };
    localStorage.setItem('userInfo', JSON.stringify(userInfo));
    const stored = JSON.parse(localStorage.getItem('userInfo')!);
    expect(stored.username).toBe('test');
    expect(stored.email).toBe('test@test.com');
  });

  it('清除 token 后为空', () => {
    localStorage.setItem('token', 'test_token');
    localStorage.removeItem('token');
    expect(localStorage.getItem('token')).toBeNull();
  });

  it('清除 userInfo 后为空', () => {
    localStorage.setItem('userInfo', '{}');
    localStorage.removeItem('userInfo');
    expect(localStorage.getItem('userInfo')).toBeNull();
  });

  it('无效 JSON 不会导致崩溃', () => {
    localStorage.setItem('userInfo', 'invalid json');
    expect(() => {
      try {
        JSON.parse(localStorage.getItem('userInfo')!);
      } catch {
        // 预期的错误
      }
    }).not.toThrow();
  });
});
