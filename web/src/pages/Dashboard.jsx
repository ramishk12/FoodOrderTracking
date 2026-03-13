import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import '../index.css';

/* ─── Formatters ─────────────────────────────── */

const fmt = {
  usd: (n) => new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0),
  usd0: (n) => new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD', maximumFractionDigits: 0,
  }).format(n ?? 0),
  num: (n) => new Intl.NumberFormat('en-US').format(Math.round(n ?? 0)),
  date: (s) => new Date(s).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
  time: (d) => d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' }),
};

const STATUS_COLORS = {
  pending:   '#c47c2b',
  preparing: '#2b5fa0',
  ready:     '#6b3fa0',
  delivered: '#2d7a4f',
  cancelled: '#a02b2b',
};

/* ─── useCountUp ─────────────────────────────── */

function useCountUp(target, duration = 800) {
  const [val, setVal] = useState(0);
  useEffect(() => {
    if (target == null) return;
    let start = null;
    const raf = (ts) => {
      if (!start) start = ts;
      const p = Math.min((ts - start) / duration, 1);
      const ease = 1 - Math.pow(1 - p, 3);
      setVal(target * ease);
      if (p < 1) requestAnimationFrame(raf);
      else setVal(target);
    };
    requestAnimationFrame(raf);
  }, [target]);
  return val;
}

/* ─── KpiCard ────────────────────────────────── */

function KpiCard({ label, value, sub, isCurrency, accent, delay }) {
  const animated = useCountUp(value);
  const display = isCurrency ? fmt.usd(animated) : fmt.num(animated);
  return (
    <div className="kpi-card" style={{ animationDelay: `${delay}ms` }}>
      <div className="kpi-label">{label}</div>
      <div className={`kpi-value${accent ? ' accent' : ''}`}>{display}</div>
      {sub && <div className="kpi-sub">{sub}</div>}
      <div className="kpi-rule" />
    </div>
  );
}

/* ─── TrendChart ─────────────────────────────── */

function TrendChart({ data, mode }) {
  const [hovered, setHovered] = useState(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const t = setTimeout(() => setReady(true), 80);
    return () => clearTimeout(t);
  }, [data]);

  if (!data?.length) return <div className="db-empty">No trend data available</div>;

  const W = 720, H = 160;
  const PAD = { t: 12, r: 12, b: 32, l: 56 };
  const iW = W - PAD.l - PAD.r;
  const iH = H - PAD.t - PAD.b;

  const vals = data.map((d) => mode === 'revenue' ? d.revenue : d.orders);
  const maxV = Math.max(...vals, 1);
  const step = Math.ceil(data.length / 7);

  const pts = data.map((d, i) => ({
    x: PAD.l + (i / (data.length - 1)) * iW,
    y: PAD.t + iH - (vals[i] / maxV) * iH,
    d,
    v: vals[i],
  }));

  // Smooth bezier path
  const bezierPath = pts.reduce((acc, p, i) => {
    if (i === 0) return `M ${p.x},${p.y}`;
    const prev = pts[i - 1];
    const cx = (prev.x + p.x) / 2;
    return `${acc} C ${cx},${prev.y} ${cx},${p.y} ${p.x},${p.y}`;
  }, '');

  const areaPath = `${bezierPath} L ${pts[pts.length - 1].x},${PAD.t + iH} L ${PAD.l},${PAD.t + iH} Z`;

  const yTicks = [0, 0.5, 1].map((t) => ({
    y: PAD.t + iH - t * iH,
    label: mode === 'revenue' ? fmt.usd0(maxV * t) : fmt.num(maxV * t),
  }));

  const hovPt = hovered !== null ? pts[hovered] : null;

  return (
    <div className="chart-wrap">
      <svg
        viewBox={`0 0 ${W} ${H}`}
        style={{ width: '100%', height: 'auto', display: 'block', overflow: 'visible', cursor: 'crosshair' }}
        onMouseLeave={() => setHovered(null)}
      >
        <defs>
          <linearGradient id="area-grad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#c47c2b" stopOpacity="0.18" />
            <stop offset="100%" stopColor="#c47c2b" stopOpacity="0" />
          </linearGradient>
          <clipPath id="chart-clip">
            <rect x={PAD.l} y={PAD.t} width={iW} height={iH} />
          </clipPath>
        </defs>

        {/* Y grid + labels */}
        {yTicks.map((tk, i) => (
          <g key={i}>
            <line x1={PAD.l} y1={tk.y} x2={PAD.l + iW} y2={tk.y}
              stroke="#d4c9b4" strokeWidth="1" strokeDasharray={i === 0 ? 'none' : '3 3'} />
            <text x={PAD.l - 8} y={tk.y + 4} textAnchor="end"
              fill="#8a7060" fontSize="9.5" fontFamily="'IBM Plex Mono', monospace">
              {tk.label}
            </text>
          </g>
        ))}

        {/* Area */}
        <path d={areaPath} fill="url(#area-grad)" clipPath="url(#chart-clip)" />

        {/* Line */}
        <path
          d={bezierPath}
          fill="none"
          stroke="#c47c2b"
          strokeWidth="2"
          strokeLinecap="round"
          clipPath="url(#chart-clip)"
          style={{
            strokeDasharray: ready ? 'none' : 3000,
            strokeDashoffset: ready ? 0 : 3000,
            transition: 'stroke-dashoffset 1.4s cubic-bezier(0.4,0,0.2,1)',
          }}
        />

        {/* X labels */}
        {pts.filter((_, i) => i % step === 0 || i === pts.length - 1).map((p, i) => (
          <text key={i} x={p.x} y={H - 4} textAnchor="middle"
            fill="#8a7060" fontSize="9.5" fontFamily="'IBM Plex Mono', monospace">
            {fmt.date(p.d.date)}
          </text>
        ))}

        {/* Hover zones */}
        {pts.map((p, i) => (
          <rect key={i}
            x={p.x - iW / data.length / 2} y={PAD.t}
            width={iW / data.length} height={iH}
            fill="transparent"
            onMouseEnter={() => setHovered(i)}
          />
        ))}

        {/* Hover indicator */}
        {hovPt && (
          <g>
            <line x1={hovPt.x} y1={PAD.t} x2={hovPt.x} y2={PAD.t + iH}
              stroke="#c47c2b" strokeWidth="1" strokeDasharray="3 2" opacity="0.5" />
            <circle cx={hovPt.x} cy={hovPt.y} r="4.5"
              fill="white" stroke="#c47c2b" strokeWidth="2" />
          </g>
        )}
      </svg>

      {/* Floating tooltip */}
      {hovPt && (() => {
        const leftPct = ((hovPt.x - PAD.l) / iW) * 100;
        return (
          <div className="chart-tooltip"
            style={{ left: `clamp(60px, ${leftPct}%, calc(100% - 60px))`, top: `${(hovPt.y / H) * 100}%` }}>
            <div className="chart-tooltip-date">{fmt.date(data[hovered].date)}</div>
            <div className="chart-tooltip-row">
              <span>Revenue</span>
              <span className="chart-tooltip-val">{fmt.usd(data[hovered].revenue)}</span>
            </div>
            <div className="chart-tooltip-row">
              <span>Orders</span>
              <span className="chart-tooltip-val">{data[hovered].orders}</span>
            </div>
          </div>
        );
      })()}
    </div>
  );
}

