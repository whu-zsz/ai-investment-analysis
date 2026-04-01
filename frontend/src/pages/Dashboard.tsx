import { Alert, Button, Card, Col, Descriptions, Divider, Avatar, Dropdown, Popover, Progress, Row, Space, Statistic, Tag, Typography } from 'antd';
import {
  BarChartOutlined,
  BulbOutlined,
  InfoCircleOutlined,
  LineChartOutlined,
  RadarChartOutlined,
  RiseOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import { useAuth } from '../hooks/useAuth';
import type { MenuProps } from 'antd';

const { Paragraph, Text, Title } = Typography;

interface ChartParam {
  axisValueLabel?: string;
  name: string;
  value: number;
}

interface KpiChartConfig {
  seriesName: string;
  unit: string;
  labels: string[];
  values: number[];
  color: string;
}

interface KpiCard {
  title: string;
  value: number;
  precision: number;
  suffix: string;
  accent: string;
  tagColor: string;
  tagText: string;
  desc: string;
  chart: KpiChartConfig;
}

const kpiCards: KpiCard[] = [
  {
    title: '总资产估值',
    value: 128.6,
    precision: 1,
    suffix: '万',
    accent: '#1677ff',
    tagColor: 'blue',
    tagText: '较上周 +3.2%',
    desc: '权益仓位维持高位，资金利用率良好。',
    chart: {
      seriesName: '总资产估值',
      unit: '万',
      labels: ['03-19', '03-20', '03-21', '03-24', '03-25', '03-26', '03-27'],
      values: [121.2, 122.4, 123.8, 125.6, 126.9, 127.8, 128.6],
      color: '#1677ff'
    }
  },
  {
    title: '今日盈亏',
    value: 1.84,
    precision: 2,
    suffix: '万',
    accent: '#52c41a',
    tagColor: 'green',
    tagText: '跑赢沪深300 +0.68%',
    desc: '科技与红利双主线贡献主要涨幅。',
    chart: {
      seriesName: '日内盈亏',
      unit: '万',
      labels: ['09:35', '10:00', '10:30', '11:00', '13:30', '14:30', '15:00'],
      values: [0.22, 0.45, 0.36, 0.72, 1.15, 1.46, 1.84],
      color: '#52c41a'
    }
  },
  {
    title: '累计收益率',
    value: 24.7,
    precision: 1,
    suffix: '%',
    accent: '#722ed1',
    tagColor: 'purple',
    tagText: '年内新高附近',
    desc: '净值曲线仍保持上升通道，回撤可控。',
    chart: {
      seriesName: '累计收益率',
      unit: '%',
      labels: ['10月', '11月', '12月', '1月', '2月', '3月', '本周'],
      values: [12.8, 14.9, 16.7, 18.4, 20.9, 22.6, 24.7],
      color: '#722ed1'
    }
  },
  {
    title: '风险健康度',
    value: 74.2,
    precision: 1,
    suffix: '/ 100',
    accent: '#fa8c16',
    tagColor: 'orange',
    tagText: '需关注集中度',
    desc: '组合进攻性较强，建议关注仓位平衡。',
    chart: {
      seriesName: '风险健康度',
      unit: '分',
      labels: ['周一', '周二', '周三', '周四', '周五', '本周', '当前'],
      values: [79.8, 78.6, 77.1, 76.5, 75.4, 74.9, 74.2],
      color: '#fa8c16'
    }
  }
];

const quickStats = [
  { label: '区间涨跌幅', value: '+6.82%', color: '#52c41a' },
  { label: '跑赢基准', value: '+1.36%', color: '#1677ff' },
  { label: '最大回撤', value: '4.90%', color: '#fa8c16' },
  { label: '年化波动率', value: '18.4%', color: '#722ed1' },
  { label: '月度换手率', value: '32%', color: '#13c2c2' },
  { label: '近30日胜率', value: '61%', color: '#eb2f96' }
];

function getKpiChartOption(chart: KpiChartConfig, mode: 'mini' | 'expanded'): EChartsOption {
  const isMini = mode === 'mini';

  return {
    animation: true,
    tooltip: isMini
      ? { show: false }
      : {
          trigger: 'axis',
          backgroundColor: 'rgba(255, 255, 255, 0.96)',
          borderColor: '#d9e6ff',
          borderWidth: 1,
          formatter: (params: unknown) => {
            const list = params as ChartParam[];
            const data = list[0];
            return `<div style="padding: 4px 6px;">
                      <div style="color: #888; margin-bottom: 4px;">${data.axisValueLabel ?? data.name}</div>
                      <div style="font-weight: bold; color: ${chart.color}; font-size: 16px;">${data.value.toLocaleString()} ${chart.unit}</div>
                    </div>`;
          }
        },
    grid: isMini
      ? { top: 6, left: 0, right: 0, bottom: 0, containLabel: false }
      : { top: 20, left: 36, right: 16, bottom: 28, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: chart.labels,
      show: !isMini,
      axisLine: isMini ? { show: false } : { lineStyle: { color: '#d9d9d9' } },
      axisTick: { show: false },
      axisLabel: { color: '#8c8c8c', fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      show: !isMini,
      axisLabel: { color: '#8c8c8c', fontSize: 11 },
      splitLine: isMini ? { show: false } : { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: chart.seriesName,
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: chart.values,
        lineStyle: { width: isMini ? 2 : 3, color: chart.color },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: `${chart.color}${isMini ? '30' : '40'}` },
              { offset: 1, color: `${chart.color}08` }
            ]
          }
        }
      }
    ]
  };
}

