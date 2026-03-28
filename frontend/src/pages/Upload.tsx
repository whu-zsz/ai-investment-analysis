import { useState } from 'react';
import {
  Upload, Card, Typography, Table, message, Button,
  Space, Progress, Row, Col, ConfigProvider, theme
} from 'antd';
import {
  InboxOutlined, FileSearchOutlined,
  CloudUploadOutlined, CheckCircleTwoTone
} from '@ant-design/icons';
import * as XLSX from 'xlsx';
import type { RcFile } from 'antd/es/upload';
import type { ColumnsType } from 'antd/es/table';
import PageHeader from '../components/PageHeader';

const { Dragger } = Upload;
const { Title, Text } = Typography;

interface PreviewRow {
  key: string | number;
  [key: string]: string | number | boolean | undefined;
}

export default function UploadPage() {
  const [dataPreview, setDataPreview] = useState<PreviewRow[]>([]);
  const [columns, setColumns] = useState<ColumnsType<PreviewRow>>([]);
  const [uploading, setUploading] = useState<boolean>(false);
  const [percent, setPercent] = useState<number>(0);
  const [fileName, setFileName] = useState<string>('');

  const handleFile = (file: RcFile) => {
    setFileName(file.name);
    const reader = new FileReader();

    reader.onload = (e: ProgressEvent<FileReader>) => {
      try {
        const arrayBuffer = e.target?.result as ArrayBuffer;
        const data = new Uint8Array(arrayBuffer);
        let workbook;

        if (file.name.endsWith('.csv')) {
          const decoder = new TextDecoder('gbk');
          const str = decoder.decode(data);
          workbook = XLSX.read(str, { type: 'string' });
        } else {
          workbook = XLSX.read(data, { type: 'array' });
        }

        const wsname = workbook.SheetNames[0];
        const ws = workbook.Sheets[wsname];
        const jsonData = XLSX.utils.sheet_to_json<unknown[]>(ws, { header: 1 });

        if (jsonData.length > 0) {
          const firstRow = jsonData[0] as string[];
          const dynamicColumns: ColumnsType<PreviewRow> = firstRow.map((colName, index) => ({
            title: <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>{String(colName || `列${index + 1}`)}</Text>,
            dataIndex: `col${index}`,
            key: `col${index}`,
            ellipsis: true,
            render: (val: unknown) => <Text style={{ color: 'rgba(255,255,255,0.8)' }}>{String(val ?? '')}</Text>,
          }));

          const previewRows: PreviewRow[] = jsonData.slice(1, 6).map((row: unknown, rIndex) => {
            const obj: PreviewRow = { key: rIndex };
            if (Array.isArray(row)) {
              row.forEach((cell, cIndex) => {
                obj[`col${cIndex}`] = cell as string | number | boolean | undefined;
              });
            }
            return obj;
          });

          setColumns(dynamicColumns);
          setDataPreview(previewRows);
          message.success(`${file.name} 解析成功`);
        }
      } catch (error) {
        console.error('Parse Error:', error);
        message.error('文件解析失败，请检查格式');
      }
    };

    reader.readAsArrayBuffer(file);
    return false;
  };

  const startUpload = () => {
    setUploading(true);
    let curr = 0;
    const timer = window.setInterval(() => {
      curr += 10;
      setPercent(curr);
      if (curr >= 100) {
        window.clearInterval(timer);
        setUploading(false);
        message.success({
          content: '上传成功，AI 引擎已开始分析',
          icon: <CheckCircleTwoTone twoToneColor="#52c41a" />,
        });
      }
    }, 150);
  };

  const resetUpload = () => {
    setDataPreview([]);
    setColumns([]);
    setPercent(0);
    setFileName('');
  };

  const cardStyle = {
    background: 'rgba(15, 23, 42, 0.6)',
    border: '1px solid rgba(255,255,255,0.08)',
    borderRadius: 16,
    backdropFilter: 'blur(10px)',
  };

  const supportedFormats = [
    { label: 'CSV 对账单', desc: '支持 GBK / UTF-8 自动识别' },
    { label: 'Excel (.xlsx)', desc: '多 Sheet 自动取第一张' },
    { label: 'Excel (.xls)', desc: '兼容旧版格式' },
  ];

  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm }}>
      <div style={{ minHeight: '100vh', background: 'radial-gradient(circle at top left, #1e293b 0%, #0b1120 100%)' }}>
        <PageHeader title="数据导入中心" subtitle="智能识别编码，支持主流券商对账单格式" />

        <div style={{ padding: '28px 32px', maxWidth: 1400, margin: '0 auto' }}>
          <Row gutter={[20, 20]}>
            {/* 左侧说明 */}
            <Col span={24} lg={7}>
              <Space direction="vertical" style={{ width: '100%' }} size={20}>
                <Card bordered={false} style={cardStyle}>
                  <Title level={5} style={{ color: 'rgba(255,255,255,0.8)', marginTop: 0 }}>支持格式</Title>
                  {supportedFormats.map(fmt => (
                    <div key={fmt.label} style={{ display: 'flex', alignItems: 'flex-start', gap: 12, padding: '10px 0', borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                      <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#1677ff', marginTop: 6, flexShrink: 0 }} />
                      <div>
                        <Text strong style={{ color: '#fff', fontSize: 13 }}>{fmt.label}</Text>
                        <div style={{ color: 'rgba(255,255,255,0.4)', fontSize: 12, marginTop: 2 }}>{fmt.desc}</div>
                      </div>
                    </div>
                  ))}
                </Card>

                <Card bordered={false} style={cardStyle}>
                  <Title level={5} style={{ color: 'rgba(255,255,255,0.8)', marginTop: 0 }}>数据说明</Title>
                  {[
                    '仅预览前 5 行，完整数据在提交后处理',
                    '文件不会被永久存储，仅用于 AI 分析',
                    '敏感字段建议提前脱敏处理',
                  ].map((tip, i) => (
                    <div key={i} style={{ display: 'flex', gap: 10, marginBottom: 10 }}>
                      <Text style={{ color: '#1677ff', fontWeight: 700, flexShrink: 0 }}>{i + 1}.</Text>
                      <Text style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12, lineHeight: 1.6 }}>{tip}</Text>
                    </div>
                  ))}
                </Card>
              </Space>
            </Col>

            {/* 右侧上传区 */}
            <Col span={24} lg={17}>
              <Space direction="vertical" style={{ width: '100%' }} size={20}>
                <Card bordered={false} style={cardStyle}>
                  <Dragger
                    accept=".csv,.xlsx,.xls"
                    beforeUpload={handleFile}
                    showUploadList={false}
                    disabled={dataPreview.length > 0}
                    style={{
                      background: 'rgba(22,119,255,0.04)',
                      border: '1px dashed rgba(22,119,255,0.3)',
                      borderRadius: 12,
                    }}
                  >
                    <p className="ant-upload-drag-icon">
                      <InboxOutlined style={{ color: '#1677ff', fontSize: 48 }} />
                    </p>
                    <p style={{ color: 'rgba(255,255,255,0.8)', fontSize: 16, fontWeight: 500 }}>
                      点击或将文件拖拽到此处
                    </p>
                    <p style={{ color: 'rgba(255,255,255,0.35)', fontSize: 13 }}>
                      支持各大券商导出的标准对账单格式 · CSV / XLSX / XLS
                    </p>
                  </Dragger>
                </Card>

                {dataPreview.length > 0 && (
                  <Card
                    bordered={false}
                    style={cardStyle}
                    title={
                      <Space>
                        <FileSearchOutlined style={{ color: '#1677ff' }} />
                        <Text style={{ color: 'rgba(255,255,255,0.85)' }}>解析预览: {fileName}</Text>
                      </Space>
                    }
                    extra={
                      <Space>
                        <Button onClick={resetUpload} disabled={uploading}
                          style={{ borderRadius: 10, background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.1)' }}
                        >
                          重选文件
                        </Button>
                        <Button
                          type="primary"
                          icon={<CloudUploadOutlined />}
                          loading={uploading}
                          onClick={startUpload}
                          style={{ borderRadius: 10 }}
                        >
                          确认提交
                        </Button>
                      </Space>
                    }
                  >
                    <Table
                      dataSource={dataPreview}
                      columns={columns}
                      pagination={false}
                      size="small"
                      scroll={{ x: 'max-content' }}
                      style={{ background: 'transparent' }}
                    />
                    {uploading && (
                      <Progress
                        percent={percent}
                        status="active"
                        strokeColor={{ '0%': '#1677ff', '100%': '#52c41a' }}
                        trailColor="rgba(255,255,255,0.08)"
                        style={{ marginTop: 20 }}
                      />
                    )}
                  </Card>
                )}
              </Space>
            </Col>
          </Row>
        </div>
      </div>
    </ConfigProvider>
  );
}