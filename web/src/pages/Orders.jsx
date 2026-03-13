import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';
import '../index.css';

/* ─── Status config ──────────────────────── */

const STATUSES = ['pending', 'preparing', 'ready', 'delivered', 'cancelled'];

const STATUS_CONFIG = {
  pending:   { label: 'Pending',   color: '#b45309', bg: '#fef3c7', border: '#fde68a' },
  preparing: { label: 'Preparing', color: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe' },
  ready:     { label: 'Ready',     color: '#6d28d9', bg: '#f5f3ff', border: '#ddd6fe' },
  delivered: { label: 'Delivered', color: '#065f46', bg: '#ecfdf5', border: '#a7f3d0' },
  cancelled: { label: 'Cancelled', color: '#991b1b', bg: '#fff1f2', border: '#fecdd3' },
};

const DEFAULT_EXPANDED = { pending: true, preparing: true, ready: true, delivered: false, cancelled: false };

/* ─── Helpers ────────────────────────────── */

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

function fmtDateTime(iso) {
  if (!iso) return null;
  return new Date(iso).toLocaleString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
    hour: 'numeric', minute: '2-digit',
  });
}

/* ─── StatusPill ─────────────────────────── */

function StatusPill({ status }) {
  const cfg = STATUS_CONFIG[status] ?? { label: status, color: '#666', bg: '#f3f4f6', border: '#e5e7eb' };
  return (
    <span className="ord-status-pill" style={{
      color: cfg.color, background: cfg.bg, borderColor: cfg.border,
    }}>
      <span className="ord-status-dot" style={{ background: cfg.color }} />
      {cfg.label}
    </span>
  );
}

/* ─── ConfirmDialog ──────────────────────── */

