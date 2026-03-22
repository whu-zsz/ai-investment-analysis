import { alerts, reportCards } from '../data/mock';

const Analysis = () => {
  return (
    <div className="page-grid">
      <section className="card">
        <div className="section-head">
          <div>
            <p className="eyebrow">AI Summary</p>
            <h2>自动生成的投资行为分析</h2>
            <p className="muted">按照需求文档中的总结、偏好分析、风险评估和模式识别进行页面组织。</p>
          </div>
          <span className="pill">最近分析时间：2026-03-20 16:40</span>
        </div>

        <div className="info-grid">
          {reportCards.map((item) => (
            <article key={item.title} className="mini-card">
              <p className="eyebrow">模块结果</p>
              <h3>{item.title}</h3>
              <p className="muted">{item.text}</p>
            </article>
          ))}
        </div>
      </section>

      <div className="split-grid">
        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Signals</p>
              <h2>行为标签</h2>
            </div>
          </div>

          <div className="tag-row">
            <span className="pill">偏成长</span>
            <span className="pill">中等风险承受</span>
            <span className="pill">偏爱 ETF 配置</span>
            <span className="pill">存在情绪化卖出</span>
            <span className="pill">适合月度复盘</span>
          </div>
        </section>

        <section className="card">
          <div className="section-head">
            <div>
              <p className="eyebrow">Warnings</p>
              <h2>风险预警</h2>
            </div>
          </div>

          <ul className="alert-list">
            {alerts.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  );
};

export default Analysis;
