import { useNavigate } from 'react-router-dom';

const Navbar = () => {
  const navigate = useNavigate();
  const username = localStorage.getItem('username') || '张盛哲';

  const handleLogout = () => {
    if (confirm('确定要退出登录吗？')) {
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      navigate('/login');
    }
  };

  return (
    <nav className="h-16 bg-white border-b flex items-center px-6 justify-between shadow-sm">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-blue-600 rounded-xl flex items-center justify-center text-white font-bold text-xl">
          AI
        </div>
        <h1 className="text-2xl font-bold text-gray-800">AI投顾助手</h1>
      </div>

      <div className="flex items-center gap-6">
        <span className="text-sm text-gray-600">欢迎，{username}</span>
        <button
          onClick={handleLogout}
          className="px-5 py-2 text-sm border border-red-300 text-red-600 rounded-2xl hover:bg-red-50 transition-all"
        >
          退出登录
        </button>
      </div>
    </nav>
  )
}

export default Navbar