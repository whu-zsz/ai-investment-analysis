import { useLocation, useNavigate } from 'react-router-dom';

const pageMeta: Record<string, { title: string; subtitle: string }> = {
  '/': { title: '投资驾驶舱', subtitle: '整合投资总览、风险雷达和近期趋势的首页。' },
  '/upload': { title: '数据导入', subtitle: '支持 CSV / Excel 上传、识别与后续清洗。' },
  '/analysis': { title: 'AI 分析报告', subtitle: '展示投资偏好、行为模式与风险提示。' },
  '/prediction': { title: '趋势预测', subtitle: '以图形化卡片展示未来收益区间与策略建议。' },
  '/history': { title: '历史归档', subtitle: '查看交易明细、上传记录与状态标签。' },
};

const Navbar = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const username = localStorage.getItem('username') ?? '演示用户';
  const meta = pageMeta[location.pathname] ?? pageMeta['/'];

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    navigate('/login');
  };

  return (
    <header className="topbar">
      <div>
        <p className="eyebrow">AI Investment Assistant</p>
        <h1>{meta.title}</h1>
        <p className="muted">{meta.subtitle}</p>
      </div>

      <div className="topbar-right">
        <span className="pill">预留 DeepSeek / OpenAI API 接口</span>
        <div className="profile">
          <div className="avatar">{username.slice(0, 1)}</div>
          <div>
            <strong>{username}</strong>
            <span>普通投资者画像</span>
          </div>
          <button className="button button-ghost" onClick={handleLogout}>
            退出
          </button>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