function ConfirmDialog({ orderId, onConfirm, onCancel }) {
  return (
    <div className="ord-overlay" onClick={onCancel}>
      <div className="ord-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="ord-dialog-title">Delete order?</div>
        <div className="ord-dialog-body">
          Order <strong>#{orderId}</strong> will be permanently removed. This cannot be undone.
        </div>
        <div className="ord-dialog-actions">
          <button className="btn-ghost" onClick={onCancel}>Cancel</button>
          <button className="btn-primary" style={{ background: 'var(--red)' }} onClick={onConfirm}>
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── OrderCard ──────────────────────────── */

function OrderCard({ order, onStatusChange, onDelete, delay }) {
  return (
    <div className="ord-card" style={{ animationDelay: `${delay}ms` }}>
      <div className="ord-card-top">
        <span className="ord-card-id">Order #{order.id}</span>
        <StatusPill status={order.status} />
      </div>

      <div className="ord-card-body">
        <div className={`ord-customer${!order.customer_name ? ' anonymous' : ''}`}>
          {order.customer_name || 'No customer'}
        </div>

        {[
          { label: 'Phone',   value: order.customer_phone },
          { label: 'Address', value: order.delivery_address },
          { label: 'Payment', value: order.payment_method === 'e-transfer' ? 'e-Transfer' : order.payment_method ? 'Cash' : null },
          { label: 'Notes',   value: order.notes },
        ].map(({ label, value }) => value ? (
          <div key={label} className="ord-detail">
            <span className="ord-detail-label">{label}</span>
            <span className="ord-detail-value">{value}</span>
          </div>
        ) : null)}

        {/* Items */}
        {order.order_items?.length > 0 && (
          <div className="ord-detail" style={{ alignItems: 'flex-start' }}>
            <span className="ord-detail-label">Items</span>
            <div className="ord-items">
              {order.order_items.map((item, i) => (
                <div key={i} className="ord-item-row">
                  <span className="ord-item-qty">{item.quantity}×</span>
                  <span>{item.item_name}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Timestamps */}
        {order.scheduled_date && (
          <div className="ord-detail">
            <span className="ord-detail-label">Sched.</span>
            <span className="ord-detail-value">{fmtDateTime(order.scheduled_date)}</span>
          </div>
        )}
        <div className="ord-detail">
          <span className="ord-detail-label">Created</span>
          <span className="ord-detail-value">{fmtDateTime(order.created_at)}</span>
        </div>
        {order.updated_at && order.updated_at !== order.created_at && (
          <div className="ord-detail">
            <span className="ord-detail-label">Updated</span>
            <span className="ord-detail-value">{fmtDateTime(order.updated_at)}</span>
          </div>
        )}

        <div className="ord-total">
          <span className="ord-total-label">Total</span>
          <span className="ord-total-value">{fmtUsd(order.total_amount)}</span>
        </div>
      </div>

      <div className="ord-card-footer">
        <Link to={`/orders/${order.id}/edit`} className="btn-primary">Edit</Link>
        <select
          className="ord-status-select"
          value={order.status}
          onChange={(e) => onStatusChange(order, e.target.value)}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>{STATUS_CONFIG[s].label}</option>
          ))}
        </select>
        <button className="btn-danger" onClick={() => onDelete(order)}>Delete</button>
      </div>
    </div>
  );
}

/* ─── StatusSection ──────────────────────── */

function StatusSection({ status, orders, expanded, onToggle, onStatusChange, onDelete }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <div className="ord-section">
      <button
        className="ord-section-head"
        style={{ borderLeftColor: cfg.color }}
        onClick={onToggle}
      >
        <ChevronIcon open={expanded} />
        <span className="ord-status-dot ord-section-dot" style={{ background: cfg.color }} />
        <span className="ord-section-label" style={{ color: cfg.color }}>{cfg.label}</span>
        <span className="ord-section-count">{orders.length}</span>
      </button>

      {expanded && (
        <div className="ord-section-body">
          {orders.length === 0 ? (
            <div className="ord-empty-section">No {cfg.label.toLowerCase()} orders</div>
          ) : (
            <div className="ord-grid">
              {orders.map((order, i) => (
                <OrderCard
                  key={order.id}
                  order={order}
                  delay={i * 35}
                  onStatusChange={onStatusChange}
                  onDelete={onDelete}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/* ─── Orders ─────────────────────────────── */

export default function Orders() {
  const [orders, setOrders]     = useState([]);
  const [loading, setLoading]   = useState(true);
  const [error, setError]       = useState(null);
  const [search, setSearch]     = useState('');
  const [statusFilter, setStatusFilter]   = useState('');
  const [paymentFilter, setPaymentFilter] = useState('');
  const [expanded, setExpanded] = useState(DEFAULT_EXPANDED);
  const [deleteTarget, setDeleteTarget]   = useState(null);
  const [toast, setToast]       = useState(null);

  /* ── Load ── */
  const load = useCallback(async () => {
    try {
      setLoading(true); setError(null);
      const data = await api.getOrders();
      setOrders(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  /* ── Toast ── */
  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  /* ── Status change ── */
  const handleStatusChange = async (order, newStatus) => {
    try {
      await api.updateOrder(order.id, {
        customer_id:      order.customer_id,
        delivery_address: order.delivery_address,
        status:           newStatus,
        payment_method:  order.payment_method,
        total_amount:     order.total_amount,
        notes:            order.notes,
        scheduled_date:   order.scheduled_date,
      });
      load();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  /* ── Delete ── */
  const confirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.deleteOrder(deleteTarget.id);
      setDeleteTarget(null);
      showToast('Order deleted');
      load();
    } catch (err) {
      setDeleteTarget(null);
      showToast(err.message, 'error');
    }
  };

  /* ── Filter ── */
  const filtered = orders.filter((o) => {
    const s = search.toLowerCase();
    const matchSearch  = !search || o.customer_name?.toLowerCase().includes(s) || o.delivery_address?.toLowerCase().includes(s);
    const matchStatus  = !statusFilter  || o.status          === statusFilter;
    const matchPayment = !paymentFilter || o.payment_method  === paymentFilter;
    return matchSearch && matchStatus && matchPayment;
  });

  const grouped = Object.fromEntries(
    STATUSES.map((s) => [s, filtered.filter((o) => o.status === s)])
  );

  const visibleStatuses = statusFilter ? [statusFilter] : STATUSES;

  /* ── Expand / collapse ── */
  const toggleSection  = (s) => setExpanded((p) => ({ ...p, [s]: !p[s] }));
  const expandAll      = () => setExpanded(Object.fromEntries(STATUSES.map((s) => [s, true])));
  const collapseAll    = () => setExpanded(Object.fromEntries(STATUSES.map((s) => [s, false])));

  /* ── Render ── */
  return (
    <>
      <div className="ord-root">

        {toast && <div className={`ord-toast ${toast.type}`}>{toast.msg}</div>}

        {deleteTarget && (
          <ConfirmDialog
            orderId={deleteTarget.id}
            onConfirm={confirmDelete}
            onCancel={() => setDeleteTarget(null)}
          />
        )}

        {/* Header */}
        <div className="ord-header">
          <div>
            <h1 className="ord-title">All <em>Orders</em></h1>
            <div className="ord-meta">
              {orders.length} order{orders.length !== 1 ? 's' : ''}
              {filtered.length !== orders.length && ` · ${filtered.length} shown`}
            </div>
          </div>
        </div>

        {loading && (
          <div className="ord-load">
            <div className="ord-spinner" />
            <span className="ord-load-text">Loading orders…</span>
          </div>
        )}

        {!loading && error && (
          <div className="ord-error">
            <p className="ord-error-msg">{error}</p>
            <button className="ord-retry" onClick={load}>Try again</button>
          </div>
        )}

        {!loading && !error && (
          <>
            {/* Filters */}
            <div className="ord-filters">
              <div className="ord-search-wrap">
                <span className="ord-search-icon"><SearchIcon /></span>
                <input
                  className="ord-search"
                  type="text"
                  placeholder="Search customer or address…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <select className="ord-select" value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}>
                <option value="">All statuses</option>
                {STATUSES.map((s) => (
                  <option key={s} value={s}>{STATUS_CONFIG[s].label}</option>
                ))}
              </select>
              <select className="ord-select" value={paymentFilter}
                onChange={(e) => setPaymentFilter(e.target.value)}>
                <option value="">All payments</option>
                <option value="cash">Cash</option>
                <option value="e-transfer">e-Transfer</option>
              </select>
            </div>

            {/* Section controls */}
            <div className="ord-controls">
              <button className="btn-ghost" onClick={expandAll}>Expand all</button>
              <button className="btn-ghost" onClick={collapseAll}>Collapse all</button>
            </div>

            {/* Empty state */}
            {orders.length === 0 && (
              <div className="ord-empty">
                <span style={{ fontSize: 32, opacity: 0.25 }}>✦</span>
                <div className="ord-empty-title">No orders yet</div>
                <div className="ord-empty-sub">Orders placed from the Menu page will appear here.</div>
              </div>
            )}

            {orders.length > 0 && filtered.length === 0 && (
              <div className="ord-empty">
                <div className="ord-empty-title">No results</div>
                <div className="ord-empty-sub">Try adjusting your filters.</div>
              </div>
            )}

            {/* Status sections */}
            {filtered.length > 0 && visibleStatuses.map((status) => (
              <StatusSection
                key={status}
                status={status}
                orders={grouped[status]}
                expanded={expanded[status]}
                onToggle={() => toggleSection(status)}
                onStatusChange={handleStatusChange}
                onDelete={(order) => setDeleteTarget(order)}
              />
            ))}
          </>
        )}

      </div>
    </>
  );
}

/* ─── Icons ──────────────────────────────── */

function ChevronIcon({ open }) {
  return (
    <svg className={`ord-section-chevron${open ? ' open' : ''}`}
      width="14" height="14" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="9 18 15 12 9 6" />
    </svg>
  );
}

function SearchIcon() {
  return (
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
    </svg>
  );
}