import { historyRows } from '../data/mock';

const History = () => {
  return (
    <div className="page-grid">
      <section className="card">
        <div className="history-head">
          <div>
            <p className="eyebrow">Archive</p>
            <h2>交易与上传历史</h2>
          </div>
          <div className="tag-row">
            <span className="pill">全部</span>
            <span className="pill">已分析</span>
            <span className="pill">待复核</span>
          </div>
        </div>

        <div className="table-wrap">
          <table className="history-table">
            <thead>
              <tr>
                <th>日期</th>
                <th>类型</th>
                <th>代码</th>
                <th>名称</th>
                <th>数量</th>
                <th>单价</th>
                <th>总额</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {historyRows.map((row) => (
                <tr key={`${row[0]}-${row[2]}`}>
                  {row.map((cell) => (
                    <td key={cell}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
};

export default History;
