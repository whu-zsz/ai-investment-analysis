const Dashboard = () => {
  return (
    <div>
      <h2 className="text-3xl font-bold mb-8">欢迎回来，张盛哲 👋</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white p-8 rounded-3xl shadow">
          <div className="text-gray-500">总资产</div>
          <div className="text-5xl font-bold text-green-600 mt-2">¥ 128,450</div>
        </div>
        <div className="bg-white p-8 rounded-3xl shadow">
          <div className="text-gray-500">本月盈亏</div>
          <div className="text-5xl font-bold text-green-600 mt-2">+12.4%</div>
        </div>
        <div className="bg-white p-8 rounded-3xl shadow">
          <div className="text-gray-500">风险等级</div>
          <div className="text-5xl font-bold text-orange-500 mt-2">中</div>
        </div>
      </div>

      <div className="mt-10 text-center text-gray-400">
        这里后面会放 ECharts 图表
      </div>
    </div>
  )
}

export default Dashboard