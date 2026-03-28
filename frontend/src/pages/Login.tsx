import { Card, Form, Input, Button, Typography, message, ConfigProvider, theme, Space, Tag } from 'antd';
import { UserOutlined, LockOutlined, ArrowRightOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

export default function Login() {
  const navigate = useNavigate();

  const onFinish = (values: { username: string; password: string }) => {
    console.log('Login:', values);
    localStorage.setItem('token', 'your_token_here');
    message.success('验证成功，正在进入系统...');
    navigate('/', { replace: true });
  };

  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        width: '100vw',
        background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)',
        position: 'relative',
        overflow: 'hidden',
      }}>

        {/* 背景光晕 —— 和 Dashboard Hero 渐变同色 */}
        <div style={{
          position: 'absolute',
          top: '-20%', left: '-10%',
          width: '55vw', height: '55vw',
          background: 'radial-gradient(circle, rgba(22,119,255,0.12) 0%, transparent 70%)',
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute',
          bottom: '-15%', right: '-10%',
          width: '40vw', height: '40vw',
          background: 'radial-gradient(circle, rgba(22,119,255,0.07) 0%, transparent 70%)',
          pointerEvents: 'none',
        }} />

        {/* 登录卡片 */}
        <Card
          bordered={false}
          style={{
            width: 440,
            borderRadius: 24,
            background: 'rgba(15, 23, 42, 0.75)',
            backdropFilter: 'blur(24px)',
            border: '1px solid rgba(22,119,255,0.2)',
            boxShadow: '0 32px 64px rgba(0,0,0,0.55), 0 0 0 1px rgba(255,255,255,0.04) inset',
            padding: '8px',
          }}
        >
          {/* 顶部品牌区 */}
          <div style={{ textAlign: 'center', marginBottom: 36 }}>
            {/* Logo 图标 */}
            <div style={{
              width: 56, height: 56,
              borderRadius: 16,
              background: 'linear-gradient(135deg, #1677ff 0%, #69b1ff 100%)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              margin: '0 auto 20px',
              boxShadow: '0 8px 24px rgba(22,119,255,0.4)',
            }}>
              <SafetyCertificateOutlined style={{ fontSize: 26, color: '#fff' }} />
            </div>

            <Space size={8} style={{ marginBottom: 14 }}>
              <Tag color="processing" style={{ borderRadius: 20, padding: '2px 12px' }}>AI 驱动</Tag>
              <Tag color="blue" style={{ borderRadius: 20, padding: '2px 12px' }}>实时市场洞察</Tag>
            </Space>

            <Title level={2} style={{ color: '#fff', marginBottom: 8, marginTop: 0 }}>AI投顾助手</Title>
            <Text style={{ color: 'rgba(255,255,255,0.4)', fontSize: 14 }}>
              请输入凭证以访问 AI 分析系统
            </Text>
          </div>

          {/* 表单 */}
          <Form onFinish={onFinish} size="large" layout="vertical">
            <Form.Item
              name="username"
              rules={[{ required: true, message: '请输入账号' }]}
              style={{ marginBottom: 16 }}
            >
              <Input
                prefix={<UserOutlined style={{ color: 'rgba(255,255,255,0.3)' }} />}
                placeholder="管理账号"
                style={{
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: 12,
                  height: 48,
                  color: '#fff',
                }}
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
              style={{ marginBottom: 32 }}
            >
              <Input.Password
                prefix={<LockOutlined style={{ color: 'rgba(255,255,255,0.3)' }} />}
                placeholder="访问密码"
                style={{
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: 12,
                  height: 48,
                  color: '#fff',
                }}
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                type="primary"
                htmlType="submit"
                block
                icon={<ArrowRightOutlined />}
                style={{
                  height: 52,
                  borderRadius: 14,
                  fontWeight: 600,
                  fontSize: 16,
                  background: 'linear-gradient(135deg, #1677ff 0%, #4096ff 100%)',
                  border: 'none',
                  boxShadow: '0 8px 24px rgba(22,119,255,0.4)',
                  letterSpacing: '0.05em',
                }}
              >
                开启智能分析
              </Button>
            </Form.Item>
          </Form>

          {/* 底部安全标注 */}
          <div style={{
            marginTop: 28,
            paddingTop: 20,
            borderTop: '1px solid rgba(255,255,255,0.07)',
            textAlign: 'center',
          }}>
            <Text style={{ color: 'rgba(255,255,255,0.25)', fontSize: 12, letterSpacing: '0.04em' }}>
              DEEPSEEK-V3 安全链路 · 端到端加密
            </Text>
          </div>
        </Card>
      </div>
    </ConfigProvider>
  );
}