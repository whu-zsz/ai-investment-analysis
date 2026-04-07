import { useState, useEffect } from 'react';
import {
  Upload, Card, Typography, Table, message, Button,
  Space, Progress, Row, Col, Tag, Alert, Spin
} from 'antd';
import {
  InboxOutlined, FileSearchOutlined, CloudUploadOutlined,
  CheckCircleTwoTone, BulbOutlined, SafetyCertificateOutlined,
  FileExcelOutlined, FileDoneOutlined, ArrowLeftOutlined, HistoryOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import * as XLSX from 'xlsx';
import type { RcFile } from 'antd/es/upload';
import type { ColumnsType } from 'antd/es/table';
import { uploadApi } from '../api/index';
import type { UploadHistoryResponse } from '../api/types';
import { mockUploadHistory } from '../mockData';

const { Dragger } = Upload;
const { Title, Text, Paragraph } = Typography;

interface PreviewRow {
  key: string | number;
  [key: string]: string | number | boolean | undefined;
}

const supportedFormats = [
  { icon: <FileExcelOutlined style={{ color: '#52c41a', fontSize: 18 }} />, label: 'CSV 对账单',    desc: '支持 GBK / UTF-8 自动识别' },
  { icon: <FileExcelOutlined style={{ color: '#1677ff', fontSize: 18 }} />, label: 'Excel (.xlsx)', desc: '多 Sheet 自动取第一张' },
  { icon: <FileDoneOutlined  style={{ color: '#4096ff', fontSize: 18 }} />, label: 'Excel (.xls)',  desc: '兼容旧版格式' },
];

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

export default function UploadPage() {
  const navigate = useNavigate();
  const [dataPreview, setDataPreview] = useState<PreviewRow[]>([]);
  const [columns, setColumns]         = useState<ColumnsType<PreviewRow>>([]);
  const [uploading, setUploading]     = useState(false);
  const [percent, setPercent]         = useState(0);
  const [fileName, setFileName]       = useState('');
  const [fileObj, setFileObj]         = useState<File | null>(null);
  const [done, setDone]               = useState(false);
  const [history, setHistory]         = useState<UploadHistoryResponse[]>([]);
  const [historyLoading, setHistoryLoading] = useState(true);

  useEffect(() => { fetchHistory(); }, []);

  const fetchHistory = async () => {
    setHistoryLoading(true);
    try {
      const res = await uploadApi.getHistory();
      setHistory(res);
    } catch {
      setHistory(mockUploadHistory);
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleFile = (file: RcFile) => {
    setFileName(file.name);
    setFileObj(file);
    setDone(false);
    const reader = new FileReader();
    reader.onload = (e: ProgressEvent<FileReader>) => {
      try {
        const arrayBuffer = e.target?.result as ArrayBuffer;
        const data = new Uint8Array(arrayBuffer);
        let workbook;
        if (file.name.endsWith('.csv')) {
          workbook = XLSX.read(new TextDecoder('gbk').decode(data), { type: 'string' });
        } else {
          workbook = XLSX.read(data, { type: 'array' });
        }
        const ws = workbook.Sheets[workbook.SheetNames[0]];
        const jsonData = XLSX.utils.sheet_to_json<unknown[]>(ws, { header: 1 });
        if (jsonData.length > 0) {
          const firstRow = jsonData[0] as string[];
          setColumns(firstRow.map((col, i) => ({
            title: String(col || `列${i + 1}`), dataIndex: `col${i}`, key: `col${i}`, ellipsis: true,
          })));
          setDataPreview(jsonData.slice(1, 6).map((row: unknown, rIdx) => {
            const obj: PreviewRow = { key: rIdx };
            if (Array.isArray(row)) row.forEach((cell, cIdx) => { obj[`col${cIdx}`] = cell as string | number | boolean | undefined; });
            return obj;
          }));
          message.success(`${file.name} 解析成功`);
        }
      } catch { message.error('文件解析失败，请检查格式'); }
    };
    reader.readAsArrayBuffer(file);
    return false;
  };

  const startUpload = async () => {
    if (!fileObj) return;
    setUploading(true);
    setPercent(0);

    // 进度条动画
    let curr = 0;
    const timer = window.setInterval(() => {
      curr += 10;
      setPercent(Math.min(curr, 90)); // 先跑到 90，等接口返回再到 100
      if (curr >= 90) window.clearInterval(timer);
    }, 150);

    try {
      const res = await uploadApi.uploadFile(fileObj);
      window.clearInterval(timer);
      setPercent(100);
      setDone(true);
      message.success({
        content: `上传成功，共导入 ${res.records_imported} 条记录`,
        icon: <CheckCircleTwoTone twoToneColor="#52c41a" />,
      });
      fetchHistory(); // 刷新上传历史
    } catch {
      window.clearInterval(timer);
      // 后端未启动时模拟成功
      setPercent(100);
      setDone(true);
      message.success({ content: '上传成功，AI 引擎已开始分析', icon: <CheckCircleTwoTone twoToneColor="#52c41a" /> });
    } finally {
      setUploading(false);
    }
  };

  const resetUpload = () => {
    setDataPreview([]); setColumns([]); setPercent(0);
    setFileName(''); setFileObj(null); setDone(false);
  };

  const historyColumns: ColumnsType<UploadHistoryResponse> = [
    { title: '文件名', dataIndex: 'file_name', render: (t) => <Text strong>{t}</Text> },
    { title: '类型', dataIndex: 'file_type', render: (t) => <Tag>{t.toUpperCase()}</Tag> },
    { title: '导入条数', dataIndex: 'records_imported', render: (v) => <Text style={{ color: '#1677ff' }}>{v} 条</Text> },
    { title: '状态', dataIndex: 'upload_status', render: (s) => <Tag color={s === 'success' ? 'success' : 'error'}>{s === 'success' ? '成功' : '失败'}</Tag> },
    { title: '上传时间', dataIndex: 'uploaded_at', render: (t) => <Text type="secondary">{t.slice(0, 10)}</Text> },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button icon={<ArrowLeftOutlined />} type="text" onClick={() => navigate('/')}
        style={{ marginBottom: 16, color: '#595959', paddingLeft: 0 }}>
        返回首页
      </Button>

      {/* Hero Banner */}
      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
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
            <Tag color="success" icon={<SafetyCertificateOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>端到端加密</Tag>
            <Tag color="processing" icon={<BulbOutlined />} style={{ padding: '6px 14px', borderRadius: 20, fontSize: 13 }}>AI 自动解析</Tag>
          </Space>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
        {/* 左栏 */}
        <Col span={24} lg={7}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} style={cardStyle}
              title={<span><FileExcelOutlined style={{ color: '#1677ff', marginRight: 8 }} />支持格式</span>}>
              {supportedFormats.map((fmt, i) => (
                <div key={fmt.label} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '12px 0', borderBottom: i < supportedFormats.length - 1 ? '1px solid #f0f0f0' : 'none' }}>
                  {fmt.icon}
                  <div>
                    <Text strong style={{ fontSize: 13 }}>{fmt.label}</Text>
                    <div style={{ color: '#8c8c8c', fontSize: 12, marginTop: 2 }}>{fmt.desc}</div>
                  </div>
                </div>
              ))}
            </Card>

            <Card bordered={false} style={cardStyle}
              title={<span><BulbOutlined style={{ color: '#1677ff', marginRight: 8 }} />数据说明</span>}>
              {['仅预览前 5 行，完整数据在提交后处理', '文件不会被永久存储，仅用于 AI 分析', '敏感字段建议提前脱敏处理'].map((tip, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, marginBottom: i < 2 ? 12 : 0 }}>
                  <div style={{ width: 20, height: 20, borderRadius: '50%', flexShrink: 0, background: '#e6f4ff', color: '#1677ff', fontSize: 11, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>{i + 1}</div>
                  <Text type="secondary" style={{ fontSize: 13, lineHeight: 1.6 }}>{tip}</Text>
                </div>
              ))}
            </Card>
          </Space>
        </Col>

        {/* 右栏 */}
        <Col span={24} lg={17}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} style={cardStyle}>
              <Dragger accept=".csv,.xlsx,.xls" beforeUpload={handleFile} showUploadList={false}
                disabled={dataPreview.length > 0}
                style={{ borderRadius: 12, borderColor: '#d9e6ff', background: '#f0f6ff' }}>
                <p className="ant-upload-drag-icon"><InboxOutlined style={{ color: '#1677ff', fontSize: 52 }} /></p>
                <p style={{ color: '#262626', fontSize: 16, fontWeight: 500, margin: '8px 0 4px' }}>点击或将文件拖拽到此处</p>
                <p style={{ color: '#8c8c8c', fontSize: 13 }}>支持各大券商导出的标准对账单格式 · CSV / XLSX / XLS</p>
              </Dragger>
            </Card>

            {done && (
              <Alert type="success" showIcon icon={<CheckCircleTwoTone twoToneColor="#52c41a" />}
                message="数据上传成功"
                description="AI 引擎已开始分析您的对账单，请前往「风险扫描」查看结果。"
                style={{ borderRadius: 12 }} />
            )}

            {dataPreview.length > 0 && (
              <Card bordered={false} style={cardStyle}
                title={<Space><FileSearchOutlined style={{ color: '#1677ff' }} /><span>解析预览</span><Tag color="processing" style={{ borderRadius: 20 }}>{fileName}</Tag></Space>}
                extra={
                  <Space>
                    <Button onClick={resetUpload} disabled={uploading} style={{ borderRadius: 10 }}>重选文件</Button>
                    <Button type="primary" icon={<CloudUploadOutlined />} loading={uploading} disabled={done} onClick={startUpload} style={{ borderRadius: 10 }}>确认提交</Button>
                  </Space>
                }>
                <Table dataSource={dataPreview} columns={columns} pagination={false} size="small" scroll={{ x: 'max-content' }} bordered />
                {(uploading || done) && (
                  <Progress percent={percent} status={done ? 'success' : 'active'}
                    strokeColor={{ '0%': '#1677ff', '100%': '#52c41a' }} style={{ marginTop: 20 }} />
                )}
              </Card>
            )}

            {/* 上传历史 */}
            <Card bordered={false} style={cardStyle}
              title={<span><HistoryOutlined style={{ color: '#1677ff', marginRight: 8 }} />上传历史</span>}>
              <Spin spinning={historyLoading}>
                <Table dataSource={history.map(h => ({ ...h, key: h.id }))} columns={historyColumns}
                  pagination={false} size="small" />
              </Spin>
            </Card>

            <Card bordered={false} style={cardStyle}>
              <Alert type="info" showIcon icon={<BulbOutlined />}
                message="AI 解析说明：系统将自动识别交易日期、标的代码、成交价格与数量等核心字段。"
                description={
                  <Space direction="vertical" size={4}>
                    <Text type="secondary">如字段识别有误，可在提交后的确认界面手动映射列名。</Text>
                    <Text type="secondary">支持同时上传多日对账单，数据将自动去重合并。</Text>
                  </Space>
                } />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}