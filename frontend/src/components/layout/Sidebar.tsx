import { NavLink } from 'react-router-dom'

const menuItems = [
  { path: '/', label: '首页', icon: '🏠' },
  { path: '/upload', label: '上传记录', icon: '📤' },
  { path: '/analysis', label: 'AI分析报告', icon: '📊' },
  { path: '/prediction', label: '趋势预测', icon: '🔮' },
  { path: '/history', label: '历史记录', icon: '📜' },
]

const Sidebar = () => {
  return (
    <div className="w-64 bg-white border-r p-6 flex flex-col">
      <div className="text-xl font-bold text-blue-600 mb-8">菜单</div>
      
      <nav className="flex flex-col gap-2">
        {menuItems.map(item => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `flex items-center gap-3 px-4 py-3 rounded-xl text-lg transition-all ${
                isActive 
                  ? 'bg-blue-50 text-blue-600 font-medium' 
                  : 'hover:bg-gray-100 text-gray-700'
              }`
            }
          >
            <span className="text-2xl">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>
    </div>
  )
}

export default Sidebar