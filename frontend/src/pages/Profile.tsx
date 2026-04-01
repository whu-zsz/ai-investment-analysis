import { useState } from 'react';
import {
  Row, Col, Card, Avatar, Typography, Button, Input, Form,
  Tag, Progress, Space, message, Statistic, Alert, Select,
} from 'antd';
import {
  UserOutlined, PhoneOutlined, EditOutlined, SaveOutlined,
  SafetyCertificateOutlined, LogoutOutlined, BarChartOutlined,
  ClockCircleOutlined, TrophyOutlined, RiseOutlined,
  ArrowLeftOutlined, BulbOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { api } from '../types';

const { Title, Text, Paragraph } = Typography;

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const activityColor: Record<string, string> = {
  analysis: '#1677ff', upload: '#52c41a',
  prediction: '#1677ff', history: '#595959', setting: '#ff4d4f',
};

const preferenceLabelMap: Record<string, string> = {
  conservative: '稳健型',
  balanced: '平衡型',
  aggressive: '进取型',
};

export default function ProfilePage() {
  const navigate = useNavigate();
  const { userInfo, logout, updateUserInfo, refreshProfile } = useAuth();
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  const handleEdit = () => {
    form.setFieldsValue({
      phone: userInfo?.phone ?? undefined,
      avatar_url: userInfo?.avatar_url ?? undefined,
      investment_preference: userInfo?.investment_preference ?? 'balanced',
    });
    setEditing(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      await api.updateProfile(values);
      updateUserInfo(values);
      await refreshProfile();
      setEditing(false);
      message.success('个人信息已更新');
    } finally {
      setSaving(false);
    }
  };

  const accountStats = [
    { label: '累计分析次数', value: 128, suffix: '次', icon: <BarChartOutlined />, color: '#1677ff' },
    { label: '账户健康分', value: 74.2, suffix: '/ 100', icon: <SafetyCertificateOutlined />, color: '#52c41a' },
    { label: '累计收益', value: Number(userInfo?.total_profit ?? 0), suffix: '元', icon: <TrophyOutlined />, color: '#52c41a' },
    { label: '风险偏好', value: preferenceLabelMap[userInfo?.investment_preference ?? 'balanced'] ?? '平衡型', suffix: '', icon: <RiseOutlined />, color: '#1677ff' },
  ];

  const activityLog = [
    { action: '完成风险诊断', time: '今日 14:30', type: 'analysis' },
    { action: '上传招商银行对账单', time: '今日 10:15', type: 'upload' },
    { action: '查看收益预演', time: '昨日 16:40', type: 'prediction' },
    { action: '导出历史流水 CSV', time: '昨日 09:22', type: 'history' },
    { action: '更新投资偏好', time: '3 天前', type: 'setting' },
  ];

  const quickLinks = [
    { label: '风险扫描', desc: 'AI 深度诊断', path: '/app/analysis', icon: <BulbOutlined /> },
    { label: '收益预演', desc: '趋势结论查看', path: '/app/prediction', icon: <ThunderboltOutlined /> },
    { label: '数据同步', desc: '导入对账单', path: '/app/upload', icon: <RiseOutlined /> },
    { label: '归档流水', desc: '历史交易记录', path: '/app/history', icon: <BarChartOutlined /> },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}
      >
        返回首页
      </Button>

      <Card
        bordered={false}
        style={{
          marginBottom: 24, borderRadius: 20,
          background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
          boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
        }}
        bodyStyle={{ padding: 28 }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
            <Avatar
              size={72}
              src={userInfo?.avatar_url ?? undefined}
              icon={<UserOutlined />}
              style={{ background: 'rgba(255,255,255,0.2)', fontSize: 30, flexShrink: 0 }}
            />
            <div>
              <Space size={10} style={{ marginBottom: 8 }}>
                <Tag color="processing">AI 驱动</Tag>
                <Tag color="blue">已认证用户</Tag>
              </Space>
              <Title level={2} style={{ margin: 0, color: '#fff' }}>
                {userInfo?.username ?? '用户'}
              </Title>
              <Paragraph style={{ margin: '6px 0 0', color: 'rgba(255,255,255,0.7)' }}>
                {preferenceLabelMap[userInfo?.investment_preference ?? 'balanced'] ?? '平衡型'} · {userInfo?.email ?? '—'}
              </Paragraph>
            </div>
          </div>
          <Button
            danger
            icon={<LogoutOutlined />}
            onClick={() => { logout(); navigate('/'); }}
            style={{
              background: 'rgba(255,77,79,0.15)',
              border: '1px solid rgba(255,77,79,0.4)',
              color: '#fff',
              borderRadius: 10,
            }}
          >
            退出登录
          </Button>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        <Col span={24} lg={7}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} style={cardStyle}>
              <Statistic
                title="账户健康度"
                value={74.2}
                suffix="/ 100"
                prefix={<SafetyCertificateOutlined />}
                valueStyle={{ color: '#52c41a', fontSize: 32 }}
              />
              <Progress
                percent={74.2}
                showInfo={false}
                strokeColor={{ '0%': '#1677ff', '100%': '#52c41a' }}
                style={{ marginTop: 12 }}
              />
              <Text type="secondary" style={{ fontSize: 12, marginTop: 6, display: 'block' }}>
                高于 78% 的同类用户
              </Text>
            </Card>

            <Card
              bordered={false}
              style={cardStyle}
              title={<span><ClockCircleOutlined style={{ color: '#1677ff', marginRight: 8 }} />最近活动</span>}
            >
              {activityLog.map((item, i) => (
                <div key={i} style={{
                  display: 'flex', alignItems: 'center', gap: 12,
                  padding: '10px 0',
                  borderBottom: i < activityLog.length - 1 ? '1px solid #f0f0f0' : 'none',
                }}>
                  <div style={{
                    width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
                    background: activityColor[item.type],
                  }} />
                  <div style={{ flex: 1 }}>
                    <Text style={{ fontSize: 13, display: 'block' }}>{item.action}</Text>
                    <Text type="secondary" style={{ fontSize: 11 }}>{item.time}</Text>
                  </div>
                </div>
              ))}
            </Card>
          </Space>
        </Col>

        <Col span={24} lg={17}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Row gutter={[12, 12]}>
              {accountStats.map(item => (
                <Col span={12} key={item.label}>
                  <Card bordered={false} style={cardStyle}>
                    <Statistic
                      title={item.label}
                      value={item.value}
                      suffix={item.suffix ? <span style={{ fontSize: 14, color: '#bfbfbf' }}>{item.suffix}</span> : undefined}
                      prefix={<span style={{ color: item.color, marginRight: 4 }}>{item.icon}</span>}
                      valueStyle={{ color: item.color, fontSize: 28, fontWeight: 700 }}
                    />
                  </Card>
                </Col>
              ))}
            </Row>

            <Card
              bordered={false}
              style={cardStyle}
              title={<span><EditOutlined style={{ color: '#1677ff', marginRight: 8 }} />个人信息</span>}
              extra={
                editing ? (
                  <Button type="primary" size="small" loading={saving} icon={<SaveOutlined />} onClick={handleSave} style={{ borderRadius: 8 }}>保存</Button>
                ) : (
                  <Button size="small" icon={<EditOutlined />} onClick={handleEdit} style={{ borderRadius: 8 }}>编辑</Button>
                )
              }
            >
              {editing ? (
                <Form form={form} layout="vertical">
                  <Form.Item name="phone" label="联系电话">
                    <Input prefix={<PhoneOutlined />} placeholder="请输入联系电话" style={{ borderRadius: 10, height: 42 }} />
                  </Form.Item>
                  <Form.Item name="avatar_url" label="头像地址">
                    <Input prefix={<UserOutlined />} placeholder="请输入头像 URL" style={{ borderRadius: 10, height: 42 }} />
                  </Form.Item>
                  <Form.Item name="investment_preference" label="投资偏好">
                    <Select
                      options={[
                        { label: '稳健型', value: 'conservative' },
                        { label: '平衡型', value: 'balanced' },
                        { label: '进取型', value: 'aggressive' },
                      ]}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Form>
              ) : (
                <div>
                  {[
                    { label: '用户名', value: userInfo?.username, icon: <UserOutlined /> },
                    { label: '邮箱地址', value: userInfo?.email, icon: <UserOutlined /> },
                    { label: '联系电话', value: userInfo?.phone, icon: <PhoneOutlined /> },
                    { label: '投资偏好', value: preferenceLabelMap[userInfo?.investment_preference ?? 'balanced'] ?? '平衡型', icon: <RiseOutlined /> },
                    { label: '累计收益', value: userInfo?.total_profit, icon: <TrophyOutlined /> },
                    { label: '风险承受', value: userInfo?.risk_tolerance, icon: <SafetyCertificateOutlined /> },
                  ].map((row, i, arr) => (
                    <div key={row.label} style={{
                      display: 'flex', alignItems: 'center', gap: 16,
                      padding: '14px 0',
                      borderBottom: i < arr.length - 1 ? '1px solid #f0f0f0' : 'none',
                    }}>
                      <div style={{
                        width: 36, height: 36, borderRadius: 10, flexShrink: 0,
                        background: '#e6f4ff', color: '#1677ff', fontSize: 15,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                      }}>
                        {row.icon}
                      </div>
                      <div>
                        <Text type="secondary" style={{ fontSize: 11, display: 'block' }}>{row.label}</Text>
                        <Text style={{ fontSize: 14, fontWeight: 500 }}>{row.value ?? '—'}</Text>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Card>

            <Card
              bordered={false}
              style={cardStyle}
              title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />快捷入口</span>}
            >
              <Row gutter={[12, 12]}>
                {quickLinks.map(link => (
                  <Col span={12} key={link.label}>
                    <Card hoverable bordered={false} onClick={() => navigate(link.path)} style={{ borderRadius: 14, background: '#f8fafc' }}>
                      <Space direction="vertical" size={6}>
                        <span style={{ color: '#1677ff', fontSize: 20 }}>{link.icon}</span>
                        <Text strong>{link.label}</Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>{link.desc}</Text>
                      </Space>
                    </Card>
                  </Col>
                ))}
              </Row>
            </Card>

            <Card bordered={false} style={cardStyle}>
              <Alert
                type="info"
                showIcon
                icon={<BulbOutlined />}
                message="AI 一句话结论：当前账户已完成真实用户资料接入，后续建议优先保持资料与投资偏好同步。"
                description="现在展示的是后端返回的真实用户资料，编辑能力也已切换到后端支持字段。"
              />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}
