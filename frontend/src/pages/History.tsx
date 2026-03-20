const History = () => {
  return (
    <div>
      <h2 className="text-3xl font-bold mb-6">📜 历史交易记录</h2>
      <div className="bg-white rounded-3xl overflow-hidden shadow">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-4 text-left">日期</th>
              <th className="px-6 py-4 text-left">类型</th>
              <th className="px-6 py-4 text-left">资产</th>
              <th className="px-6 py-4 text-right">金额</th>
              <th className="px-6 py-4 text-right">盈亏</th>
            </tr>
          </thead>
          <tbody className="text-gray-700">
            <tr className="border-t">
              <td className="px-6 py-4">2026-03-15</td>
              <td className="px-6 py-4">买入</td>
              <td className="px-6 py-4">腾讯控股</td>
              <td className="px-6 py-4 text-right">¥12,500</td>
              <td className="px-6 py-4 text-right text-green-600">+¥890</td>
            </tr>
            {/* 更多行可后续扩展 */}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default History