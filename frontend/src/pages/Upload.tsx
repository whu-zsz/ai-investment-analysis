import { useMemo, useState } from 'react';
import {
  Upload, Card, Typography, Table, message, Button,
  Space, Progress, Row, Col, Tag, Alert
} from 'antd';
import {
  InboxOutlined, FileSearchOutlined,
  CloudUploadOutlined, CheckCircleTwoTone,
  BulbOutlined, SafetyCertificateOutlined,
  FileExcelOutlined, FileDoneOutlined, ArrowLeftOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import * as XLSX from 'xlsx';
import type { RcFile } from 'antd/es/upload';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../types';

const { Dragger } = Upload;
const { Title, Text, Paragraph } = Typography;

interface PreviewRow {
  key: string | number;
  [key: string]: string | number | boolean | undefined;
}

interface UploadHistoryItem {
  id: number;
  file_name: string;
  file_size: number;
  file_type: string;
  upload_status: string;
  records_imported: number;
  uploaded_at: string;
  processed_at: string;
}

const supportedFormats = [
  { icon: <FileExcelOutlined style={{ color: '#52c41a', fontSize: 18 }} />, label: 'CSV 对账单', desc: '支持 GBK / UTF-8 自动识别' },
  { icon: <FileExcelOutlined style={{ color: '#1677ff', fontSize: 18 }} />, label: 'Excel (.xlsx)', desc: '多 Sheet 自动取第一张' },
  { icon: <FileDoneOutlined style={{ color: '#4096ff', fontSize: 18 }} />, label: 'Excel (.xls)', desc: '兼容旧版格式' },
];

const tips = [
  '仅预览前 5 行，完整数据在提交后处理',
  '上传记录会进入系统导入历史，便于后续追踪',
  '敏感字段建议提前脱敏处理',
];

export default function UploadPage() {
  const navigate = useNavigate();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dataPreview, setDataPreview] = useState<PreviewRow[]>([]);
  const [columns, setColumns] = useState<ColumnsType<PreviewRow>>([]);
  const [uploading, setUploading] = useState(false);
  const [percent, setPercent] = useState(0);
  const [fileName, setFileName] = useState('');
  const [done, setDone] = useState(false);
  const [history, setHistory] = useState<UploadHistoryItem[]>([]);

  const handleFile = (file: RcFile) => {
    setSelectedFile(file);
    setFileName(file.name);
    setDone(false);
    const reader = new FileReader();

    reader.onload = (e: ProgressEvent<FileReader>) => {
      try {
        const arrayBuffer = e.target?.result as ArrayBuffer;
        const data = new Uint8Array(arrayBuffer);
        const workbook = file.name.endsWith('.csv')
          ? XLSX.read(new TextDecoder('gbk').decode(data), { type: 'string' })
          : XLSX.read(data, { type: 'array' });

        const ws = workbook.Sheets[workbook.SheetNames[0]];
        const jsonData = XLSX.utils.sheet_to_json<unknown[]>(ws, { header: 1 });

        if (jsonData.length > 0) {
          const firstRow = jsonData[0] as string[];

          setColumns(firstRow.map((colName, i) => ({
            title: String(colName || `列${i + 1}`),
            dataIndex: `col${i}`,
            key: `col${i}`,
            ellipsis: true,
          })));

          setDataPreview(
            jsonData.slice(1, 6).map((row: unknown, rIdx) => {
              const obj: PreviewRow = { key: rIdx };
              if (Array.isArray(row)) {
                row.forEach((cell, cIdx) => {
                  obj[`col${cIdx}`] = cell as string | number | boolean | undefined;
                });
              }
              return obj;
            })
          );
          message.success(`${file.name} 解析成功`);
        }
      } catch {
        message.error('文件解析失败，请检查格式');
      }
    };

    reader.readAsArrayBuffer(file);
    return false;
  };

  const loadUploadHistory = async () => {
    const response = await api.getUploadHistory();
    setHistory(response.data);
  };

  const startUpload = async () => {
    if (!selectedFile) {
      message.warning('请先选择要上传的文件');
      return;
    }

    setUploading(true);
    setPercent(30);
    try {
      const response = await api.uploadFile(selectedFile);
      setPercent(100);
      setDone(true);
      message.success({
        content: response.data.message || '上传成功，AI 引擎已开始分析',
        icon: <CheckCircleTwoTone twoToneColor="#52c41a" />,
      });
      await loadUploadHistory();
    } finally {
      setUploading(false);
    }
  };

  const resetUpload = () => {
    setSelectedFile(null);
    setDataPreview([]);
    setColumns([]);
    setPercent(0);
    setFileName('');
    setDone(false);
  };

  const historyColumns = useMemo<ColumnsType<UploadHistoryItem>>(() => ([
    { title: '文件名', dataIndex: 'file_name', key: 'file_name' },
    { title: '文件类型', dataIndex: 'file_type', key: 'file_type' },
    { title: '导入条数', dataIndex: 'records_imported', key: 'records_imported' },
    { title: '状态', dataIndex: 'upload_status', key: 'upload_status', render: value => <Tag>{value}</Tag> },
    { title: '上传时间', dataIndex: 'uploaded_at', key: 'uploaded_at' },
  ]), []);

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
          marginBottom: 24,
          borderRadius: 20,
          background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
          boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
        }}
        bodyStyle={{ padding: 28 }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">数据导入</Tag>
              <Tag color="blue">智能解析</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>数据同步中心</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              上传您的券商对账单，AI 引擎将自动识别格式、清洗数据并纳入分析模型。
            </Paragraph>
          </div>
          <Space wrap>
            <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              认证接口接入完成
            </Tag>
            <Tag color="processing" icon={<BulbOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>
              AI 自动解析
            </Tag>
          </Space>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        <Col span={24} lg={7}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} title={<span><FileExcelOutlined style={{ color: '#1677ff', marginRight: 8 }} />支持格式</span>} style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
              {supportedFormats.map(fmt => (
                <div key={fmt.label} style={{
                  display: 'flex', alignItems: 'center', gap: 14,
                  padding: '12px 0',
                  borderBottom: fmt.label !== 'Excel (.xls)' ? '1px solid #f0f0f0' : 'none',
                }}>
                  {fmt.icon}
                  <div>
                    <Text strong style={{ fontSize: 13 }}>{fmt.label}</Text>
                    <div style={{ color: '#8c8c8c', fontSize: 12, marginTop: 2 }}>{fmt.desc}</div>
                  </div>
                </div>
              ))}
            </Card>

            <Card bordered={false} title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />数据说明</span>} style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
              {tips.map((tip, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, marginBottom: i < tips.length - 1 ? 12 : 0 }}>
                  <div style={{
                    width: 20, height: 20, borderRadius: '50%', flexShrink: 0,
                    background: '#e6f4ff', color: '#1677ff',
                    fontSize: 11, fontWeight: 700,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                  }}>
                    {i + 1}
                  </div>
                  <Text type="secondary" style={{ fontSize: 13, lineHeight: 1.6 }}>{tip}</Text>
                </div>
              ))}
            </Card>
          </Space>
        </Col>

        <Col span={24} lg={17}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
              <Dragger
                accept=".csv,.xlsx,.xls"
                beforeUpload={handleFile}
                showUploadList={false}
                disabled={dataPreview.length > 0}
                style={{ borderRadius: 12, borderColor: '#d9e6ff', background: '#f0f6ff' }}
              >
                <p className="ant-upload-drag-icon">
                  <InboxOutlined style={{ color: '#1677ff', fontSize: 52 }} />
                </p>
                <p style={{ color: '#262626', fontSize: 16, fontWeight: 500, margin: '8px 0 4px' }}>
                  点击或将文件拖拽到此处
                </p>
                <p style={{ color: '#8c8c8c', fontSize: 13 }}>
                  支持各大券商导出的标准对账单格式 · CSV / XLSX / XLS
                </p>
              </Dragger>
            </Card>

            {fileName && (
              <Card bordered={false} style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
                <Space direction="vertical" style={{ width: '100%' }} size={12}>
                  <Space>
                    <FileSearchOutlined style={{ color: '#1677ff' }} />
                    <Text strong>{fileName}</Text>
                  </Space>
                  <Progress percent={percent} status={done ? 'success' : uploading ? 'active' : 'normal'} />
                  <Space>
                    <Button type="primary" icon={<CloudUploadOutlined />} loading={uploading} onClick={startUpload}>开始上传</Button>
                    <Button onClick={resetUpload}>重新选择</Button>
                    <Button onClick={loadUploadHistory}>刷新历史</Button>
                  </Space>
                </Space>
              </Card>
            )}

            {dataPreview.length > 0 && (
              <Card bordered={false} title="文件预览" style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
                <Table columns={columns} dataSource={dataPreview} pagination={false} size="small" />
              </Card>
            )}

            <Card bordered={false} title="上传历史" style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
              <Table rowKey="id" columns={historyColumns} dataSource={history} pagination={{ pageSize: 5 }} />
            </Card>

            <Card bordered={false} style={{ borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' }}>
              <Alert
                type="info"
                showIcon
                icon={<BulbOutlined />}
                message="文件将进入真实上传接口 /upload，并可在 /upload/history 中查看导入历史。"
              />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}
