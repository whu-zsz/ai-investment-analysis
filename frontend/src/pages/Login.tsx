import { Card, Form, Input, Button, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

export default function Login() {
  const navigate = useNavigate();

  const onFinish = (values: any) => {
    console.log('登录数据:', values);
    // 模拟登录成功，存入 token
    localStorage.setItem('token', 'mock_token_shangzhezhe');
    message.success('登录成功！');
    navigate('/app');
  };

  return (
    <div style={{ 
      display: 'flex', justifyContent: 'center', alignItems: 'center', 
      height: '100vh', background: '#f0f2f5' 
    }}>
      <Card style={{ width: 400, borderRadius: '15px', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <div style={{ textAlign: 'center', marginBottom: 30 }}>
          <Typography.Title level={2}>AI 投顾助手</Typography.Title>
          <Typography.Text type="secondary">请登录您的账户</Typography.Text>
        </div>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名 (张盛哲)" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>立即登录</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}