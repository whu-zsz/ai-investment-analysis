import { forecastBars, scenarioCards } from '../data/mock';

const Prediction = () => {
  return (
    <div className="page-grid">
      <section className="card">
        <div className="section-head">
          <div>
            <p className="eyebrow">Forecast</p>
            <h2>未来 30 天收益区间模拟</h2>
            <p className="muted">以可视化占位替代真实图表，方便你后续接 ECharts 与 AI 返回数据。</p>
          </div>
          <span className="pill">模型置信度 89%</span>
        </div>

        <div className="bars">
          {forecastBars.map((item) => (
            <div key={item.label} className="bar-item">
              <div className="bar-track">
                <div className="bar actual" style={{ height: `${item.actual * 2}px` }} />
                <div className="bar forecast" style={{ height: `${item.forecast * 2.1}px` }} />
              </div>
              <strong>{item.label}</strong>
            </div>
          ))}
        </div>
      </section>

      <section className="info-grid">
        {scenarioCards.map((item) => (
          <article key={item.title} className="mini-card">
            <p className="eyebrow">Scenario</p>
            <h3>{item.title}</h3>
            <strong className="scenario-range">{item.range}</strong>
            <p className="muted">{item.detail}</p>
          </article>
        ))}
      </section>
    </div>
  );
};

export default Prediction;
