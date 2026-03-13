import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import '../index.css';

/* ─── Helpers ────────────────────────────── */

const EMPTY_CUST = { name: '', phone: '', email: '', address: '' };

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

function fmtDate(iso) {
  if (!iso) return 'N/A';
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  });
}

function utcToLocalInput(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function localInputToUtcIso(local) {
  if (!local) return null;
  return new Date(local).toISOString();
}

/* ─── OrderEdit ──────────────────────────── */

export default function OrderEdit() {
  const { id }    = useParams();
  const navigate  = useNavigate();

  const [order, setOrder]       = useState(null);
  const [customers, setCustomers] = useState([]);
  const [items, setItems]       = useState([]);
  const [loading, setLoading]   = useState(true);
  const [error, setError]       = useState(null);
  const [saving, setSaving]     = useState(false);
  const [saveError, setSaveError] = useState(null);
  const [toast, setToast]       = useState(null);

  const [formData, setFormData] = useState({
    customer_id: '', delivery_address: '',
    notes: '', payment_method: 'cash', scheduled_date: '',
  });
  const [orderItems, setOrderItems] = useState({});

  const [showNewCust, setShowNewCust] = useState(false);
  const [custForm, setCustForm]       = useState(EMPTY_CUST);
  const [custSaving, setCustSaving]   = useState(false);
  const [custError, setCustError]     = useState(null);

  const [orderHistory, setOrderHistory]       = useState([]);
  const [historyVisible, setHistoryVisible]   = useState(false);

  /* ── Load ── */
  const load = useCallback(async () => {
    try {
      setLoading(true); setError(null);
      const [orderData, customersData, itemsData] = await Promise.all([
        api.getOrder(id), api.getCustomers(), api.getItems(),
      ]);
      setOrder(orderData);
      setCustomers(customersData || []);
      setItems(itemsData || []);

      setFormData({
        customer_id:      orderData.customer_id ? String(orderData.customer_id) : '',
        delivery_address: orderData.delivery_address || '',
        notes:            orderData.notes || '',
        payment_method:   orderData.payment_method || 'cash',
        scheduled_date:   utcToLocalInput(orderData.scheduled_date),
      });

      const qty = {};
      (orderData.order_items || []).forEach((oi) => { qty[oi.item_id] = oi.quantity; });
      setOrderItems(qty);

      if (orderData.customer_id) {
        fetchHistory(orderData.customer_id, parseInt(id));
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  /* ── Helpers ── */
  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const fetchHistory = async (customerId, excludeId) => {
    try {
      const orders = await api.getOrdersByCustomer(customerId);
      setOrderHistory((orders || []).filter((o) => o.id !== excludeId));
      setHistoryVisible(true);
    } catch { setOrderHistory([]); }
  };

  /* ── Quantity ── */
  const setQty = (itemId, raw) => {
    const qty = Math.max(0, parseInt(raw) || 0);
    setOrderItems((p) => ({ ...p, [itemId]: qty }));
  };

  /* ── Derived ── */
  const selectedItems = Object.entries(orderItems)
    .filter(([, q]) => q > 0)
    .map(([itemId, quantity]) => {
      const item = items.find((i) => i.id === parseInt(itemId));
      return { item_id: parseInt(itemId), quantity, price: item?.price || 0, name: item?.name || 'Unknown' };
    });

  const total = selectedItems.reduce((s, i) => s + i.price * i.quantity, 0);

  const groupedItems = items.reduce((acc, item) => {
    const cat = item.category || 'Uncategorised';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(item);
    return acc;
  }, {});

  /* ── Customer change ── */
  const handleCustomerChange = async (customerId) => {
    const customer = customers.find((c) => c.id === parseInt(customerId));
    setFormData((p) => ({
      ...p, customer_id: customerId,
      delivery_address: customer?.address || p.delivery_address,
    }));
    setOrderHistory([]); setHistoryVisible(false);
    if (customerId) fetchHistory(parseInt(customerId), parseInt(id));
  };

  /* ── New customer ── */
  const handleNewCustomer = async (e) => {
    e.preventDefault();
    setCustSaving(true); setCustError(null);
    try {
      const newCust = await api.createCustomer(custForm);
      const updated = await api.getCustomers();
      setCustomers(updated || []);
      setShowNewCust(false);
      setCustForm(EMPTY_CUST);
      handleCustomerChange(String(newCust.id));
    } catch (err) {
      setCustError(err.message);
    } finally {
      setCustSaving(false);
    }
  };

  /* ── Save ── */
  const handleSave = async () => {
    if (selectedItems.length === 0) {
      setSaveError("Add at least one item to the order before saving.");
      return;
    }
    setSaving(true); setSaveError(null);
    try {
      await api.updateOrder(parseInt(id), {
        customer_id:      parseInt(formData.customer_id) || null,
        delivery_address: formData.delivery_address,
        status:           order?.status || 'pending',
        notes:            formData.notes,
        payment_method:   formData.payment_method,
        total_amount:     total,
        scheduled_date:   localInputToUtcIso(formData.scheduled_date),
        items:            selectedItems,
      });
      showToast('Order saved');
      setTimeout(() => navigate('/orders'), 900);
    } catch (err) {
      setSaveError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const field = (key) => ({
    value: formData[key],
    onChange: (e) => setFormData((p) => ({ ...p, [key]: e.target.value })),
    className: 'oe-input',
  });

  const custField = (key) => ({
    value: custForm[key],
    onChange: (e) => setCustForm((p) => ({ ...p, [key]: e.target.value })),
    className: 'oe-input',
  });

  /* ── Render ── */
  if (loading) return (
    <>
      <div className="oe-root">
        <div className="oe-load">
          <div className="oe-spinner" />
          <span className="oe-load-text">Loading order…</span>
        </div>
      </div>
    </>
  );

  if (error) return (
    <>
      <div className="oe-root">
        <div className="oe-error-page">
          <p className="oe-error-msg">{error}</p>
          <button className="oe-retry" onClick={load}>Try again</button>
        </div>
      </div>
    </>
  );

  return (
    <>
      <div className="oe-root">

        {toast && <div className={`oe-toast ${toast.type}`}>{toast.msg}</div>}

        {/* Header */}
        <div className="oe-header">
          <div>
            <h1 className="oe-title">Edit <em>Order #{id}</em></h1>
            {order?.status && (
              <div className="oe-meta">Status: {order.status}</div>
            )}
          </div>
          <div className="oe-header-actions">
            <button className="btn-ghost" onClick={() => navigate('/orders')}>← Back</button>
            <button className="btn-primary" onClick={handleSave} disabled={saving}>
              {saving ? 'Saving…' : 'Save order'}
            </button>
          </div>
        </div>

        {saveError && <div className="oe-error-banner">{saveError}</div>}

        <div className="oe-layout">

          {/* ── Left column ── */}
          <div className="oe-left">

            {/* Order details */}
            <div className="oe-section">
              <div className="oe-section-head">
                <div className="oe-section-title">Order <em>details</em></div>
              </div>
              <div className="oe-section-body">

                {/* Customer */}
                <div className="oe-customer-row">
                  <div className="oe-field">
                    <label className="oe-label">Customer</label>
                    <select className="oe-select" value={formData.customer_id}
                      onChange={(e) => handleCustomerChange(e.target.value)}>
                      <option value="">No customer</option>
                      {customers.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </select>
                  </div>
                  <button className="btn-text" onClick={() => { setShowNewCust((o) => !o); setCustError(null); }}>
                    {showNewCust ? '✕ cancel' : '+ new'}
                  </button>
                </div>

                {/* Inline new customer */}
                <div className={`oe-new-cust-wrap ${showNewCust ? 'open' : 'closed'}`}>
                  <form className="oe-new-cust" onSubmit={handleNewCustomer}>
                    <div className="oe-new-cust-title">New customer</div>
                    {custError && <div className="oe-error-banner" style={{ margin: 0 }}>{custError}</div>}
                    <div className="oe-new-cust-grid">
                      <div className="oe-field full">
                        <label className="oe-label">Name *</label>
                        <input {...custField('name')} placeholder="Full name" required />
                      </div>
                      <div className="oe-field">
                        <label className="oe-label">Phone</label>
                        <input {...custField('phone')} placeholder="604-555-0100" />
                      </div>
                      <div className="oe-field">
                        <label className="oe-label">Email</label>
                        <input {...custField('email')} type="email" placeholder="name@example.com" />
                      </div>
                      <div className="oe-field full">
                        <label className="oe-label">Address</label>
                        <input {...custField('address')} placeholder="Delivery address" />
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <button type="submit" className="btn-primary" disabled={custSaving}>
                        {custSaving ? 'Adding…' : 'Add customer'}
                      </button>
                      <button type="button" className="btn-ghost"
                        onClick={() => { setShowNewCust(false); setCustError(null); }}>
                        Cancel
                      </button>
                    </div>
                  </form>
                </div>

                {/* Order history */}
                {historyVisible && orderHistory.length > 0 && (
                  <div className="oe-history">
                    <div className="oe-history-title">
                      Order history · {orderHistory.length} previous order{orderHistory.length !== 1 ? 's' : ''}
                    </div>
                    {orderHistory.map((o) => (
                      <div key={o.id} className="oe-hist-row">
                        <div className="oe-hist-top">
                          <span className="oe-hist-id">#{o.id}</span>
                          <span className="oe-hist-date">{fmtDate(o.created_at)}</span>
                        </div>
                        <div className="oe-hist-items">
                          {o.order_items?.length > 0 ? (
                            <div>
                              {o.order_items.map((i) => (
                                <div key={i.item_id}>{i.quantity}× {i.item_name}</div>
                              ))}
                            </div>
                          ) : 'No items'}
                        </div>
                        <div className="oe-hist-footer">
                          <span className="oe-hist-total">{fmtUsd(o.total_amount)}</span>
                          <span className="oe-hist-status">{o.status}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Address + notes */}
                <div className="oe-field">
                  <label className="oe-label">Delivery Address</label>
                  <input {...field('delivery_address')} placeholder="Street address" />
                </div>
                <div className="oe-field">
                  <label className="oe-label">Notes</label>
                  <input {...field('notes')} placeholder="Special instructions" />
                </div>

                {/* Payment + schedule */}
                <div className="oe-field-row">
                  <div className="oe-field">
                    <label className="oe-label">Payment Method</label>
                    <select className="oe-select" value={formData.payment_method}
                      onChange={(e) => setFormData((p) => ({ ...p, payment_method: e.target.value }))}>
                      <option value="cash">Cash</option>
                      <option value="e-transfer">e-Transfer</option>
                    </select>
                  </div>
                  <div className="oe-field">
                    <label className="oe-label">Scheduled Date</label>
                    <input className="oe-input" type="datetime-local"
                      value={formData.scheduled_date}
                      onChange={(e) => setFormData((p) => ({ ...p, scheduled_date: e.target.value }))} />
                  </div>
                </div>

              </div>
            </div>

            {/* Current items */}
            <div className="oe-section">
              <div className="oe-section-head">
                <div className="oe-section-title">Items <em>in order</em></div>
              </div>
              <div className="oe-section-body">
                {selectedItems.length === 0 ? (
                  <p className="oe-no-items">No items selected — add from the menu below.</p>
                ) : (
                  <div className="oe-items-list">
                    {selectedItems.map((sel) => (
                      <div key={sel.item_id} className="oe-item-row">
                        <span className="oe-item-name">{sel.name}</span>
                        <span className="oe-item-price">{fmtUsd(sel.price)}</span>
                        <div className="oe-qty">
                          <button type="button" className="oe-qty-btn"
                            onClick={() => setQty(sel.item_id, sel.quantity - 1)}>−</button>
                          <input type="number" min="0" className="oe-qty-input"
                            value={sel.quantity}
                            onChange={(e) => setQty(sel.item_id, e.target.value)} />
                          <button type="button" className="oe-qty-btn"
                            onClick={() => setQty(sel.item_id, sel.quantity + 1)}>+</button>
                        </div>
                        <span className="oe-item-sub">{fmtUsd(sel.price * sel.quantity)}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Menu — add items */}
            <div className="oe-section">
              <div className="oe-section-head">
                <div className="oe-section-title">Add from <em>menu</em></div>
              </div>
              <div className="oe-section-body">
                {Object.entries(groupedItems).map(([cat, catItems]) => (
                  <div key={cat}>
                    <div className="oe-cat-head">
                      <h3 className="oe-cat-name">{cat}</h3>
                      <span className="oe-cat-count">{catItems.length}</span>
                    </div>
                    <div className="oe-menu-grid">
                      {catItems.map((item) => {
                        const qty = orderItems[item.id] || 0;
                        return (
                          <div key={item.id} className={`oe-menu-card${qty > 0 ? ' selected' : ''}`}>
                            <div className="oe-menu-card-top">
                              <span className="oe-menu-name">{item.name}</span>
                              <span className="oe-menu-price">{fmtUsd(item.price)}</span>
                            </div>
                            <div className="oe-qty">
                              <button type="button" className="oe-qty-btn"
                                onClick={() => setQty(item.id, qty - 1)}>−</button>
                              <input type="number" min="0" className="oe-qty-input"
                                value={qty}
                                onChange={(e) => setQty(item.id, e.target.value)} />
                              <button type="button" className="oe-qty-btn"
                                onClick={() => setQty(item.id, qty + 1)}>+</button>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </div>

          </div>{/* end .oe-left */}

          {/* ── Right column: sticky summary ── */}
          <div className="oe-right">
            <div className="oe-summary">
              <div className="oe-summary-head">
                <div className="oe-summary-title">Order <em>summary</em></div>
              </div>
              <div className="oe-summary-body">
                {selectedItems.length === 0
                  ? <div className="oe-sum-empty">No items selected</div>
                  : selectedItems.map((i) => (
                      <div key={i.item_id} className="oe-sum-line">
                        <span className="oe-sum-name">{i.name}</span>
                        <span className="oe-sum-qty">×{i.quantity}</span>
                        <span className="oe-sum-amt">{fmtUsd(i.price * i.quantity)}</span>
                      </div>
                    ))
                }
              </div>
              <div className="oe-sum-total">
                <span className="oe-sum-total-label">Total</span>
                <span className="oe-sum-total-value">{fmtUsd(total)}</span>
              </div>
              <div className="oe-summary-actions">
                <button className="btn-primary" onClick={handleSave} disabled={saving}>
                  {saving ? 'Saving…' : 'Save order'}
                </button>
                <button className="btn-ghost" onClick={() => navigate('/orders')}>← Back to orders</button>
              </div>
            </div>
          </div>

        </div>{/* end .oe-layout */}
      </div>
    </>
  );
}