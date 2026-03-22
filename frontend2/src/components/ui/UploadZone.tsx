interface UploadZoneProps {
  onPick: () => void;
}

const UploadZone = ({ onPick }: UploadZoneProps) => {
  return (
    <button type="button" className="upload-zone" onClick={onPick}>
      <div className="upload-icon">+</div>
      <strong>拖拽投资记录到这里，或点击选择文件</strong>
      <p>支持 CSV、XLS、XLSX。当前为 UI demo，后续可接真实上传、解析和进度条。</p>
      <span className="pill">自动识别日期 / 代码 / 金额 / 交易方向</span>
    </button>
  );
};

export default UploadZone;
