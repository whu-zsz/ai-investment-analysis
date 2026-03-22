import { useState } from 'react';
import UploadZone from '../components/ui/UploadZone';
import { uploadPreview } from '../data/mock';

const Upload = () => {
  const [fileName, setFileName] = useState('investment_record_2026_q1.xlsx');

  return (
    <div className="page-grid">
      <section className="card">
        <div className="section-head">
          <div>
            <p className="eyebrow">Import</p>
            <h2>上传投资记录</h2>
            <p className="muted">支持 CSV / Excel，后续可接数据清洗、字段映射和上传进度。</p>
          </div>
          <button className="button button-primary">开始分析</button>
        </div>

        <UploadZone onPick={() => setFileName('portfolio_upload_demo.xlsx')} />
      </section>

      <div className="split-grid">
        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Preview</p>
              <h2>字段识别结果</h2>
            </div>
            <span className="pill">{fileName}</span>
          </div>

          <div className="stack-list">
            {uploadPreview.map((item) => (
              <div key={item.field} className="field-row">
                <strong>{item.field}</strong>
                <span className={item.status === '已识别' ? 'status success' : 'status pending'}>
                  {item.status}
                </span>
                <span className="muted">{item.note}</span>
              </div>
            ))}
          </div>
        </section>

        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Workflow</p>
              <h2>上传后处理流程</h2>
            </div>
          </div>

          <div className="stack-list">
            <div className="field-row">
              <strong>01 文件上传</strong>
              <span className="status success">完成</span>
              <span className="muted">支持 CSV、XLS、XLSX</span>
            </div>
            <div className="field-row">
              <strong>02 字段映射</strong>
              <span className="status pending">待确认</span>
              <span className="muted">统一买卖方向、证券类型与金额格式</span>
            </div>
            <div className="field-row">
              <strong>03 数据清洗</strong>
              <span className="status pending">待执行</span>
              <span className="muted">去重、补全空值、识别异常交易</span>
            </div>
            <div className="field-row">
              <strong>04 AI 分析</strong>
              <span className="status pending">待执行</span>
              <span className="muted">生成总结、风险评估与趋势预测</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
};

export default Upload;
