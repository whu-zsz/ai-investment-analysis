import { alerts, forecastBars, holdings, summaryStats } from '../data/mock';

const Dashboard = () => {
  return (
    <div className="page-grid">
      <section className="stats-grid">
        {summaryStats.map((item) => (
          <article key={item.label} className={`card metric-card ${item.tone}`}>
            <span className="metric-label">{item.label}</span>
            <strong>{item.value}</strong>
            <p className="muted">{item.hint}</p>
          </article>
        ))}
      </section>

      <div className="split-grid">
        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Trend</p>
              <h2>近 6 个月收益与预测对比</h2>
              <p className="muted">当前用轻量柱状视觉模拟图表，后续可替换为 ECharts。</p>
            </div>
            <button className="button button-secondary">导出周报</button>
          </div>

          <div className="bars">
            {forecastBars.map((item) => (
              <div key={item.label} className="bar-item">
                <div className="bar-track">
                  <div className="bar actual" style={{ height: `${item.actual * 2}px` }} />
                  <div className="bar forecast" style={{ height: `${item.forecast * 2}px` }} />
                </div>
                <strong>{item.label}</strong>
              </div>
            ))}
          </div>
        </section>

        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Portfolio</p>
              <h2>当前资产分布</h2>
            </div>
          </div>

          <div className="stack-list">
            {holdings.map((item) => (
              <div key={item.code} className="holding-row">
                <div>
                  <strong>
                    {item.name} · {item.code}
                  </strong>
                  <div className="progress">
                    <span style={{ width: `${item.ratio}%` }} />
                  </div>
                </div>
                <span>{item.ratio}%</span>
                <span>{item.profit}</span>
              </div>
            ))}
          </div>
        </section>
      </div>

      <section className="card">
        <div className="section-head">
          <div>
            <p className="eyebrow">Risk Radar</p>
            <h2>AI 今日提醒</h2>
          </div>
          <span className="pill">适合接入实时分析接口</span>
        </div>

        <ul className="alert-list">
          {alerts.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>
    </div>
  );
};

export default Dashboard;
