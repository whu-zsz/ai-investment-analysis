import React from 'react';
import { Card, Row, Col, Statistic, Tag, Typography } from 'antd';
import { 
  ArrowUpOutlined, 
  LineChartOutlined, 
  ThunderboltOutlined 
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';

const { Text } = Typography;

// 定义一个简单的接口来替代 any，或者直接在函数中使用类型断言
interface ChartParam {
  name: string;
  value: number;
  data: number;
}

export default function Dashboard() {
  
  const getOption = (): EChartsOption => {
    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255, 255, 255, 0.9)',
        // 修复点：将 params: any 修改为明确的类型处理
        formatter: (params: unknown) => {
          const list = params as ChartParam[];
          const data = list[0];
          return `<div style="padding: 4px;">
                    <div style="color: #888;">${data.name} 净值</div>
                    <div style="font-weight: bold; color: #1677ff;">¥${data.value.toLocaleString()}</div>
                  </div>`;
        }
      },
      grid: {
        top: '10%',
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['03-17', '03-18', '03-19', '03-20', '03-21', '03-22', '03-23'],
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { type: 'dashed' } }
      },
      series: [
        {
          name: '总资产',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: [415000, 418000, 416500, 422000, 425000, 423000, 428560],
          lineStyle: { width: 3, color: '#1677ff' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(22, 119, 255, 0.3)' },
                { offset: 1, color: 'rgba(22, 119, 255, 0)' }
              ]
            }
          }
        }
      ]
    };
  };

  return (
    <div style={{ padding: '4px' }}>
      <Row gutter={[16, 16]}>
        {/* 组合总资产 */}
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic 
              title="组合总资产" 
              value={428560} 
              precision={2} 
              prefix={<span style={{ marginRight: 4 }}>¥</span>} 
            />
            <div style={{ marginTop: 8 }}>
              <Text type="secondary">较上周 </Text>
              <Text type="success"><ArrowUpOutlined /> 3.2%</Text>
            </div>
          </Card>
        </Col>

        {/* 今日盈亏 */}
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic 
              title="今日盈亏" 
              value={1240.5} 
              precision={2} 
              prefix={<span style={{ color: '#52c41a' }}>+¥</span>}
              valueStyle={{ color: '#52c41a' }}
            />
            <Tag color="green" icon={<ThunderboltOutlined />} style={{ marginTop: 8 }}>运行稳健</Tag>
          </Card>
        </Col>

        {/* 月度胜率 */}
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic 
              title="月度胜率" 
              value={68.5} 
              suffix={<span style={{ fontSize: 14, marginLeft: 4 }}>%</span>}
              valueStyle={{ color: '#1677ff' }}
            />
            <div style={{ marginTop: 8 }}>
              <Tag color="blue">超越 85% 用户</Tag>
            </div>
          </Card>
        </Col>

        {/* 风险等级 */}
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} hoverable>
            <Statistic 
              title="风险等级" 
              value="中偏稳健" 
              valueStyle={{ color: '#fa8c16' }} 
            />
            <div style={{ marginTop: 8 }}>
              <Tag color="orange">建议平衡仓位</Tag>
            </div>
          </Card>
        </Col>
      </Row>

      <Card 
        title={
          <span>
            <LineChartOutlined style={{ marginRight: 8, color: '#1677ff' }} />
            收益趋势分析
          </span>
        } 
        bordered={false} 
        style={{ marginTop: 24, borderRadius: '12px', boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}
      >
        <ReactECharts option={getOption()} style={{ height: '400px' }} />
      </Card>
    </div>
  );
}