/* ─── StatusBars ─────────────────────────────── */

function StatusBars({ statusMap }) {
  const [ready, setReady] = useState(false);
  useEffect(() => {
    const t = setTimeout(() => setReady(true), 150);
    return () => clearTimeout(t);
  }, []);

  const total = Object.values(statusMap).reduce((a, b) => a + b, 0);
  if (total === 0) return <div className="db-empty">No orders yet</div>;

  const order = ['pending', 'preparing', 'ready', 'delivered', 'cancelled'];
  const entries = order
    .filter((s) => statusMap[s] > 0)
    .map((s) => ({ status: s, count: statusMap[s], pct: statusMap[s] / total }));

  return (
    <div className="status-grid">
      {entries.map(({ status, count, pct }, i) => (
        <div key={status} className="status-row" style={{ animationDelay: `${i * 50}ms` }}>
          <div className="status-name">
            <span className="status-dot" style={{ background: STATUS_COLORS[status] || '#999' }} />
            {status}
          </div>
          <div className="status-track">
            <div className="status-fill" style={{
              background: STATUS_COLORS[status] || '#999',
              width: ready ? `${pct * 100}%` : '0%',
              transitionDelay: `${i * 80}ms`,
            }} />
          </div>
          <div className="status-count">{count}</div>
        </div>
      ))}
      <div className="db-divider" />
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 4, fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--muted)' }}>
        <span>Total</span>
        <span style={{ fontWeight: 600, color: 'var(--espresso)' }}>{total}</span>
      </div>
    </div>
  );
}

/* ─── BestItems ──────────────────────────────── */

function BestItems({ items }) {
  const [ready, setReady] = useState(false);
  useEffect(() => {
    const t = setTimeout(() => setReady(true), 200);
    return () => clearTimeout(t);
  }, []);

  if (!items?.length) return <div className="db-empty">No sales data yet</div>;
  const maxQ = Math.max(...items.map((i) => i.quantity), 1);

  return (
    <div className="items-list">
      {items.map((item, i) => (
        <div key={item.name} className="item-row">
          <div className="item-rank">{i + 1}</div>
          <div className="item-info">
            <div className="item-name">{item.name}</div>
            <div className="item-track">
              <div className="item-fill" style={{
                width: ready ? `${(item.quantity / maxQ) * 100}%` : '0%',
                transitionDelay: `${i * 60}ms`,
              }} />
            </div>
          </div>
          <div className="item-qty">{item.quantity} sold</div>
          <div className="item-rev">{fmt.usd0(item.revenue)}</div>
        </div>
      ))}
    </div>
  );
}

