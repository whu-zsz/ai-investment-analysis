import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useLocation, useNavigate } from 'react-router-dom';

interface LoginForm {
  username: string;
  password: string;
}

const Login = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    defaultValues: {
      username: '盛哲',
      password: '123456',
    },
  });

  const onSubmit = (values: LoginForm) => {
    setLoading(true);
    setError('');

    window.setTimeout(() => {
      if (values.username.trim().length < 2 || values.password.trim().length < 6) {
        setError('请输入有效账号。当前 demo 默认账号为 盛哲 / 123456。');
        setLoading(false);
        return;
      }

      localStorage.setItem('token', 'frontend2-demo-token');
      localStorage.setItem('username', values.username.trim());
      const target = (location.state as { from?: string } | null)?.from ?? '/';
      navigate(target);
    }, 650);
  };

  return (
    <div className="login-shell">
      <div className="login-grid">
        <section className="login-hero">
          <p className="eyebrow">Wuhan University · CS Project</p>
          <h1>基于 AI 大模型的投资记录分析与预测 Web 应用</h1>
          <p className="hero-text">
            为普通投资者提供上传投资记录、生成 AI 分析报告、查看风险预警与未来趋势预测的完整前端体验。
          </p>

          <div className="hero-panel">
            <strong>当前 demo 覆盖内容</strong>
            <ul>
              <li>登录与路由保护</li>
              <li>投资总览首页</li>
              <li>上传和字段识别界面</li>
              <li>AI 报告、趋势预测、历史记录页面</li>
            </ul>
          </div>
        </section>

        <section className="login-card">
          <p className="eyebrow">Sign In</p>
          <h2>欢迎进入 AI 投顾助手</h2>
          <p className="muted">先登录，再体验完整的投资分析流程。</p>

          <form className="form-grid" onSubmit={handleSubmit(onSubmit)}>
            <label className="field">
              <span>用户名 / 学号</span>
              <input
                {...register('username', { required: '请输入用户名' })}
                placeholder="例如：盛哲 / 2023302111118"
              />
              {errors.username ? <small>{errors.username.message}</small> : null}
            </label>

            <label className="field">
              <span>密码</span>
              <input
                type="password"
                {...register('password', { required: '请输入密码' })}
                placeholder="至少 6 位"
              />
              {errors.password ? <small>{errors.password.message}</small> : null}
            </label>

            <div className="login-tip">
              <span>默认演示账号：盛哲 / 123456</span>
              <span>后续可替换为 JWT 登录接口</span>
            </div>

            {error ? <div className="error-box">{error}</div> : null}

            <button className="button button-primary" type="submit" disabled={loading}>
              {loading ? '登录中...' : '进入系统'}
            </button>
          </form>
        </section>
      </div>
    </div>
  );
};

export default Login;
