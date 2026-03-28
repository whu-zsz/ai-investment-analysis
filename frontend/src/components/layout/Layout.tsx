import React from 'react';
import { Layout, Menu, Avatar, Space, ConfigProvider, theme, Button } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { 
  DashboardOutlined, CloudUploadOutlined, BarChartOutlined, 
  HistoryOutlined, BulbOutlined, UserOutlined, LogoutOutlined 
} from '@ant-design/icons';

const { Header, Content, Sider } = Layout;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    // 使用 ConfigProvider 强制注入暗色主题变量
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <Layout style={{ minHeight: '100vh', background: '#0b1120' }}>
        <Sider 
          width={260} 
          style={{ 
            background: 'rgba(15, 23, 42, 0.8)', 
            borderRight: '1px solid rgba(255, 255, 255, 0.1)',
            backdropFilter: 'blur(10px)'
          }}
        >
          <div style={{ 
            height: 80, display: 'flex', alignItems: 'center', justifyContent: 'center', 
            color: '#1677ff', fontSize: 22, fontWeight: 'bold', letterSpacing: '2px' 
          }}>
            AI 投顾助手
          </div>
          <Menu 
            theme="dark" 
            mode="inline" 
            selectedKeys={[location.pathname]} 
            onClick={({ key }) => navigate(key)}
            style={{ background: 'transparent', border: 'none' }}
            items={[
              { key: '/app', icon: <DashboardOutlined />, label: '工作控制台' },
              { key: '/app/upload', icon: <CloudUploadOutlined />, label: '数据同步' },
              { key: '/app/analysis', icon: <BulbOutlined />, label: '风险扫描' },
              { key: '/app/prediction', icon: <BarChartOutlined />, label: '收益预演' },
              { key: '/app/history', icon: <HistoryOutlined />, label: '归档流水' },
            ]}
          />
        </Sider>

        <Layout style={{ background: 'transparent' }}>
          <Header style={{ 
            background: 'rgba(15, 23, 42, 0.4)', 
            backdropFilter: 'blur(10px)', 
            padding: '0 24px', 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            borderBottom: '1px solid rgba(255, 255, 255, 0.1)'
          }}>
            <span style={{ color: '#fff', fontSize: 16, fontWeight: 500 }}>
              {location.pathname.split('/').pop()?.toUpperCase() || 'DASHBOARD'}
            </span>
            <Space size={20}>
              <span style={{ color: 'rgba(255, 255, 255, 0.45)', fontSize: 12 }}>DEEPSEEK-V3 安全链路</span>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
              <Button 
                type="text" 
                icon={<LogoutOutlined />} 
                onClick={handleLogout} 
                style={{ color: '#ff4d4f' }}
              />
            </Space>
          </Header>

          <Content style={{ 
            padding: '24px', 
            background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)',
            overflowY: 'auto' 
          }}>
            {/* 子页面内容注入处 */}
            <div style={{ maxWidth: 1400, margin: '0 auto' }}>
              <Outlet />
            </div>
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};

export default MainLayout;