/* ─── TopCustomers ───────────────────────────── */

function TopCustomers({ customers }) {
  if (!customers?.length) return <div className="db-empty">No customer data yet</div>;

  return (
    <div className="cust-list">
      {customers.map((c) => (
        <div key={c.name} className="cust-row">
          <div className="cust-avatar">{c.name.charAt(0).toUpperCase()}</div>
          <div className="cust-info">
            <div className="cust-name">{c.name}</div>
            <div className="cust-meta">{c.order_count} order{c.order_count !== 1 ? 's' : ''}</div>
          </div>
          <div className="cust-right">
            <div className="cust-spent">{fmt.usd0(c.total_spent)}</div>
            <div className="cust-avg">{fmt.usd(c.total_spent / c.order_count)} avg</div>
          </div>
        </div>
      ))}
    </div>
  );
}

/* ─── Panel wrapper ──────────────────────────── */

function Panel({ title, badge, action, children }) {
  return (
    <div className="panel">
      <div className="panel-head">
        <div className="panel-title">
          {title}
          {badge != null && <span className="panel-badge">{badge}</span>}
        </div>
        {action}
      </div>
      <div className="panel-body">{children}</div>
    </div>
  );
}

/* ─── Dashboard ──────────────────────────────── */

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [trendMode, setTrendMode] = useState('revenue');
  const [updatedAt, setUpdatedAt] = useState(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getDashboardStats();
      setStats(data);
      setUpdatedAt(new Date());
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return (
    <>
      <div className="db-root">

        {loading && (
          <div className="db-load">
            <div className="db-spinner" />
            <span className="db-load-text">Loading dashboard…</span>
          </div>
        )}

        {!loading && error && (
          <div className="db-error">
            <span style={{ fontSize: 28 }}>✦</span>
            <p className="db-error-msg">{error}</p>
            <button className="db-retry" onClick={load}>Try again</button>
          </div>
        )}

        {!loading && !error && stats && (() => {
          const s = stats;

          return (
            <>
              {/* Header */}
              <div className="db-header">
                <div>
                  <h1 className="db-title">Sales <em>Dashboard</em></h1>
                  <div className="db-subtitle">
                    {updatedAt ? `Last updated · ${fmt.time(updatedAt)}` : 'Food Order Tracking'}
                  </div>
                </div>
                <button className="db-refresh" onClick={load}>
                  <RefreshIcon /> Refresh
                </button>
              </div>

              {/* KPI strip */}
              <div className="db-kpi">
                <KpiCard label="Total Revenue"   value={s.total_revenue}      isCurrency delay={0}   />
                <KpiCard label="Monthly Revenue" value={s.monthly_revenue}    isCurrency delay={60}  />
                <KpiCard label="Daily Revenue"   value={s.daily_revenue}      isCurrency delay={120} accent />
                <KpiCard label="Avg Order Value" value={s.average_order_value} isCurrency delay={180} />
                <KpiCard label="Total Orders"    value={s.total_orders}       delay={240}
                  sub={`${s.monthly_orders} this month`} />
                <KpiCard label="Today's Orders"  value={s.daily_orders}       delay={300}
                  sub={`${fmt.usd(s.daily_revenue)} revenue`} />
              </div>

              {/* Mid: trend + status */}
              <div className="db-mid">
                <Panel
                  title="Sales Trend"
                  badge="30 days"
                  action={
                    <div className="trend-toggle">
                      {['revenue', 'orders'].map((m) => (
                        <button key={m} className={`trend-btn${trendMode === m ? ' active' : ''}`}
                          onClick={() => setTrendMode(m)}>
                          {m}
                        </button>
                      ))}
                    </div>
                  }
                >
                  <TrendChart data={s.sales_trend} mode={trendMode} />
                </Panel>

                <Panel title="By Status">
                  <StatusBars statusMap={s.orders_by_status ?? {}} />
                </Panel>
              </div>

              {/* Bottom: items + customers */}
              <div className="db-bottom">
                <Panel title="Best Selling Items" badge={s.best_selling_items?.length ?? 0}>
                  <BestItems items={s.best_selling_items} />
                </Panel>

                <Panel title="Top Customers" badge={s.top_customers?.length ?? 0}>
                  <TopCustomers customers={s.top_customers} />
                </Panel>
              </div>
            </>
          );
        })()}

      </div>
    </>
  );
}

function RefreshIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="23 4 23 10 17 10"/>
      <polyline points="1 20 1 14 7 14"/>
      <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
    </svg>
  );
}