export default function Dashboard() {
  const navigate = useNavigate();
  const location = useLocation();
  const isPublicHome = location.pathname === '/';
  const { isLoggedIn, userInfo, logout } = useAuth();

  const guardNavigate = (path: string) => {
    if (!isLoggedIn) navigate('/login', { state: { from: path } });
    else navigate(path);
  };

  const userMenuItems: MenuProps['items'] = [
    { key: 'profile', icon: <SettingOutlined />, label: '个人中心' },
    { type: 'divider' },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
  ];

  const handleUserMenu: MenuProps['onClick'] = ({ key }) => {
    if (key === 'profile') navigate('/profile');
    if (key === 'logout') { logout(); navigate('/'); }
  };

  const getOption = (): EChartsOption => ({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: '#d9e6ff',
      borderWidth: 1,
      formatter: (params: unknown) => {
        const list = params as ChartParam[];
        const data = list[0];
        return `<div style="padding: 4px 6px;">
                  <div style="color: #888; margin-bottom: 4px;">${data.name} 指数</div>
                  <div style="font-weight: bold; color: #1677ff; font-size: 16px;">${data.value.toLocaleString()} 点</div>
                </div>`;
      }
    },
    grid: {
      top: '10%',
      left: '3%',
      right: '4%',
      bottom: '8%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: ['03-17', '03-18', '03-19', '03-20', '03-21', '03-24', '03-25'],
      axisLine: { lineStyle: { color: '#d9d9d9' } }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(0,0,0,0.08)' } }
    },
    series: [
      {
        name: '上证指数',
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: [3058, 3072, 3064, 3096, 3108, 3116, 3128.42],
        lineStyle: { width: 3, color: '#1677ff' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22, 119, 255, 0.35)' },
              { offset: 1, color: 'rgba(22, 119, 255, 0.02)' }
            ]
          }
        }
      }
    ]
  });

  return (
    <div style={{ padding: isPublicHome ? '24px' : '4px' }}>
      {isPublicHome && (
        <Card
          bordered={false}
          style={{
            marginBottom: 24,
            borderRadius: 20,
            background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
            boxShadow: '0 18px 40px rgba(22,119,255,0.18)'
          }}
          bodyStyle={{ padding: 28 }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
            <div>
              <Space size={12} wrap style={{ marginBottom: 12 }}>
                <Tag color="processing">AI 驱动</Tag>
                <Tag color="gold">实时市场洞察</Tag>
              </Space>
              <Title level={2} style={{ margin: 0, color: '#fff' }}>观势智投</Title>
              <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 720 }}>
                聚合指数趋势、仓位风险、交易效率与 AI 建议，快速看到今日市场温度与组合状态。
              </Paragraph>
            </div>
            <Space wrap>
              <Button ghost onClick={() => guardNavigate('/app/upload')}>上传记录</Button>
              <Button ghost onClick={() => guardNavigate('/app/analysis')}>AI 风险分析</Button>
              <Button ghost onClick={() => guardNavigate('/app/prediction')}>趋势预测</Button>
              <Button ghost onClick={() => guardNavigate('/app/history')}>历史归档</Button>
              {isLoggedIn ? (
                <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenu }} placement="bottomRight" arrow>
                  <Button
                    type="primary"
                    icon={<Avatar size={18} icon={<UserOutlined />} style={{ background: 'rgba(255,255,255,0.25)', verticalAlign: 'middle' }} />}
                    style={{ display: 'flex', alignItems: 'center', gap: 6 }}
                  >
                    {userInfo?.username ?? '用户'}
                  </Button>
                </Dropdown>
              ) : (
                <Button type="primary" icon={<UserOutlined />} onClick={() => navigate('/login')}>
                  登录
                </Button>
              )}
            </Space>
          </div>
        </Card>
      )}

      <Row gutter={[16, 16]}>
        {kpiCards.map((item) => (
          <Col xs={24} sm={12} lg={6} key={item.title}>
            <Card
              bordered={false}
              hoverable
              onClick={() => navigate('/app/analysis')}
              style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15, 23, 42, 0.06)' }}
            >
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 16 }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Text type="secondary">{item.title}</Text>
                  <Statistic
                    value={item.value}
                    precision={item.precision}
                    suffix={item.suffix}
                    valueStyle={{ color: item.accent, fontSize: 30 }}
                    style={{ marginTop: 8 }}
                  />
                </div>
                <div style={{ width: 124, flexShrink: 0, display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 10 }}>
                  <Tag color={item.tagColor} style={{ marginInlineEnd: 0, textAlign: 'center' }}>{item.tagText}</Tag>
                  <Popover
                    trigger="hover"
                    placement="bottom"
                    overlayStyle={{ maxWidth: 360 }}
                    content={
                      <div style={{ width: 320 }}>
                        <Text strong>{item.title}走势明细</Text>
                        <ReactECharts option={getKpiChartOption(item.chart, 'expanded')} style={{ height: 220, marginTop: 8 }} />
                      </div>
                    }
                  >
                    <div style={{ cursor: 'zoom-in', borderRadius: 12, background: '#fafcff', border: '1px solid #eef2f6', padding: '4px 6px' }}>
                      <ReactECharts option={getKpiChartOption(item.chart, 'mini')} style={{ height: 62 }} />
                    </div>
                  </Popover>
                </div>
              </div>
              <Paragraph type="secondary" style={{ margin: '14px 0 0', minHeight: 44 }}>
                {item.desc}
              </Paragraph>
            </Card>
          </Col>
        ))}
      </Row>

      <Card
        title={
          <span>
            <LineChartOutlined style={{ marginRight: 8, color: '#1677ff' }} />
            今日大盘走势
          </span>
        }
        extra={
          <Space size={8} wrap>
            <Tag color="processing" icon={<RiseOutlined />}>上证指数近 7 日</Tag>
            <Text type="secondary">更新时间 15:00</Text>
          </Space>
        }
        bordered={false}
        style={{ marginTop: 24, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.06)' }}
      >
        <ReactECharts option={getOption()} style={{ height: '400px' }} />

        <Row gutter={[12, 12]} style={{ marginTop: 8 }}>
          {quickStats.map((item) => (
            <Col xs={12} md={8} lg={4} key={item.label}>
              <div style={{ background: '#f8fafc', borderRadius: 12, padding: '14px 16px', border: '1px solid #eef2f6' }}>
                <Text type="secondary" style={{ fontSize: 12 }}>{item.label}</Text>
                <div style={{ marginTop: 6, fontSize: 20, fontWeight: 700, color: item.color }}>{item.value}</div>
              </div>
            </Col>
          ))}
        </Row>
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 8 }}>
        <Col span={24} lg={12}>
          <Card
            bordered={false}
            title={<span><RadarChartOutlined style={{ color: '#1677ff', marginRight: 8 }} />风险与仓位摘要</span>}
            style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Card bordered={false} style={{ background: '#f8fafc' }}>
                  <Statistic
                    title="风险健康分"
                    value={74.2}
                    suffix="/ 100"
                    prefix={<SafetyCertificateOutlined />}
                    valueStyle={{ color: '#52c41a' }}
                  />
                  <Progress percent={74.2} showInfo={false} strokeColor="#52c41a" style={{ marginTop: 8 }} />
                </Card>
              </Col>
              <Col span={12}>
                <Card bordered={false} style={{ background: '#fff7e6' }}>
                  <Statistic
                    title="仓位使用率"
                    value={81}
                    suffix="%"
                    prefix={<BarChartOutlined />}
                    valueStyle={{ color: '#fa8c16' }}
                  />
                  <Progress percent={81} showInfo={false} strokeColor="#fa8c16" style={{ marginTop: 8 }} />
                </Card>
              </Col>
            </Row>

            <Divider><InfoCircleOutlined /> 关键参数</Divider>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="持仓集中度">
                前 3 大持仓占比 <Text strong>58%</Text>，单一行业暴露偏高。
              </Descriptions.Item>
              <Descriptions.Item label="组合 Beta">
                当前 Beta 为 <Text strong>1.18</Text>，高于稳健配置区间。
              </Descriptions.Item>
              <Descriptions.Item label="现金仓位">
                可用现金 <Text strong>19%</Text>，具备一定防御和补仓弹性。
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col span={24} lg={12}>
          <Card
            bordered={false}
            title={<span><ThunderboltOutlined style={{ color: '#722ed1', marginRight: 8 }} />交易与行为摘要</span>}
            style={{ borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
          >
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="近 30 日交易次数">
                共交易 <Text strong>26</Text> 次，节奏略高于当前市场波动所需。
              </Descriptions.Item>
              <Descriptions.Item label="月度胜率与盈亏比">
                胜率 <Text strong>61%</Text>，盈亏比 <Text strong>1.47</Text>，策略仍有优化空间。
              </Descriptions.Item>
              <Descriptions.Item label="平均持仓天数">
                平均持仓 <Text strong>11</Text> 天，短线切换偏多。
              </Descriptions.Item>
              <Descriptions.Item label="手续费侵蚀估计">
                本月摩擦成本约 <Text type="danger">0.8%</Text>，建议降低无效换手。
              </Descriptions.Item>
            </Descriptions>

            <div style={{ background: '#f8fafc', padding: 18, borderRadius: 12, marginTop: 16 }}>
              <Title level={5} style={{ marginTop: 0 }}>行为特征提示</Title>
              <Paragraph style={{ marginBottom: 0 }}>
                近期止盈执行较为积极，但亏损头寸处理偏慢，存在轻微“盈利先跑、亏损后拖”倾向。若市场进入震荡期，建议提高纪律性阈值。
              </Paragraph>
            </div>
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        style={{ marginTop: 16, borderRadius: 16, boxShadow: '0 8px 24px rgba(15,23,42,0.05)' }}
      >
        <Alert
          type="info"
          showIcon
          icon={<BulbOutlined />}
          message="AI 一句话结论：当前组合仍处于偏进攻状态，收益动能尚可，但需尽快压低集中度与高频换手。"
          description={
            <Space direction="vertical" size={6}>
              <Text>主要风险点：科技权重偏高、Beta 略高、交易摩擦成本持续累积。</Text>
              <Text>建议动作：将高波动板块仓位下调 10%-15%，增加红利低波或现金缓冲，并进一步查看 AI 风险分析页获取调仓细项。</Text>
            </Space>
          }
        />
      </Card>
    </div>
  );
}