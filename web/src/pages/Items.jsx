import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import '../index.css';

/* ─── Constants ──────────────────────────── */

const EMPTY_ITEM_FORM = { name: '', description: '', price: '', category: '' };
const EMPTY_ORDER_FORM = {
  customer_id: '', delivery_address: '', notes: '',
  payment_method: 'cash', scheduled_date: '',
};

/* ─── Helpers ────────────────────────────── */

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

function fmtDate(iso) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  });
}

function localDatetimeToUtcIso(localStr) {
  if (!localStr) return null;
  const d = new Date(localStr);
  return d.toISOString();
}

/* ─── Toast ──────────────────────────────── */

function Toast({ message, type }) {
  return <div className={`itm-toast ${type}`}>{message}</div>;
}

/* ─── ConfirmDialog ──────────────────────── */

function ConfirmDialog({ title, body, onConfirm, onCancel }) {
  return (
    <div className="itm-overlay" onClick={onCancel}>
      <div className="itm-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="itm-dialog-title">{title}</div>
        <div className="itm-dialog-body">{body}</div>
        <div className="itm-dialog-actions">
          <button className="btn-ghost" onClick={onCancel}>Cancel</button>
          <button className="btn-primary" style={{ background: 'var(--red)' }} onClick={onConfirm}>
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── ItemCard ───────────────────────────── */

function ItemCard({ item, qty, onQtyChange, onEdit, onDelete, onActivate, delay }) {
  const selected = qty > 0;
  const unavailable = !item.available;
  return (
    <div
      className={`itm-card${selected ? ' selected' : ''}${unavailable ? ' unavailable' : ''}`}
      style={{ animationDelay: `${delay}ms` }}
    >
      <div className="itm-card-body">
        <div className="itm-card-top">
          <div className="itm-card-name">{item.name}</div>
          <div className="itm-card-price">{fmtUsd(item.price)}</div>
        </div>
        {item.description && (
          <div className="itm-card-desc">{item.description}</div>
        )}
        {unavailable ? (
          <div className="itm-unavailable-badge">Unavailable</div>
        ) : (
          <div className="itm-qty-row">
            <button className="itm-qty-btn"
              onClick={() => onQtyChange(item.id, qty - 1)}>−</button>
            <input
              className="itm-qty-input"
              type="number" min="0"
              value={qty}
              onChange={(e) => onQtyChange(item.id, e.target.value)}
            />
            <button className="itm-qty-btn"
              onClick={() => onQtyChange(item.id, qty + 1)}>+</button>
          </div>
        )}
      </div>
      <div className="itm-card-footer">
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--muted)' }}>
          {unavailable ? '\u00a0' : (selected ? `${fmtUsd(item.price * qty)} selected` : '\u00a0')}
        </span>
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn-ghost" style={{ padding: '5px 11px', fontSize: 11 }}
            onClick={() => onEdit(item)}>Edit</button>
          {unavailable ? (
            <button className="btn-primary" style={{ background: 'var(--green)' }} onClick={() => onActivate(item)}>
              Activate
            </button>
          ) : (
            <button className="btn-danger" onClick={() => onDelete(item)}>Delete</button>
          )}
        </div>
      </div>
    </div>
  );
}

/* ─── Items ──────────────────────────────── */

export default function Items() {
  const [items, setItems]         = useState([]);
  const [customers, setCustomers] = useState([]);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState(null);
  const [quantities, setQuantities] = useState({});

  const [itemFormOpen, setItemFormOpen]   = useState(false);
  const [editingItem, setEditingItem]     = useState(null);
  const [itemForm, setItemForm]           = useState(EMPTY_ITEM_FORM);
  const [itemFormError, setItemFormError] = useState(null);
  const [itemSubmitting, setItemSubmitting] = useState(false);

  const [orderFormOpen, setOrderFormOpen]   = useState(false);
  const [orderForm, setOrderForm]           = useState(EMPTY_ORDER_FORM);
  const [orderFormError, setOrderFormError] = useState(null);
  const [orderSubmitting, setOrderSubmitting] = useState(false);
  const [orderHistory, setOrderHistory]     = useState([]);
  const [historyLoading, setHistoryLoading] = useState(false);

  const [search, setSearch]             = useState('');
  const [catFilter, setCatFilter]       = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [toast, setToast]               = useState(null);

  /* ── Derived ── */
  const availableItems = items.filter((i) => i.available);
  const unavailableItems = items.filter((i) => !i.available);
  const categories = [...new Set(availableItems.map((i) => i.category).filter(Boolean))].sort();

  const filteredItems = availableItems.filter((item) => {
    const s = search.toLowerCase();
    const matchSearch = !search ||
      item.name?.toLowerCase().includes(s) ||
      item.description?.toLowerCase().includes(s);
    const matchCat = !catFilter || item.category === catFilter;
    return matchSearch && matchCat;
  });

  const groupedItems = filteredItems.reduce((acc, item) => {
    const cat = item.category || 'Uncategorised';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(item);
    return acc;
  }, {});

  const selectedItems = availableItems.filter((i) => (quantities[i.id] || 0) > 0);
  const totalAmount = selectedItems.reduce(
    (sum, i) => sum + i.price * (quantities[i.id] || 0), 0
  );
  const totalQty = selectedItems.reduce((sum, i) => sum + (quantities[i.id] || 0), 0);

  /* ── Helpers ── */
  const showToast = (message, type = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  const resetQty = () => {
    setQuantities((prev) => {
      const r = { ...prev };
      Object.keys(r).forEach((k) => (r[k] = 0));
      return r;
    });
  };

  /* ── Load ── */
  const load = useCallback(async () => {
    try {
      setLoading(true); setError(null);
      const [itemsData, customersData] = await Promise.all([
        api.getItems(), api.getCustomers(),
      ]);
      setItems(itemsData || []);
      setCustomers(customersData || []);
      setQuantities((prev) => {
        const next = {};
        (itemsData || []).forEach((i) => { next[i.id] = prev[i.id] || 0; });
        return next;
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  /* ── Quantity ── */
  const handleQty = (itemId, raw) => {
    const qty = Math.max(0, parseInt(raw) || 0);
    setQuantities((prev) => ({ ...prev, [itemId]: qty }));
  };

  /* ── Item form ── */
  const openCreate = () => {
    setItemForm(EMPTY_ITEM_FORM);
    setEditingItem(null);
    setItemFormError(null);
    setItemFormOpen(true);
    setOrderFormOpen(false);
  };

  const openEdit = (item) => {
    setItemForm({
      name: item.name, description: item.description || '',
      price: String(item.price), category: item.category || '',
    });
    setEditingItem(item);
    setItemFormError(null);
    setItemFormOpen(true);
    setOrderFormOpen(false);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const closeItemForm = () => {
    setItemFormOpen(false); setEditingItem(null);
    setItemForm(EMPTY_ITEM_FORM); setItemFormError(null);
  };

  const handleItemSubmit = async (e) => {
    e.preventDefault();
    setItemSubmitting(true); setItemFormError(null);
    try {
      const payload = { ...itemForm, price: parseFloat(itemForm.price), available: true };
      if (editingItem) {
        await api.updateItem(editingItem.id, payload);
        showToast('Item updated');
      } else {
        await api.createItem(payload);
        showToast('Item created');
      }
      closeItemForm(); load();
    } catch (err) {
      setItemFormError(err.message);
    } finally {
      setItemSubmitting(false);
    }
  };

  const itemField = (key) => ({
    value: itemForm[key],
    onChange: (e) => setItemForm((p) => ({ ...p, [key]: e.target.value })),
    className: 'itm-input',
  });

  /* ── Delete ── */
  const confirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.deleteItem(deleteTarget.id);
      setDeleteTarget(null);
      showToast('Item deleted');
      load();
    } catch (err) {
      setDeleteTarget(null);
      showToast(err.message, 'error');
    }
  };

  /* ── Activate ── */
  const handleActivate = async (item) => {
    try {
      await api.activateItem(item.id);
      showToast(`${item.name} is now available`);
      load();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  /* ── Order form ── */
  const openOrderForm = () => {
    setOrderFormOpen(true);
    setItemFormOpen(false);
    setOrderFormError(null);
    setOrderHistory([]);
  };

  const closeOrderForm = () => {
    setOrderFormOpen(false);
    setOrderForm(EMPTY_ORDER_FORM);
    setOrderFormError(null);
    setOrderHistory([]);
    resetQty();
  };

  const handleCustomerChange = async (customerId) => {
    const customer = customers.find((c) => c.id === parseInt(customerId));
    setOrderForm((p) => ({
      ...p,
      customer_id: customerId,
      delivery_address: customer?.address || p.delivery_address,
    }));
    setOrderHistory([]);
    if (!customerId) return;
    try {
      setHistoryLoading(true);
      const orders = await api.getOrdersByCustomer(customerId);
      setOrderHistory(orders || []);
    } catch {
      setOrderHistory([]);
    } finally {
      setHistoryLoading(false);
    }
  };

  const handleOrderSubmit = async (e) => {
    e.preventDefault();
    if (selectedItems.length === 0) {
      setOrderFormError('Select at least one item from the menu below.');
      return;
    }
    setOrderSubmitting(true); setOrderFormError(null);
    try {
      await api.createOrder({
        customer_id: parseInt(orderForm.customer_id) || null,
        delivery_address: orderForm.delivery_address,
        notes: orderForm.notes,
        payment_method: orderForm.payment_method,
        scheduled_date: localDatetimeToUtcIso(orderForm.scheduled_date),
        items: selectedItems.map((i) => ({ item_id: i.id, quantity: quantities[i.id] })),
      });
      showToast('Order placed successfully');
      closeOrderForm();
      resetQty();
    } catch (err) {
      setOrderFormError(err.message);
    } finally {
      setOrderSubmitting(false);
    }
  };

  const ordField = (key) => ({
    value: orderForm[key],
    onChange: (e) => setOrderForm((p) => ({ ...p, [key]: e.target.value })),
  });

  /* ── Render ── */
  return (
    <>
      <div className="itm-root">

        {toast && <Toast message={toast.message} type={toast.type} />}

        {deleteTarget && (
          <ConfirmDialog
            title="Delete item?"
            body={<><strong>{deleteTarget.name}</strong> will be permanently removed from the menu.</>}
            onConfirm={confirmDelete}
            onCancel={() => setDeleteTarget(null)}
          />
        )}

        {/* Header */}
        <div className="itm-header">
          <div>
            <h1 className="itm-title">The <em>Menu</em></h1>
            <div className="itm-meta">
              {availableItems.length} available item{availableItems.length !== 1 ? 's' : ''}
              {categories.length > 0 && ` · ${categories.length} categories`}
              {unavailableItems.length > 0 && ` · ${unavailableItems.length} unavailable`}
            </div>
          </div>
          <div className="itm-header-actions">
            {itemFormOpen
              ? <button className="btn-ghost" onClick={closeItemForm}>✕ Cancel</button>
              : <button className="btn-primary" onClick={openCreate}>+ Add Item</button>
            }
          </div>
        </div>

        {/* Item form */}
        <div className={`itm-panel-wrap ${itemFormOpen ? 'open' : 'closed'}`}>
          <form className="itm-panel" onSubmit={handleItemSubmit}>
            <div className="itm-panel-head">
              <div className="itm-panel-title">
                {editingItem ? <>Edit <em>item</em></> : <>New <em>item</em></>}
              </div>
            </div>
            {itemFormError && <div className="itm-form-error">{itemFormError}</div>}
            <div className="itm-form-body">
              <div className="itm-field full">
                <label className="itm-label">Name *</label>
                <input {...itemField('name')} placeholder="e.g. Margherita Pizza" required />
              </div>
              <div className="itm-field full">
                <label className="itm-label">Description</label>
                <input {...itemField('description')} placeholder="Short description" />
              </div>
              <div className="itm-field">
                <label className="itm-label">Price *</label>
                <input {...itemField('price')} type="number" step="0.01" min="0"
                  placeholder="0.00" required />
              </div>
              <div className="itm-field">
                <label className="itm-label">Category *</label>
                <input {...itemField('category')} placeholder="e.g. mains, drinks…" required />
              </div>
            </div>
            <div className="itm-form-actions">
              <button type="submit" className="btn-primary" disabled={itemSubmitting}>
                {itemSubmitting ? 'Saving…' : editingItem ? 'Update item' : 'Create item'}
              </button>
              <button type="button" className="btn-ghost" onClick={closeItemForm}>Cancel</button>
            </div>
          </form>
        </div>

        {/* Order form */}
        <div className={`itm-panel-wrap ${orderFormOpen ? 'open' : 'closed'}`}>
          <form className="itm-panel" onSubmit={handleOrderSubmit}>
            <div className="itm-panel-head">
              <div className="itm-panel-title">New <em>order</em></div>
              <button type="button" className="btn-ghost" onClick={closeOrderForm}
                style={{ padding: '5px 12px', fontSize: 11 }}>✕</button>
            </div>
            {orderFormError && <div className="itm-form-error">{orderFormError}</div>}
            <div className="ord-body">
              {/* Left col — form fields */}
              <div className="ord-col">
                <div className="itm-field">
                  <label className="itm-label">Customer</label>
                  <select
                    className="itm-select"
                    value={orderForm.customer_id}
                    onChange={(e) => handleCustomerChange(e.target.value)}
                  >
                    <option value="">No customer</option>
                    {customers.map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
                <div className="itm-field">
                  <label className="itm-label">Delivery Address</label>
                  <input className="itm-input" {...ordField('delivery_address')}
                    placeholder="Street address" />
                </div>
                <div className="itm-field">
                  <label className="itm-label">Payment Method</label>
                  <select className="itm-select" value={orderForm.payment_method}
                    onChange={(e) => setOrderForm((p) => ({ ...p, payment_method: e.target.value }))}>
                    <option value="cash">Cash</option>
                    <option value="e-transfer">e-Transfer</option>
                  </select>
                </div>
                <div className="itm-field">
                  <label className="itm-label">Scheduled Date</label>
                  <input className="itm-input" type="datetime-local"
                    value={orderForm.scheduled_date}
                    onChange={(e) => setOrderForm((p) => ({ ...p, scheduled_date: e.target.value }))} />
                </div>
                <div className="itm-field">
                  <label className="itm-label">Notes</label>
                  <input className="itm-input" {...ordField('notes')}
                    placeholder="Any special instructions" />
                </div>
              </div>

              {/* Right col — summary + history */}
              <div className="ord-col">
                {/* Order summary */}
                <div className="ord-summary">
                  <div className="ord-summary-title">Order summary</div>
                  {selectedItems.length === 0
                    ? <div className="ord-summary-empty">Select items from the menu below</div>
                    : <>
                        {selectedItems.map((item) => (
                          <div key={item.id} className="ord-line">
                            <span className="ord-line-name">{item.name}</span>
                            <span className="ord-line-qty">×{quantities[item.id]}</span>
                            <span className="ord-line-price">
                              {fmtUsd(item.price * quantities[item.id])}
                            </span>
                          </div>
                        ))}
                        <div className="ord-total">
                          <span className="ord-total-label">Total</span>
                          <span className="ord-total-value">{fmtUsd(totalAmount)}</span>
                        </div>
                      </>
                  }
                </div>

                {/* Customer order history */}
                {orderForm.customer_id && (
                  <div className="ord-history">
                    <div className="ord-history-title">Order history</div>
                    {historyLoading
                      ? <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--muted)' }}>
                          Loading…
                        </div>
                      : orderHistory.length === 0
                        ? <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--rule)', fontStyle: 'italic' }}>
                            No previous orders
                          </div>
                        : orderHistory.map((order) => (
                            <div key={order.id} className="ord-hist-row">
                              <div className="ord-hist-top">
                                <span className="ord-hist-id">#{order.id}</span>
                                <span className="ord-hist-date">{fmtDate(order.created_at)}</span>
                              </div>
                              <div className="ord-hist-items">
                                {order.order_items?.length > 0 ? (
                                  <div>
                                    {order.order_items.map((i) => (
                                      <div key={i.item_id}>{i.quantity}× {i.item_name}</div>
                                    ))}
                                  </div>
                                ) : 'No items'}
                              </div>
                              <div className="ord-hist-footer">
                                <span className="ord-hist-total">{fmtUsd(order.total_amount)}</span>
                                <span className="ord-hist-status">{order.status}</span>
                              </div>
                            </div>
                          ))
                    }
                  </div>
                )}
              </div>
            </div>

            <div className="itm-form-actions">
              <button type="submit" className="btn-order" disabled={orderSubmitting || selectedItems.length === 0}>
                {orderSubmitting ? 'Placing…' : `Place order · ${fmtUsd(totalAmount)}`}
              </button>
              <button type="button" className="btn-ghost" onClick={closeOrderForm}>Cancel</button>
            </div>
          </form>
        </div>

        {/* Loading */}
        {loading && (
          <div className="itm-load">
            <div className="itm-spinner" />
            <span className="itm-load-text">Loading menu…</span>
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="itm-error">
            <p className="itm-error-msg">{error}</p>
            <button className="itm-retry" onClick={load}>Try again</button>
          </div>
        )}

        {/* Content */}
        {!loading && !error && (
          <>
            {/* Filters */}
            <div className="itm-filters">
              <div className="itm-search-wrap">
                <span className="itm-search-icon"><SearchIcon /></span>
                <input
                  className="itm-search"
                  type="text"
                  placeholder="Search menu…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <div className="itm-cat-pills">
                <button
                  className={`itm-cat-pill${catFilter === '' ? ' active' : ''}`}
                  onClick={() => setCatFilter('')}
                >All</button>
                {categories.map((cat) => (
                  <button
                    key={cat}
                    className={`itm-cat-pill${catFilter === cat ? ' active' : ''}`}
                    onClick={() => setCatFilter(cat === catFilter ? '' : cat)}
                  >{cat}</button>
                ))}
              </div>
            </div>

            {/* Empty states */}
            {availableItems.length === 0 && unavailableItems.length === 0 && (
              <div className="itm-empty">
                <span style={{ fontSize: 32, opacity: 0.25 }}>✦</span>
                <div className="itm-empty-title">Menu is empty</div>
                <div className="itm-empty-sub">Add your first item to get started.</div>
              </div>
            )}

            {availableItems.length > 0 && filteredItems.length === 0 && (
              <div className="itm-empty">
                <div className="itm-empty-title">No items found</div>
                <div className="itm-empty-sub">Try a different search or category.</div>
              </div>
            )}

            {/* Category sections */}
            {Object.entries(groupedItems).map(([category, catItems], gi) => (
              <div
                key={category}
                className="itm-category"
                style={{ animationDelay: `${gi * 60}ms` }}
              >
                <div className="itm-cat-head">
                  <h2 className="itm-cat-name">{category}</h2>
                  <span className="itm-cat-count">
                    {catItems.length} item{catItems.length !== 1 ? 's' : ''}
                  </span>
                </div>
                <div className="itm-grid">
                  {catItems.map((item, ii) => (
                    <ItemCard
                      key={item.id}
                      item={item}
                      qty={quantities[item.id] || 0}
                      onQtyChange={handleQty}
                      onEdit={openEdit}
                      onDelete={(i) => setDeleteTarget(i)}
                      onActivate={handleActivate}
                      delay={gi * 60 + ii * 40}
                    />
                  ))}
                </div>
              </div>
            ))}

            {/* Unavailable items section */}
            {unavailableItems.length > 0 && (
              <div className="itm-unavailable-section">
                <div className="itm-unavailable-head">
                  <h2 className="itm-unavailable-title">Unavailable Items</h2>
                  <span className="itm-unavailable-count">
                    {unavailableItems.length} item{unavailableItems.length !== 1 ? 's' : ''}
                  </span>
                </div>
                <div className="itm-grid">
                  {unavailableItems.map((item, ii) => (
                    <ItemCard
                      key={item.id}
                      item={item}
                      qty={0}
                      onQtyChange={() => {}}
                      onEdit={openEdit}
                      onDelete={(i) => setDeleteTarget(i)}
                      onActivate={handleActivate}
                      delay={ii * 40}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Sticky order bar — visible when items selected */}
            {totalQty > 0 && (
              <div className="itm-order-bar">
                <div className="itm-order-bar-left">
                  <span className="itm-order-bar-label">
                    {totalQty} item{totalQty !== 1 ? 's' : ''} selected
                  </span>
                  <span className="itm-order-bar-items">
                    {selectedItems.map((i) => `${quantities[i.id]}× ${i.name}`).join(' · ')}
                  </span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
                  <span className="itm-order-bar-total">{fmtUsd(totalAmount)}</span>
                  <div className="itm-order-bar-actions">
                    <button className="itm-order-bar-clear" onClick={resetQty}>Clear</button>
                    <button className="itm-order-bar-btn" onClick={openOrderForm}>
                      Place Order →
                    </button>
                  </div>
                </div>
              </div>
            )}
          </>
        )}

      </div>
    </>
  );
}

function SearchIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
    </svg>
  );
}