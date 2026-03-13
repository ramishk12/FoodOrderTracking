import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';
import '../index.css';

/* ─── Constants ─────────────────────────────── */

const STATUS_COLORS = {
  pending:   '#c47c2b',
  preparing: '#2b5fa0',
  ready:     '#6b3fa0',
  delivered: '#2d7a4f',
  cancelled: '#a02b2b',
};

const GROUP_ORDER = ['Overdue', 'Today', 'Tomorrow', 'This Week'];

/* ─── Helpers ────────────────────────────── */

function diffDays(scheduledDate) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const d = new Date(scheduledDate);
  d.setHours(0, 0, 0, 0);
  return Math.floor((d - today) / 86400000);
}

function groupLabel(scheduledDate) {
  const diff = diffDays(scheduledDate);
  if (diff < 0)  return 'Overdue';
  if (diff === 0) return 'Today';
  if (diff === 1) return 'Tomorrow';
  if (diff <= 7)  return 'This Week';
  const d = new Date(scheduledDate);
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
}

function dateStatus(scheduledDate) {
  const diff = diffDays(scheduledDate);
  if (diff < 0)  return 'overdue';
  if (diff === 0) return 'today';
  return 'normal';
}

function formatTime(scheduledDate) {
  const date = new Date(scheduledDate);
  const diff = diffDays(scheduledDate);
  const time = date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
  if (diff === 0) return `Today · ${time}`;
  if (diff === 1) return `Tomorrow · ${time}`;
  if (diff < 0)   return `${Math.abs(diff)}d overdue · ${time}`;
  return `${date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} · ${time}`;
}

function groupOrders(orders) {
  const groups = {};
  orders.forEach((order) => {
    if (!order.scheduled_date) return;
    const label = groupLabel(order.scheduled_date);
    if (!groups[label]) groups[label] = [];
    groups[label].push(order);
  });

  // Sort: predefined order first, then chronological for any extra labels
  const keys = Object.keys(groups).sort((a, b) => {
    const ai = GROUP_ORDER.indexOf(a);
    const bi = GROUP_ORDER.indexOf(b);
    if (ai !== -1 && bi !== -1) return ai - bi;
    if (ai !== -1) return -1;
    if (bi !== -1) return 1;
    return 0;
  });

  return keys.map((label) => ({ label, orders: groups[label] }));
}

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

/* ─── OrderCard ──────────────────────────── */

function OrderCard({ order }) {
  const status = dateStatus(order.scheduled_date);
  const color = STATUS_COLORS[order.status] || '#8a7060';

  return (
    <div className={`sch-card ${status}`}>
      <div className="sch-card-bar" />
      <div className="sch-card-body">
        <div className="sch-card-top">
          <span className="sch-order-id">#{order.id}</span>
          <span className="sch-order-time">{formatTime(order.scheduled_date)}</span>
        </div>

        <div>
          <span className="sch-status-pill" style={{ color }}>
            <span className="sch-status-dot" />
            {order.status}
          </span>
        </div>

        <div className="sch-customer">{order.customer_name || 'No customer'}</div>
        <div className="sch-items-list">
          {order.order_items?.length
            ? order.order_items.map((item) => (
                <div key={item.id ?? item.item_name}>
                  {item.quantity}× {item.item_name}
                </div>
              ))
            : <div>No items</div>}
        </div>
      </div>

      <div className="sch-card-footer">
        <span className="sch-total">{fmtUsd(order.total_amount)}</span>
        <Link to={`/orders/${order.id}/edit`} className="sch-edit-btn">
          Edit →
        </Link>
      </div>
    </div>
  );
}

/* ─── Schedule ───────────────────────────── */

export default function Schedule() {
  const [orders, setOrders]   = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState(null);
  const [days, setDays]       = useState(7);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getScheduledOrders(days);
      setOrders(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [days]);

  useEffect(() => { load(); }, [load]);

  const groups  = groupOrders(orders);
  const overdue = orders.filter((o) => o.scheduled_date && diffDays(o.scheduled_date) < 0).length;
  const today   = orders.filter((o) => o.scheduled_date && diffDays(o.scheduled_date) === 0).length;

  return (
    <>
      <div className="sch-root">

        {/* Header */}
        <div className="sch-header">
          <div>
            <h1 className="sch-title">Order <em>Schedule</em></h1>
            <div className="sch-meta">
              Showing next {days} day{days !== 1 ? 's' : ''} · {orders.length} order{orders.length !== 1 ? 's' : ''}
            </div>
          </div>
          <div className="sch-controls">
            <div className="sch-window">
              {WINDOWS.map((w) => (
                <button
                  key={w}
                  className={`sch-window-btn${days === w ? ' active' : ''}`}
                  onClick={() => setDays(w)}
                >
                  {w}d
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Loading */}
        {loading && (
          <div className="sch-load">
            <div className="sch-spinner" />
            <span className="sch-load-text">Loading schedule…</span>
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="sch-error">
            <p className="sch-error-msg">{error}</p>
            <button className="sch-retry" onClick={load}>Try again</button>
          </div>
        )}

        {/* Content */}
        {!loading && !error && (
          <>
            {/* Summary strip — only when there are orders */}
            {orders.length > 0 && (
              <div className="sch-summary">
                <div className="sch-summary-item">
                  <span className="sch-summary-label">Total</span>
                  <span className="sch-summary-value">{orders.length}</span>
                </div>
                {today > 0 && (
                  <>
                    <div className="sch-summary-div" />
                    <div className="sch-summary-item">
                      <span className="sch-summary-label" style={{ color: 'var(--amber)' }}>Today</span>
                      <span className="sch-summary-value" style={{ color: 'var(--amber)' }}>{today}</span>
                    </div>
                  </>
                )}
                {overdue > 0 && (
                  <>
                    <div className="sch-summary-div" />
                    <div className="sch-summary-item">
                      <span className="sch-summary-label" style={{ color: 'var(--red)' }}>Overdue</span>
                      <span className="sch-summary-value" style={{ color: 'var(--red)' }}>{overdue}</span>
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Empty state */}
            {orders.length === 0 && (
              <div className="sch-empty">
                <span style={{ fontSize: 32, opacity: 0.3 }}>✦</span>
                <div className="sch-empty-title">Nothing scheduled</div>
                <div className="sch-empty-sub">
                  No orders in the next {days} days. Create an order and set a delivery date.
                </div>
              </div>
            )}

            {/* Groups */}
            <div className="sch-groups">
              {groups.map(({ label, orders: groupOrders }, i) => {
                const isOverdue = label === 'Overdue';
                const isToday   = label === 'Today';
                return (
                  <div
                    key={label}
                    className={`sch-group${isOverdue ? ' overdue-group' : ''}${isToday ? ' today-group' : ''}`}
                    style={{ animationDelay: `${i * 80}ms` }}
                  >
                    <div className="sch-group-head">
                      <h2 className={`sch-group-label${isOverdue ? ' overdue' : ''}${isToday ? ' today' : ''}`}>
                        {label}
                      </h2>
                      <span className="sch-group-count">
                        {groupOrders.length} order{groupOrders.length !== 1 ? 's' : ''}
                      </span>
                    </div>
                    <div className="sch-cards">
                      {groupOrders.map((order) => (
                        <OrderCard key={order.id} order={order} />
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </>
        )}

      </div>
    </>
  );
}