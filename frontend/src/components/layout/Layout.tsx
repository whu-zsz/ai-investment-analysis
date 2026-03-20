import { Outlet } from 'react-router-dom'
import Navbar from './Navbar'
import Sidebar from './Sidebar'

const Layout = () => {
  return (
    <div className="flex h-screen bg-gray-50">
      {/* 左侧 Sidebar */}
      <Sidebar />

      {/* 右侧内容区 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <Navbar />
        
        {/* 主内容 */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />   {/* 这里会自动渲染 Dashboard / Upload 等页面 */}
        </main>
      </div>
    </div>
  )
}

export default Layout