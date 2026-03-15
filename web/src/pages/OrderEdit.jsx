import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import CustomerOrderHistory from '../components/CustomerOrderHistory';


/* ─── Helpers ────────────────────────────── */

const EMPTY_CUST = { name: '', phone: '', email: '', address: '' };

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
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
  // lineItems: array so the same item can appear multiple times with different modifiers.
  // Each entry: { lineId, itemId, quantity, modifiers[] }
  // Item-specific modifiers come from item.modifiers (loaded with items list).
  const [lineItems, setLineItems] = useState([]);
																					   
  const [openModPicker, setOpenModPicker] = useState(null); // lineId with picker open

  const [showNewCust, setShowNewCust] = useState(false);
  const [custForm, setCustForm]       = useState(EMPTY_CUST);
  const [custSaving, setCustSaving]   = useState(false);
  const [custError, setCustError]     = useState(null);

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

      // Hydrate line items from existing order, preserving saved modifiers
      const lines = (orderData.order_items || []).map((oi) => ({
        lineId: `line-${oi.id}`,
        itemId: oi.item_id,
        quantity: oi.quantity,
        modifiers: (oi.modifiers || []).map((m) => ({
          modifier_id: m.modifier_id,
          name: m.modifier_name,
          price_adjustment: m.price_adjustment,
        })),
      }));
      setLineItems(lines);
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

  /* ── Line item helpers ── */
  const addLine = (item) => {
    setLineItems((p) => [...p, {
      lineId: `line-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      itemId: item.id,
      quantity: 1,
      modifiers: [],
    }]);
  };

  const setLineQty = (lineId, raw) => {
    const qty = Math.max(0, parseInt(raw) || 0);
    setLineItems((p) => p.map((l) => l.lineId === lineId ? { ...l, quantity: qty } : l));
  };

  const removeLine = (lineId) => {
    setLineItems((p) => p.filter((l) => l.lineId !== lineId));
    if (openModPicker === lineId) setOpenModPicker(null);
  };

  const toggleModifier = (lineId, mod) => {
    setLineItems((p) => p.map((l) => {
      if (l.lineId !== lineId) return l;
      const has = l.modifiers.some((m) => m.modifier_id === mod.id);
      return {
        ...l,
        modifiers: has
          ? l.modifiers.filter((m) => m.modifier_id !== mod.id)
          : [...l.modifiers, { modifier_id: mod.id, name: mod.name, price_adjustment: mod.price_adjustment }],
      };
    }));
  };

  /* ── Derived ── */
  // Total including per-line modifier adjustments
  const total = lineItems.reduce((sum, line) => {
    if (line.quantity <= 0) return sum;
    const item = items.find((i) => i.id === line.itemId);
    if (!item) return sum;
    const modAdj = line.modifiers.reduce((ms, m) => ms + (m.price_adjustment || 0), 0);
    return sum + (item.price + modAdj) * line.quantity;
  }, 0);

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
      if (newCust?.id != null) handleCustomerChange(String(newCust.id));
    } catch (err) {
      setCustError(err.message);
    } finally {
      setCustSaving(false);
    }
  };

  /* ── Save ── */
  const handleSave = async () => {
    const activeLines = lineItems.filter((l) => l.quantity > 0);
    if (activeLines.length === 0) {
      setSaveError('Add at least one item to the order before saving.');
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
        items: activeLines.map((l) => ({
          item_id:   l.itemId,
          quantity:  l.quantity,
          modifiers: l.modifiers.map((m) => ({
            modifier_id:      m.modifier_id,
            name:             m.name,
            price_adjustment: m.price_adjustment,
          })),
        })),
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
    <div className="oe-root">
      <div className="oe-load">
        <div className="oe-spinner" />
        <span className="oe-load-text">Loading order…</span>
			  
      </div>
    </div>
  );

  if (error) return (
	  
						  
    <div className="oe-root">
      <div className="oe-error-page">
        <p className="oe-error-msg">{error}</p>
        <button className="oe-retry" onClick={load}>Try again</button>
			  
      </div>
    </div>
  );

  return (
	  
						  
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
                <CustomerOrderHistory 
                  customerId={formData.customer_id ? parseInt(formData.customer_id) : null} 
                  variant="oe"
                  excludeId={parseInt(id)}
                />

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
                {lineItems.length === 0 ? (
                  <p className="oe-no-items">No items selected — add from the menu below.</p>
                ) : (
                  <div className="oe-items-list">
                    {lineItems
                      .map((line, idx) => {
                        const item = items.find((i) => i.id === line.itemId);
                        if (!item) return null;
                        return { ...line, item, idx };
                      })
                      .filter(Boolean)
                      .sort((a, b) => {
                        const nameA = a.item?.name || '';
                        const nameB = b.item?.name || '';
                        if (nameA !== nameB) return nameA.localeCompare(nameB);
                        const modsA = a.modifiers.map((m) => m.name).sort().join(',');
                        const modsB = b.modifiers.map((m) => m.name).sort().join(',');
                        return modsA.localeCompare(modsB);
                      })
                      .map((line) => {
                      const item = line.item;
                      const modAdj = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
                      const lineTotal = (item.price + modAdj) * line.quantity;
                      const isPickerOpen = openModPicker === line.lineId;
                      return (
                        <div key={line.lineId} className="oe-item-row">
                          <div className="oe-item-row-main">
                            <span className="oe-item-name">{item.name}</span>
                            <span className="oe-item-price">{fmtUsd(item.price)}</span>
                            <div className="oe-qty">
                              <button type="button" className="oe-qty-btn"
                                onClick={() => setLineQty(line.lineId, line.quantity - 1)}>−</button>
                              <input type="number" min="0" className="oe-qty-input"
                                value={line.quantity}
                                onChange={(e) => setLineQty(line.lineId, e.target.value)} />
                              <button type="button" className="oe-qty-btn"
                                onClick={() => setLineQty(line.lineId, line.quantity + 1)}>+</button>
                            </div>
                            <span className="oe-item-sub">{fmtUsd(lineTotal)}</span>
                            <button type="button" className="oe-mod-toggle"
                              onClick={() => setOpenModPicker(isPickerOpen ? null : line.lineId)}>
                              {isPickerOpen ? 'Hide mods' : `Mods${line.modifiers.length > 0 ? ` (${line.modifiers.length})` : ''}`}
                            </button>
                            <button type="button" className="oe-remove-line"
                              onClick={() => removeLine(line.lineId)} title="Remove line">✕</button>
                          </div>
                          {/* Applied modifiers */}
                          {line.modifiers.length > 0 && (
                            <div className="oe-mods-applied">
                              {line.modifiers.map((m) => (
                                <span key={m.modifier_id} className="oe-mod-pill">
                                  {m.name}
                                  {m.price_adjustment !== 0 && (
                                    <span className="oe-mod-adj">
                                      {m.price_adjustment > 0 ? '+' : ''}{fmtUsd(m.price_adjustment)}
                                    </span>
                                  )}
                                </span>
                              ))}
                            </div>
                          )}
                          {/* Modifier picker — uses this item's own modifiers */}
                          {isPickerOpen && (() => {
                            const itemMods = item.modifiers || [];
                            if (itemMods.length === 0) return (
                              <div className="oe-mod-picker">
                                <div className="oe-mod-picker-empty">
                                  No modifiers for this item — add them on the Menu page.
                                </div>
                              </div>
                            );
                            return (
                              <div className="oe-mod-picker">
                                {itemMods.map((mod) => {
                                  const active = line.modifiers.some((m) => m.modifier_id === mod.id);
                                  return (
                                    <button key={mod.id} type="button"
                                      className={`oe-mod-option${active ? ' active' : ''}`}
                                      onClick={() => toggleModifier(line.lineId, mod)}>
                                      {mod.name}
                                      {mod.price_adjustment !== 0 && (
                                        <span className="oe-mod-opt-adj">
                                          {mod.price_adjustment > 0 ? '+' : ''}{fmtUsd(mod.price_adjustment)}
                                        </span>
                                      )}
                                    </button>
                                  );
                                })}
                              </div>
                            );
                          })()}
                        </div>
                      );
                    })}
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
                        const timesAdded = lineItems.filter((l) => l.itemId === item.id).length;
                        return (
                          <div key={item.id} className={`oe-menu-card${timesAdded > 0 ? ' selected' : ''}`}>
                            <div className="oe-menu-card-top">
                              <span className="oe-menu-name">{item.name}</span>
                              <span className="oe-menu-price">{fmtUsd(item.price)}</span>
                            </div>
                            {timesAdded > 0 && (
                              <div className="oe-menu-added-badge">{timesAdded}× in order</div>
                            )}
                            <button type="button" className="oe-add-line-btn"
                              onClick={() => addLine(item)}>
                              + Add{timesAdded > 0 ? ' another' : ''}
                            </button>
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
                {lineItems.length === 0
                  ? <div className="oe-sum-empty">No items selected</div>
                  : lineItems.filter((l) => l.quantity > 0).map((line) => {
                      const item = items.find((i) => i.id === line.itemId);
                      if (!item) return null;
                      const modAdj = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
                      return (
                        <div key={line.lineId} className="oe-sum-line">
                          <div className="oe-sum-name">
                            {item.name}
                            {line.modifiers.length > 0 && (
                              <div className="oe-sum-mods">
                                {line.modifiers.map((m) => m.name).join(', ')}
                              </div>
                            )}
                          </div>
                          <span className="oe-sum-qty">×{line.quantity}</span>
                          <span className="oe-sum-amt">{fmtUsd((item.price + modAdj) * line.quantity)}</span>
                        </div>
                      );
                    })
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
  );
}