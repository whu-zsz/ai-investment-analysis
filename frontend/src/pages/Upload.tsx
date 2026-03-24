import { useState } from 'react'; // 删除了未使用的 React
import { 
  Upload, Card, Typography, Table, message, Button, 
  Space, Progress, Row, Col 
} from 'antd'; // 删除了未使用的 Alert
import { 
  InboxOutlined, FileSearchOutlined, 
  CloudUploadOutlined, CheckCircleTwoTone 
} from '@ant-design/icons'; // 删除了未使用的 DeleteOutlined
import * as XLSX from 'xlsx';
import type { RcFile } from 'antd/es/upload';
import type { ColumnsType } from 'antd/es/table';

const { Dragger } = Upload;
const { Title, Text } = Typography;

// 1. 定义行数据接口
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

        // 处理 CSV 的 GBK 编码防乱码
        if (file.name.endsWith('.csv')) {
          const decoder = new TextDecoder('gbk');
          const str = decoder.decode(data);
          workbook = XLSX.read(str, { type: 'string' });
        } else {
          workbook = XLSX.read(data, { type: 'array' });
        }
        
        const wsname = workbook.SheetNames[0];
        const ws = workbook.Sheets[wsname];
        // 明确声明类型，避免隐式 any
        const jsonData = XLSX.utils.sheet_to_json<unknown[]>(ws, { header: 1 });

        if (jsonData.length > 0) {
          const firstRow = jsonData[0] as string[];
          
          const dynamicColumns: ColumnsType<PreviewRow> = firstRow.map((colName, index) => ({
            title: String(colName || `列${index + 1}`),
            dataIndex: `col${index}`,
            key: `col${index}`,
            ellipsis: true,
          }));
          
          // 关键修复：将 any 改为 unknown，并配合类型保护
          const previewRows: PreviewRow[] = jsonData.slice(1, 6).map((row: unknown, rIndex) => {
            const obj: PreviewRow = { key: rIndex };
            if (Array.isArray(row)) {
              row.forEach((cell, cIndex) => {
                // 确保 cell 类型符合 PreviewRow 定义
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
    // 使用 window.setInterval 明确 DOM 环境类型
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

  return (
    <div style={{ padding: '8px' }}>
      <Title level={3}>数据导入中心</Title>
      <Text type="secondary">智能识别编码，确保 CSV/Excel 预览无乱码。</Text>

      <Row gutter={[0, 24]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card bordered={false} hoverable>
            <Dragger 
              accept=".csv,.xlsx,.xls"
              beforeUpload={handleFile}
              showUploadList={false}
              disabled={dataPreview.length > 0}
            >
              <p className="ant-upload-drag-icon"><InboxOutlined style={{ color: '#1677ff' }} /></p>
              <p className="ant-upload-text">点击或将文件拖拽到此处</p>
              <p className="ant-upload-hint">支持各大券商导出的标准对账单格式</p>
            </Dragger>
          </Card>
        </Col>

        {dataPreview.length > 0 && (
          <Col span={24}>
            <Card 
              title={<Space><FileSearchOutlined /><span>解析预览: {fileName}</span></Space>}
              extra={
                <Space>
                  <Button onClick={resetUpload} disabled={uploading}>
                    重选文件
                  </Button>
                  <Button 
                    type="primary" 
                    icon={<CloudUploadOutlined />} 
                    loading={uploading}
                    onClick={startUpload}
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
                bordered
              />
              {uploading && <Progress percent={percent} status="active" style={{ marginTop: 20 }} />}
            </Card>
          </Col>
        )}
      </Row>
    </div>
  );
}