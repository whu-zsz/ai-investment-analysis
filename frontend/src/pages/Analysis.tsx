const Analysis = () => {
  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h2 className="text-3xl font-bold">📊 AI分析报告</h2>
        <button className="px-6 py-3 bg-green-600 text-white rounded-2xl hover:bg-green-700">
          重新生成AI分析
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 总结卡片 */}
        <div className="bg-white p-8 rounded-3xl shadow">
          <h3 className="text-lg font-semibold mb-4">投资总结</h3>
          <p className="text-gray-600">总盈亏：<span className="text-green-600 font-bold">+¥28,450</span></p>
          <p className="text-gray-600 mt-2">交易次数：47 次</p>
          <p className="text-gray-600 mt-2">投资偏好：成长型 + 科技股</p>
        </div>

        {/* 风险卡片 */}
        <div className="bg-white p-8 rounded-3xl shadow">
          <h3 className="text-lg font-semibold mb-4">风险评估</h3>
          <div className="text-6xl font-bold text-orange-500">中</div>
          <p className="mt-4 text-gray-600">主要风险：单一行业集中度高</p>
        </div>
      </div>

      <div className="mt-8 text-center text-gray-400">
        （后面我会给你完整 ECharts 版本）
      </div>
    </div>
  )
}

export default Analysis