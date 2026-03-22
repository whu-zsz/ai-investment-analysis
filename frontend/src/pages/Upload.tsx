import { Upload, Card, Typography } from 'antd';
import { InboxOutlined } from '@ant-design/icons';

export default function UploadPage() {
  return (
    <Card title="上传投资记录">
      <Upload.Dragger>
        <p><InboxOutlined /></p>
        <p>拖拽文件至此处</p>
      </Upload.Dragger>
    </Card>
  );
}