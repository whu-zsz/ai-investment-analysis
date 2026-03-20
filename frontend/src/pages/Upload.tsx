const Upload = () => {
  return (
    <div>
      <h2 className="text-3xl font-bold mb-6">📤 上传投资记录</h2>
      <div className="border-2 border-dashed border-blue-300 bg-white rounded-3xl p-20 text-center hover:border-blue-500 transition-all">
        <div className="text-6xl mb-4">📁</div>
        <p className="text-xl text-gray-600 mb-2">拖拽 CSV / Excel 文件到此处</p>
        <p className="text-sm text-gray-400">或点击下方按钮选择文件</p>
        <button className="mt-6 px-8 py-3 bg-blue-600 text-white rounded-2xl hover:bg-blue-700">
          选择文件
        </button>
      </div>
      <p className="text-center text-gray-400 mt-4">支持 .csv、.xlsx 格式，后端会自动清洗</p>
    </div>
  )
}

export default Upload