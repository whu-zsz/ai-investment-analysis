import { useState, useEffect } from 'react';
import {
  Row, Col, Card, Avatar, Typography, Button, Input, Form,
  Divider, Tag, Space, message, Statistic, Spin, Select, Empty,
} from 'antd';
import {
  UserOutlined, MailOutlined, EditOutlined, SaveOutlined,
  SafetyCertificateOutlined, LogoutOutlined, BarChartOutlined,
  TrophyOutlined, RiseOutlined,
  ArrowLeftOutlined, PhoneOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { userApi } from '../api/index';
import type { UserResponse } from '../api/types';

const { Title, Text, Paragraph } = Typography;
const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const preferenceMap: Record<string, string> = {
  conservative: '保守型', balanced: '稳健型', aggressive: '激进型',
};

export default function ProfilePage() {
  const navigate = useNavigate();
  const { logout, updateUserInfo } = useAuth();
  const [user, setUser] = useState<UserResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => { void fetchProfile(); }, []);

  const fetchProfile = async () => {
    setLoading(true);
    try {
      const res = await userApi.getProfile();
      setUser(res);
      updateUserInfo(res);
    } catch (err: any) {
      setUser(null);
      const msg = err?.message ?? err?.data?.message ?? '个人资料加载失败';
      message.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    form.setFieldsValue({
      phone: user?.phone ?? '',
      investment_preference: user?.investment_preference ?? 'balanced',
    });
    setEditing(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      const res = await userApi.updateProfile(values);
      setUser(res);
      updateUserInfo(res);
      message.success('个人信息已更新');
      setEditing(false);
    } catch (err: any) {
      const msg = err?.message ?? err?.data?.message ?? '保存失败';
      message.error(msg);
    } finally {
      setSaving(false);
    }
  };

  const quickLinks = [
    { label: '风险扫描', desc: '查看分析报告', path: '/app/analysis', icon: <BarChartOutlined /> },
    { label: '收益预演', desc: '查看趋势结论', path: '/app/prediction', icon: <RiseOutlined /> },
    { label: '数据同步', desc: '导入交易记录', path: '/app/upload', icon: <RiseOutlined /> },
    { label: '归档流水', desc: '查看交易明细', path: '/app/history', icon: <BarChartOutlined /> },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
        返回首页
      </Button>

      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
            <Avatar size={72} src={user?.avatar_url} icon={<UserOutlined />}
              style={{ background: 'rgba(255,255,255,0.2)', fontSize: 30, flexShrink: 0 }} />
            <div>
              <Space size={10} style={{ marginBottom: 8 }}>
                <Tag color="processing">个人资料</Tag>
                <Tag color="blue">{user ? preferenceMap[user.investment_preference] : '—'}</Tag>
              </Space>
              <Title level={2} style={{ margin: 0, color: '#fff' }}>{user?.username ?? '用户'}</Title>
              <Paragraph style={{ margin: '6px 0 0', color: 'rgba(255,255,255,0.7)' }}>
                {user?.email ?? '—'}
              </Paragraph>
            </div>
          </div>
          <Button danger icon={<LogoutOutlined />}
            onClick={() => { logout(); navigate('/'); }}
            style={{ background: 'rgba(255,77,79,0.15)', border: '1px solid rgba(255,77,79,0.4)', color: '#fff', borderRadius: 10 }}>
            退出登录
          </Button>
        </div>
      </Card>

      <Spin spinning={loading}>
        {!user ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description="当前无法获取个人资料" />
          </Card>
        ) : (
          <Row gutter={[16, 16]}>
            <Col span={24} lg={7}>
              <Space direction="vertical" style={{ width: '100%' }} size={16}>
                <Card bordered={false} style={cardStyle}>
                  <Statistic title="累计盈亏" value={parseFloat(user.total_profit).toLocaleString()}
                    prefix="¥" valueStyle={{ color: '#52c41a', fontSize: 32 }} />
                  <Divider style={{ margin: '16px 0' }} />
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                    <Text type="secondary" style={{ fontSize: 13 }}>风险承受能力</Text>
                    <Tag color="processing">{user.risk_tolerance}</Tag>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Text type="secondary" style={{ fontSize: 13 }}>投资偏好</Text>
                    <Tag color="blue">{preferenceMap[user.investment_preference]}</Tag>
                  </div>
                </Card>
              </Space>
            </Col>

            <Col span={24} lg={17}>
              <Space direction="vertical" style={{ width: '100%' }} size={16}>
                <Row gutter={[12, 12]}>
                  {[
                    { label: '用户 ID', value: user.id, suffix: '', icon: <UserOutlined />, color: '#1677ff' },
                    { label: '累计盈亏', value: `¥${parseFloat(user.total_profit).toLocaleString()}`, suffix: '', icon: <TrophyOutlined />, color: '#52c41a' },
                    { label: '风险承受', value: user.risk_tolerance, suffix: '', icon: <SafetyCertificateOutlined />, color: '#1677ff' },
                    { label: '投资偏好', value: preferenceMap[user.investment_preference], suffix: '', icon: <BarChartOutlined />, color: '#52c41a' },
                  ].map(item => (
                    <Col span={12} key={item.label}>
                      <Card bordered={false} style={cardStyle}>
                        <Statistic title={item.label} value={item.value}
                          suffix={<span style={{ fontSize: 14, color: '#bfbfbf' }}>{item.suffix}</span>}
                          prefix={<span style={{ color: item.color, marginRight: 4 }}>{item.icon}</span>}
                          valueStyle={{ color: item.color, fontSize: 24, fontWeight: 700 }} />
                      </Card>
                    </Col>
                  ))}
                </Row>

                <Card bordered={false} style={cardStyle}
                  title={<span><EditOutlined style={{ color: '#1677ff', marginRight: 8 }} />个人信息</span>}
                  extra={
                    editing
                      ? <Button type="primary" size="small" icon={<SaveOutlined />} loading={saving} onClick={handleSave} style={{ borderRadius: 8 }}>保存</Button>
                      : <Button size="small" icon={<EditOutlined />} onClick={handleEdit} style={{ borderRadius: 8 }}>编辑</Button>
                  }>
                  {editing ? (
                    <Form form={form} layout="vertical">
                      <Form.Item name="phone" label="手机号码">
                        <Input prefix={<PhoneOutlined />} placeholder="请输入手机号码" style={{ borderRadius: 10, height: 42 }} />
                      </Form.Item>
                      <Form.Item name="investment_preference" label="投资偏好">
                        <Select style={{ height: 42 }} options={[
                          { value: 'conservative', label: '保守型' },
                          { value: 'balanced', label: '稳健型' },
                          { value: 'aggressive', label: '激进型' },
                        ]} />
                      </Form.Item>
                    </Form>
                  ) : (
                    <div>
                      {[
                        { label: '用户名', value: user.username, icon: <UserOutlined /> },
                        { label: '邮箱', value: user.email, icon: <MailOutlined /> },
                        { label: '手机号', value: user.phone ?? '未设置', icon: <PhoneOutlined /> },
                        { label: '投资偏好', value: preferenceMap[user.investment_preference], icon: <RiseOutlined /> },
                      ].map((row, i, arr) => (
                        <div key={row.label} style={{ display: 'flex', alignItems: 'center', gap: 16, padding: '14px 0', borderBottom: i < arr.length - 1 ? '1px solid #f0f0f0' : 'none' }}>
                          <div style={{ width: 36, height: 36, borderRadius: 10, flexShrink: 0, background: '#e6f4ff', color: '#1677ff', fontSize: 15, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            {row.icon}
                          </div>
                          <div>
                            <Text type="secondary" style={{ fontSize: 11, display: 'block' }}>{row.label}</Text>
                            <Text style={{ fontSize: 14, fontWeight: 500 }}>{row.value}</Text>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </Card>

                <Card bordered={false} style={cardStyle}
                  title={<span><RiseOutlined style={{ color: '#1677ff', marginRight: 8 }} />快捷功能入口</span>}>
                  <Row gutter={[12, 12]}>
                    {quickLinks.map(item => (
                      <Col span={12} key={item.label}>
                        <div onClick={() => navigate(item.path)} style={{ background: '#f8fafc', border: '1px solid #eef2f6', borderRadius: 12, padding: '16px 18px', cursor: 'pointer', transition: 'all 0.2s', display: 'flex', alignItems: 'center', gap: 12 }}
                          onMouseEnter={e => { const el = e.currentTarget as HTMLDivElement; el.style.background = '#e6f4ff'; el.style.borderColor = '#91caff'; }}
                          onMouseLeave={e => { const el = e.currentTarget as HTMLDivElement; el.style.background = '#f8fafc'; el.style.borderColor = '#eef2f6'; }}>
                          <div style={{ color: '#1677ff', fontSize: 18 }}>{item.icon}</div>
                          <div>
                            <Text strong style={{ fontSize: 14, display: 'block' }}>{item.label}</Text>
                            <Text type="secondary" style={{ fontSize: 12 }}>{item.desc}</Text>
                          </div>
                        </div>
                      </Col>
                    ))}
                  </Row>
                </Card>
              </Space>
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
}
