import { Card, Typography, Timeline, Tag, Row, Col } from 'antd';
import { WarningOutlined, CheckCircleOutlined } from '@ant-design/icons';

export default function Analysis() {
  return (
    <div style={{ padding: '24px' }}>
      <Typography.Title level={2}>AI 风险评估报告</Typography.Title>
      <Row gutter={24}>
        <Col span={16}>
          <Card title="行为模式识别" bordered={false}>
            <Timeline items={[
              { color: 'red', children: '高频交易警报：近期短线操作过于频繁，增加了交易成本。' },
              { color: 'green', children: '持仓分布优化：资产已从单一股票扩展至指数基金，风险分散度提升。' },
              { children: '投资偏好：系统识别您的风格为“进取型”，偏好科技板块。' },
            ]} />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="当前风险等级" textAlign="center">
            <div style={{ fontSize: '32px', color: '#fa8c16', fontWeight: 'bold' }}>中等风险</div>
            <Tag color="orange" style={{ marginTop: 16 }}>建议增加防御性资产占比</Tag>
          </Card>
        </Col>
      </Row>
    </div>
  );
}