import React from 'react';
import { Layout, Menu, Avatar, Space } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { 
  DashboardOutlined, 
  CloudUploadOutlined, 
  BarChartOutlined, 
  HistoryOutlined, 
  BulbOutlined,
  UserOutlined,
  LogoutOutlined
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
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={240} theme="dark" collapsible>
        <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: 18, fontWeight: 'bold' }}>
          AI 投顾助手
        </div>
        <Menu 
          theme="dark" 
          mode="inline" 
          selectedKeys={[location.pathname]} 
          onClick={({ key }) => navigate(key)}
          items={[
            { key: '/', icon: <DashboardOutlined />, label: '投资驾驶舱' },
            { key: '/upload', icon: <CloudUploadOutlined />, label: '上传记录' },
            { key: '/analysis', icon: <BulbOutlined />, label: 'AI 风险分析' },
            { key: '/prediction', icon: <BarChartOutlined />, label: '趋势预测' },
            { key: '/history', icon: <HistoryOutlined />, label: '历史归档' },
          ]}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #f0f0f0' }}>
          <span style={{ fontSize: 16, fontWeight: 500 }}>
            {location.pathname === '/' ? '数据总览' : '功能模块'}
          </span>
          <Space size={16}>
            <span style={{ fontSize: 12, color: '#999' }}>DeepSeek API V3 已连接</span>
            <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
            <LogoutOutlined onClick={handleLogout} style={{ cursor: 'pointer', color: '#ff4d4f' }} title="退出登录" />
          </Space>
        </Header>
        <Content style={{ margin: '24px', minHeight: 280 }}>
          <Outlet /> {/* 核心：这里只负责渲染 pages 里的内容 */}
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;