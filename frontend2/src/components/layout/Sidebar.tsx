import { NavLink } from 'react-router-dom';

const navItems = [
  { to: '/', label: '总览', index: '01' },
  { to: '/upload', label: '上传记录', index: '02' },
  { to: '/analysis', label: 'AI 报告', index: '03' },
  { to: '/prediction', label: '趋势预测', index: '04' },
  { to: '/history', label: '历史归档', index: '05' },
];

const Sidebar = () => {
  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="brand-logo">AI</div>
        <div>
          <h2>AI 投顾助手</h2>
          <p>课程设计前端 demo</p>
        </div>
      </div>

      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
          >
            <span>{item.label}</span>
            <small>{item.index}</small>
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-panel">
        <p className="eyebrow">Project Scope</p>
        <h3>面向普通投资者的分析流程原型</h3>
        <p>覆盖登录、上传、总结、预测、记录归档，后续很适合继续接后端接口与真实图表。</p>
      </div>
    </aside>
  );
};

export default Sidebar;
