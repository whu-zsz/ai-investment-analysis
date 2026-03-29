import { useState } from 'react';
import {
  Row, Col, Card, Avatar, Typography, Button, Input, Form,
  Divider, Tag, Progress, Space, message, Statistic, Alert,
} from 'antd';
import {
  UserOutlined, MailOutlined, EditOutlined, SaveOutlined,
  SafetyCertificateOutlined, LogoutOutlined, BarChartOutlined,
  ClockCircleOutlined, TrophyOutlined, RiseOutlined,
  ArrowLeftOutlined, BulbOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const { Title, Text, Paragraph } = Typography;

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

const activityColor: Record<string, string> = {
  analysis: '#1677ff', upload: '#52c41a',
  prediction: '#1677ff', history: '#595959', setting: '#ff4d4f',
};

export default function ProfilePage() {
  const navigate = useNavigate();
  const { userInfo, logout, updateUserInfo } = useAuth();
  const [editing, setEditing] = useState(false);
  const [form] = Form.useForm();

  const handleEdit = () => {
    form.setFieldsValue({
      displayName: userInfo?.displayName,
      email: userInfo?.email,
      role: userInfo?.role,
    });
    setEditing(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    updateUserInfo(values);
    setEditing(false);
    message.success('个人信息已更新');
  };

  const accountStats = [
    { label: '累计分析次数', value: 128,  suffix: '次',    icon: <BarChartOutlined />,         color: '#1677ff', bg: '#e6f4ff' },
    { label: '账户健康分',   value: 74.2, suffix: '/ 100', icon: <SafetyCertificateOutlined />, color: '#52c41a', bg: '#f6ffed' },
    { label: '使用天数',     value: 88,   suffix: '天',    icon: <ClockCircleOutlined />,       color: '#1677ff', bg: '#e6f4ff' },
    { label: '累计收益率',   value: 24.7, suffix: '%',     icon: <TrophyOutlined />,            color: '#52c41a', bg: '#f6ffed' },
  ];

  const activityLog = [
    { action: '完成风险诊断',       time: '今日 14:30', type: 'analysis' },
    { action: '上传招商银行对账单', time: '今日 10:15', type: 'upload' },
    { action: '查看收益预演',       time: '昨日 16:40', type: 'prediction' },
    { action: '导出历史流水 CSV',   time: '昨日 09:22', type: 'history' },
    { action: '更新风险因子权重',   time: '3 天前',     type: 'setting' },
  ];

  const quickLinks = [
    { label: '风险扫描',  desc: 'AI 深度诊断',       path: '/app/analysis',   icon: <BulbOutlined /> },
    { label: '收益预演',  desc: 'Monte Carlo 模拟',  path: '/app/prediction', icon: <ThunderboltOutlined /> },
    { label: '数据同步',  desc: '导入对账单',        path: '/app/upload',     icon: <RiseOutlined /> },
    { label: '归档流水',  desc: '历史交易记录',      path: '/app/history',    icon: <BarChartOutlined /> },
  ];

  return (
    <div style={{ padding: '24px' }}>

      {/* 返回按钮 */}
      <Button
        icon={<ArrowLeftOutlined />}
        type="text"
        onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}
      >
        返回首页
      </Button>

      {/* Hero Banner */}
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
              icon={<UserOutlined />}
              style={{ background: 'rgba(255,255,255,0.2)', fontSize: 30, flexShrink: 0 }}
            />
            <div>
              <Space size={10} style={{ marginBottom: 8 }}>
                <Tag color="processing">AI 驱动</Tag>
                <Tag color="blue">高级会员</Tag>
              </Space>
              <Title level={2} style={{ margin: 0, color: '#fff' }}>
                {userInfo?.displayName ?? '用户'}
              </Title>
              <Paragraph style={{ margin: '6px 0 0', color: 'rgba(255,255,255,0.7)' }}>
                {userInfo?.role ?? '分析师'} · {userInfo?.email ?? '—'}
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
        {/* ── 左栏 ── */}
        <Col span={24} lg={7}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>

            {/* 健康度 */}
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

            {/* 最近活动 */}
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

        {/* ── 右栏 ── */}
        <Col span={24} lg={17}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>

            {/* 使用统计 */}
            <Row gutter={[12, 12]}>
              {accountStats.map(item => (
                <Col span={12} key={item.label}>
                  <Card bordered={false} style={cardStyle}>
                    <Statistic
                      title={item.label}
                      value={item.value}
                      suffix={<span style={{ fontSize: 14, color: '#bfbfbf' }}>{item.suffix}</span>}
                      prefix={<span style={{ color: item.color, marginRight: 4 }}>{item.icon}</span>}
                      valueStyle={{ color: item.color, fontSize: 28, fontWeight: 700 }}
                    />
                  </Card>
                </Col>
              ))}
            </Row>

            {/* 编辑个人信息 */}
            <Card
              bordered={false}
              style={cardStyle}
              title={<span><EditOutlined style={{ color: '#1677ff', marginRight: 8 }} />个人信息</span>}
              extra={
                editing ? (
                  <Button type="primary" size="small" icon={<SaveOutlined />} onClick={handleSave} style={{ borderRadius: 8 }}>保存</Button>
                ) : (
                  <Button size="small" icon={<EditOutlined />} onClick={handleEdit} style={{ borderRadius: 8 }}>编辑</Button>
                )
              }
            >
              {editing ? (
                <Form form={form} layout="vertical">
                  {[
                    { name: 'displayName', label: '显示名称', prefix: <UserOutlined />, placeholder: '请输入显示名称' },
                    { name: 'email',       label: '邮箱地址', prefix: <MailOutlined />, placeholder: '请输入邮箱地址' },
                    { name: 'role',        label: '职位角色', prefix: <RiseOutlined />, placeholder: '请输入职位角色' },
                  ].map(field => (
                    <Form.Item key={field.name} name={field.name} label={field.label}>
                      <Input
                        prefix={field.prefix}
                        placeholder={field.placeholder}
                        style={{ borderRadius: 10, height: 42 }}
                      />
                    </Form.Item>
                  ))}
                </Form>
              ) : (
                <div>
                  {[
                    { label: '显示名称', value: userInfo?.displayName, icon: <UserOutlined /> },
                    { label: '邮箱地址', value: userInfo?.email,       icon: <MailOutlined /> },
                    { label: '职位角色', value: userInfo?.role,        icon: <RiseOutlined /> },
                    { label: '注册时间', value: userInfo?.joinDate,    icon: <ClockCircleOutlined /> },
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

            {/* 快捷入口 */}
            <Card
              bordered={false}
              style={cardStyle}
              title={<span><RiseOutlined style={{ color: '#1677ff', marginRight: 8 }} />快捷功能入口</span>}
            >
              <Row gutter={[12, 12]}>
                {quickLinks.map(item => (
                  <Col span={12} key={item.label}>
                    <div
                      onClick={() => navigate(item.path)}
                      style={{
                        background: '#f8fafc',
                        border: '1px solid #eef2f6',
                        borderRadius: 12,
                        padding: '16px 18px',
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                        display: 'flex', alignItems: 'center', gap: 12,
                      }}
                      onMouseEnter={e => {
                        const el = e.currentTarget as HTMLDivElement;
                        el.style.background = '#e6f4ff';
                        el.style.borderColor = '#91caff';
                      }}
                      onMouseLeave={e => {
                        const el = e.currentTarget as HTMLDivElement;
                        el.style.background = '#f8fafc';
                        el.style.borderColor = '#eef2f6';
                      }}
                    >
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

            {/* AI 提示 */}
            <Card bordered={false} style={cardStyle}>
              <Alert
                type="info"
                showIcon
                icon={<BulbOutlined />}
                message="AI 账户洞察：您的使用频率高于 85% 的用户，分析质量持续提升。"
                description={
                  <Space direction="vertical" size={4}>
                    <Text type="secondary">近 30 天完成 12 次风险诊断，组合健康度从 68 提升至 74.2。</Text>
                    <Text type="secondary">建议每周至少上传一次对账单以保持数据新鲜度。</Text>
                  </Space>
                }
              />
            </Card>

          </Space>
        </Col>
      </Row>
    </div>
  );
}