import { useState, useEffect } from 'react';
import {
  Upload, Card, Typography, Table, message, Button,
  Space, Progress, Row, Col, Tag, Alert, Spin, Empty,
} from 'antd';
import {
  InboxOutlined, FileSearchOutlined, CloudUploadOutlined,
  CheckCircleTwoTone, BulbOutlined,
  FileExcelOutlined, FileDoneOutlined, ArrowLeftOutlined, HistoryOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import * as XLSX from 'xlsx';
import type { RcFile } from 'antd/es/upload';
import type { ColumnsType } from 'antd/es/table';
import { uploadApi } from '../api/index';
import type { UploadHistoryResponse } from '../api/types';

const { Dragger } = Upload;
const { Title, Text, Paragraph } = Typography;

interface PreviewRow {
  key: string | number;
  [key: string]: string | number | boolean | undefined;
}

const supportedFormats = [
  { icon: <FileExcelOutlined style={{ color: '#52c41a', fontSize: 18 }} />, label: 'CSV 对账单', desc: '建议使用 UTF-8 或常见券商导出格式' },
  { icon: <FileExcelOutlined style={{ color: '#1677ff', fontSize: 18 }} />, label: 'Excel (.xlsx)', desc: '服务端按第一张工作表解析' },
  { icon: <FileDoneOutlined style={{ color: '#4096ff', fontSize: 18 }} />, label: 'Excel (.xls)', desc: '兼容旧版 Excel 文件' },
];

const cardStyle = { borderRadius: 16, boxShadow: '0 6px 22px rgba(15,23,42,0.06)' };

export default function UploadPage() {
  const navigate = useNavigate();
  const [dataPreview, setDataPreview] = useState<PreviewRow[]>([]);
  const [columns, setColumns] = useState<ColumnsType<PreviewRow>>([]);
  const [uploading, setUploading] = useState(false);
  const [percent, setPercent] = useState(0);
  const [fileName, setFileName] = useState('');
  const [fileObj, setFileObj] = useState<File | null>(null);
  const [done, setDone] = useState(false);
  const [uploadResult, setUploadResult] = useState<{ message: string; recordsImported: number } | null>(null);
  const [history, setHistory] = useState<UploadHistoryResponse[]>([]);
  const [historyLoading, setHistoryLoading] = useState(true);
  const [historyError, setHistoryError] = useState('');

  useEffect(() => { void fetchHistory(); }, []);

  const fetchHistory = async () => {
    setHistoryLoading(true);
    setHistoryError('');
    try {
      const res = await uploadApi.getHistory();
      setHistory(Array.isArray(res) ? res : []);
    } catch (err: any) {
      setHistory([]);
      setHistoryError(err?.message ?? err?.data?.message ?? '上传历史加载失败');
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleFile = (file: RcFile) => {
    setFileName(file.name);
    setFileObj(file);
    setDone(false);
    setPercent(0);
    setUploadResult(null);
    const reader = new FileReader();
    reader.onload = (e: ProgressEvent<FileReader>) => {
      try {
        const arrayBuffer = e.target?.result as ArrayBuffer;
        const data = new Uint8Array(arrayBuffer);
        const workbook = file.name.endsWith('.csv')
          ? XLSX.read(new TextDecoder('utf-8', { fatal: false }).decode(data), { type: 'string' })
          : XLSX.read(data, { type: 'array' });
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
          message.success(`${file.name} 预览已生成`);
        }
      } catch {
        message.error('文件解析失败，请检查格式');
      }
    };
    reader.readAsArrayBuffer(file);
    return false;
  };

  const startUpload = async () => {
    if (!fileObj) return;
    setUploading(true);
    setPercent(0);

    let curr = 0;
    const timer = window.setInterval(() => {
      curr += 10;
      setPercent(Math.min(curr, 90));
      if (curr >= 90) window.clearInterval(timer);
    }, 150);

    try {
      const res = await uploadApi.uploadFile(fileObj);
      window.clearInterval(timer);
      setPercent(100);
      setDone(true);
      setUploadResult({ message: res.message, recordsImported: res.records_imported });
      message.success({
        content: `${res.message}，共导入 ${res.records_imported} 条记录`,
        icon: <CheckCircleTwoTone twoToneColor="#52c41a" />,
      });
      void fetchHistory();
    } catch (err: any) {
      window.clearInterval(timer);
      setPercent(0);
      setDone(false);
      setUploadResult(null);
      const msg = err?.message ?? err?.data?.message ?? '上传失败';
      message.error(msg);
    } finally {
      setUploading(false);
    }
  };

  const resetUpload = () => {
    setDataPreview([]);
    setColumns([]);
    setPercent(0);
    setFileName('');
    setFileObj(null);
    setDone(false);
    setUploadResult(null);
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

      <Card bordered={false} style={{
        marginBottom: 24, borderRadius: 20,
        background: 'linear-gradient(135deg, #0f172a 0%, #1677ff 65%, #69b1ff 100%)',
        boxShadow: '0 18px 40px rgba(22,119,255,0.18)',
      }} bodyStyle={{ padding: 28 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
          <div>
            <Space size={12} style={{ marginBottom: 12 }}>
              <Tag color="processing">交易导入</Tag>
              <Tag color="blue">对齐后端解析规则</Tag>
            </Space>
            <Title level={2} style={{ margin: 0, color: '#fff' }}>交易记录导入</Title>
            <Paragraph style={{ margin: '12px 0 0', color: 'rgba(255,255,255,0.82)', maxWidth: 600 }}>
              上传券商导出的 CSV、XLSX 或 XLS 文件，服务端会按当前解析规则导入交易记录。
            </Paragraph>
          </div>
        </div>
      </Card>

      <Row gutter={[16, 16]}>
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
              {['本地仅预览前 5 行，便于提交前核对列顺序。', '服务端当前按第一张工作表解析 Excel 文件。', '是否导入成功以服务端返回的 message 和导入条数为准。'].map((tip, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, marginBottom: i < 2 ? 12 : 0 }}>
                  <div style={{ width: 20, height: 20, borderRadius: '50%', flexShrink: 0, background: '#e6f4ff', color: '#1677ff', fontSize: 11, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>{i + 1}</div>
                  <Text type="secondary" style={{ fontSize: 13, lineHeight: 1.6 }}>{tip}</Text>
                </div>
              ))}
            </Card>
          </Space>
        </Col>

        <Col span={24} lg={17}>
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Card bordered={false} style={cardStyle}>
              <Dragger accept=".csv,.xlsx,.xls" beforeUpload={handleFile} showUploadList={false}
                disabled={dataPreview.length > 0}
                style={{ borderRadius: 12, borderColor: '#d9e6ff', background: '#f0f6ff' }}>
                <p className="ant-upload-drag-icon"><InboxOutlined style={{ color: '#1677ff', fontSize: 52 }} /></p>
                <p style={{ color: '#262626', fontSize: 16, fontWeight: 500, margin: '8px 0 4px' }}>点击或将文件拖拽到此处</p>
                <p style={{ color: '#8c8c8c', fontSize: 13 }}>支持 CSV / XLSX / XLS</p>
              </Dragger>
            </Card>

            {done && uploadResult && (
              <Alert type="success" showIcon icon={<CheckCircleTwoTone twoToneColor="#52c41a" />}
                message={uploadResult.message}
                description={`本次共导入 ${uploadResult.recordsImported} 条记录。`}
                style={{ borderRadius: 12 }} />
            )}

            {dataPreview.length > 0 && (
              <Card bordered={false} style={cardStyle}
                title={<Space><FileSearchOutlined style={{ color: '#1677ff' }} /><span>本地预览</span><Tag color="processing" style={{ borderRadius: 20 }}>{fileName}</Tag></Space>}
                extra={
                  <Space>
                    <Button onClick={resetUpload} disabled={uploading} style={{ borderRadius: 10 }}>重选文件</Button>
                    <Button type="primary" icon={<CloudUploadOutlined />} loading={uploading} disabled={done} onClick={startUpload} style={{ borderRadius: 10 }}>确认提交</Button>
                  </Space>
                }>
                <Alert
                  type="info"
                  showIcon
                  message="该预览仅用于本地检查，不代表服务端最终解析结果。"
                  style={{ marginBottom: 16, borderRadius: 12 }}
                />
                <Table dataSource={dataPreview} columns={columns} pagination={false} size="small" scroll={{ x: 'max-content' }} bordered />
                {(uploading || done) && (
                  <Progress percent={percent} status={done ? 'success' : 'active'}
                    strokeColor={{ '0%': '#1677ff', '100%': '#52c41a' }} style={{ marginTop: 20 }} />
                )}
              </Card>
            )}

            <Card bordered={false} style={cardStyle}
              title={<span><HistoryOutlined style={{ color: '#1677ff', marginRight: 8 }} />上传历史</span>}>
              <Spin spinning={historyLoading}>
                {historyError ? (
                  <Alert type="error" showIcon message={historyError} />
                ) : history.length ? (
                  <Table dataSource={history.map(h => ({ ...h, key: h.id }))} columns={historyColumns}
                    pagination={false} size="small" />
                ) : (
                  <Empty description="暂无上传历史" />
                )}
              </Spin>
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}
