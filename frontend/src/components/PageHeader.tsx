import { Avatar, Button, Space, Typography } from 'antd';
import { ArrowLeftOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Text } = Typography;

interface PageHeaderProps {
  title: string;
  subtitle?: string;
}

export default function PageHeader({ title, subtitle }: PageHeaderProps) {
  const navigate = useNavigate();

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '0 32px',
      height: 64,
      background: 'rgba(15, 23, 42, 0.6)',
      backdropFilter: 'blur(12px)',
      borderBottom: '1px solid rgba(255,255,255,0.08)',
      position: 'sticky',
      top: 0,
      zIndex: 100,
    }}>
      <Space size={16} align="center">
        <Button
          type="text"
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/')}
          style={{
            color: 'rgba(255,255,255,0.55)',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
          }}
        >
          返回
        </Button>
        <div style={{ width: 1, height: 20, background: 'rgba(255,255,255,0.12)' }} />
        <div>
          <div style={{ color: '#fff', fontWeight: 600, fontSize: 15, lineHeight: 1.3 }}>{title}</div>
          {subtitle && <div style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>{subtitle}</div>}
        </div>
      </Space>

      <Space size={16}>
        <Text style={{ color: 'rgba(255,255,255,0.3)', fontSize: 11, letterSpacing: '0.05em' }}>
          DEEPSEEK-V3 安全链路
        </Text>
        <Avatar icon={<UserOutlined />} size={32} style={{ backgroundColor: '#1677ff' }} />
      </Space>
    </div>
  );
}