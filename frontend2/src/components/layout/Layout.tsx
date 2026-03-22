import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';
import Sidebar from './Sidebar';

const Layout = () => {
  return (
    <div className="shell">
      <Sidebar />
      <div className="content-shell">
        <Navbar />
        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;
