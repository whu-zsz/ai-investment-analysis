const Prediction = () => {
  return (
    <div>
      <h2 className="text-3xl font-bold mb-6">🔮 AI趋势预测</h2>
      <div className="bg-white rounded-3xl p-8 shadow">
        <p className="text-xl text-gray-700 mb-4">未来30天预测（置信度 78%）</p>
        <div className="text-5xl font-bold text-green-600">+8.6%</div>
        <p className="text-gray-500 mt-2">建议：轻仓加仓科技板块，观察大盘</p>
      </div>
      <div className="mt-10 text-center text-gray-400">
        （后面加实时折线图 + Yahoo Finance 数据）
      </div>
    </div>
  )
}

export default Prediction