import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import CustomerOrderHistory from '../components/CustomerOrderHistory';
import '../index.css';

/* ─── Constants ──────────────────────────── */

const EMPTY_ITEM_FORM = { name: '', description: '', price: '', category: '' };
const EMPTY_ORDER_FORM = {
  customer_id: '', delivery_address: '', notes: '',
  payment_method: 'cash', scheduled_date: '',
};
const EMPTY_MOD_FORM  = { name: '', price_adjustment: '0' };
const EMPTY_CUST      = { name: '', phone: '', email: '', address: '' };

/* ─── Helpers ────────────────────────────── */

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

function localDatetimeToUtcIso(localStr) {
  if (!localStr) return null;
  return new Date(localStr).toISOString();
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

function ItemCard({
  item, timesAdded,
  onAddLine, onEdit, onDelete, onActivate, delay,
  modPanelOpen, onToggleModPanel,
  modForm, onModFormChange,
  modError, modSubmitting,
  editingMod, onStartEditMod, onCancelEditMod,
  onModSubmit, onDeleteMod,
}) {
  const inOrder     = timesAdded > 0;
  const unavailable = !item.available;
  const mods        = item.modifiers || [];

  return (
    <div
      className={`itm-card${inOrder ? ' selected' : ''}${unavailable ? ' unavailable' : ''}`}
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
          /* mirrors oe-menu-card bottom */
          <div className="itm-card-add-row">
            {inOrder && (
              <div className="oe-menu-added-badge">{timesAdded}× in order</div>
            )}
            <button
              type="button"
              className="oe-add-line-btn"
              onClick={() => onAddLine(item)}
            >
              + Add{inOrder ? ' another' : ''}
            </button>
          </div>
        )}
      </div>

      {/* Per-item modifier management panel (admin CRUD) */}
      <div className={`itm-mod-panel ${modPanelOpen ? 'open' : 'closed'}`}>
        <div className="itm-mod-panel-inner">
          {modError && (
            <div className="itm-form-error" style={{ marginBottom: 8, fontSize: 12 }}>
              {modError}
            </div>
          )}

          {mods.length > 0 && (
            <div className="itm-mod-list">
              {mods.map((mod) => {
                const isEditing = editingMod?.mod?.id === mod.id;
                return (
                  <div key={mod.id} className="itm-mod-row">
                    {isEditing ? (
                      <form className="itm-mod-inline-form" onSubmit={(e) => onModSubmit(e, item)}>
                        <input className="itm-mod-edit-input"
                          value={modForm?.name || ''}
                          onChange={(e) => onModFormChange(item.id, 'name', e.target.value)}
                          placeholder="Name" required />
                        <input className="itm-mod-edit-input itm-mod-price-input"
                          type="number" step="0.01"
                          value={modForm?.price_adjustment ?? '0'}
                          onChange={(e) => onModFormChange(item.id, 'price_adjustment', e.target.value)} />
                        <button type="submit" className="btn-ghost" style={{ fontSize: 11 }}
                          disabled={modSubmitting}>Save</button>
                        <button type="button" className="btn-ghost" style={{ fontSize: 11 }}
                          onClick={() => onCancelEditMod(item.id)}>✕</button>
                      </form>
                    ) : (
                      <>
                        <span className="itm-mod-name">{mod.name}</span>
                        <span className={`itm-mod-price${mod.price_adjustment > 0 ? ' positive' : mod.price_adjustment < 0 ? ' negative' : ''}`}>
                          {mod.price_adjustment > 0 ? '+' : ''}{mod.price_adjustment.toFixed(2)}
                        </span>
                        <div className="itm-mod-actions">
                          <button className="btn-ghost" style={{ padding: '3px 8px', fontSize: 11 }}
                            onClick={() => onStartEditMod(item, mod)}>Edit</button>
                          <button className="btn-danger" style={{ padding: '3px 8px', fontSize: 11 }}
                            onClick={() => onDeleteMod(item, mod)}>✕</button>
                        </div>
                      </>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {!editingMod && (
            <form className="itm-mod-inline-form itm-mod-add-form" onSubmit={(e) => onModSubmit(e, item)}>
              <input className="itm-mod-edit-input"
                value={modForm?.name || ''}
                onChange={(e) => onModFormChange(item.id, 'name', e.target.value)}
                placeholder="e.g. Extra Cheese" required />
              <input className="itm-mod-edit-input itm-mod-price-input"
                type="number" step="0.01"
                value={modForm?.price_adjustment ?? '0'}
                onChange={(e) => onModFormChange(item.id, 'price_adjustment', e.target.value)}
                placeholder="0.00" />
              <button type="submit" className="btn-primary" style={{ fontSize: 11, padding: '5px 12px' }}
                disabled={modSubmitting}>+ Add</button>
            </form>
          )}
        </div>
      </div>

      <div className="itm-card-footer">
        <button
          className={`itm-mod-toggle-btn${modPanelOpen ? ' active' : ''}`}
          onClick={onToggleModPanel}
        >
          {modPanelOpen ? 'Hide mods' : `Mods${mods.length > 0 ? ` (${mods.length})` : ''}`}
        </button>
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn-ghost" style={{ padding: '5px 11px', fontSize: 11 }}
            onClick={() => onEdit(item)}>Edit</button>
          {unavailable ? (
            <button className="btn-primary" style={{ background: 'var(--green)' }}
              onClick={() => onActivate(item)}>Activate</button>
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

  /*
   * orderLines mirrors OrderEdit's lineItems exactly:
   *   { lineId, itemId, quantity, modifiers[] }
   *   modifiers: { modifier_id, name, price_adjustment }
   */
  const [orderLines, setOrderLines]           = useState([]);
  const [openModPickers, setOpenModPickers] = useState({}); // lineId -> boolean

  const [itemFormOpen, setItemFormOpen]     = useState(false);
  const [editingItem, setEditingItem]       = useState(null);
  const [itemForm, setItemForm]             = useState(EMPTY_ITEM_FORM);
  const [itemFormError, setItemFormError]   = useState(null);
  const [itemSubmitting, setItemSubmitting] = useState(false);

  const [orderFormOpen, setOrderFormOpen]     = useState(false);
  const [orderForm, setOrderForm]             = useState(EMPTY_ORDER_FORM);
  const [orderFormError, setOrderFormError]   = useState(null);
  const [orderSubmitting, setOrderSubmitting] = useState(false);

  const [showNewCust, setShowNewCust] = useState(false);
  const [custForm, setCustForm]       = useState(EMPTY_CUST);
  const [custSaving, setCustSaving]   = useState(false);
  const [custError, setCustError]     = useState(null);

  const [search, setSearch]             = useState('');
  const [catFilter, setCatFilter]       = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [toast, setToast]               = useState(null);
  const [showUnavailable, setShowUnavailable] = useState(false);

  const [expandedModItems, setExpandedModItems] = useState({});
  const [modForms, setModForms]         = useState({});
  const [modErrors, setModErrors]       = useState({});
  const [modSubmitting, setModSubmitting] = useState(false);
  const [editingMod, setEditingMod]     = useState(null);

  /* ── Derived ── */
  const availableItems   = items.filter((i) => i.available);
  const unavailableItems = items.filter((i) => !i.available);
  const categories = [...new Set(availableItems.map((i) => i.category).filter(Boolean))].sort();

  const filteredItems = availableItems.filter((item) => {
    const s = search.toLowerCase();
    return (
      (!search || item.name?.toLowerCase().includes(s) || item.description?.toLowerCase().includes(s)) &&
      (!catFilter || item.category === catFilter)
    );
  });

  const groupedItems = filteredItems.reduce((acc, item) => {
    const cat = item.category || 'Uncategorised';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(item);
    return acc;
  }, {});

  const activeLines = orderLines.filter((l) => l.quantity > 0);

  const total = orderLines.reduce((sum, line) => {
    if (line.quantity <= 0) return sum;
    const item = items.find((i) => i.id === line.itemId);
    if (!item) return sum;
    const modAdj = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
    return sum + (item.price + modAdj) * line.quantity;
  }, 0);

  const totalQty = orderLines.reduce((sum, l) => sum + l.quantity, 0);

  /* ── Helpers ── */
  const showToast = (message, type = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  /* ── Line helpers — identical to OrderEdit ── */
  const addLine = (item) => {
    setOrderLines((p) => [...p, {
      lineId:    `line-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      itemId:    item.id,
      quantity:  1,
      modifiers: [],
    }]);
  };

  const removeLine = (lineId) => {
    setOrderLines((p) => p.filter((l) => l.lineId !== lineId));
    setOpenModPickers((p) => {
      const next = { ...p };
      delete next[lineId];
      return next;
    });
  };

  const setLineQty = (lineId, raw) => {
    const qty = Math.max(0, parseInt(raw) || 0);
    setOrderLines((p) => p.map((l) => l.lineId === lineId ? { ...l, quantity: qty } : l));
  };

  const toggleModifier = (lineId, mod) => {
    setOrderLines((p) => p.map((l) => {
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

  /* ── Load ── */
  const load = useCallback(async () => {
    try {
      setLoading(true); setError(null);
      const [itemsData, customersData] = await Promise.all([
        api.getItems(), api.getCustomers(),
      ]);
      setItems(itemsData || []);
      setCustomers(customersData || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    if (totalQty > 0 && !orderFormOpen) setOrderFormOpen(true);
    if (totalQty === 0 && orderFormOpen) setOrderFormOpen(false);
  }, [totalQty]);

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

  /* ── Delete / Activate ── */
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
  const closeOrderForm = () => {
    setOrderFormOpen(false);
    setOrderForm(EMPTY_ORDER_FORM);
    setOrderFormError(null);
    setOrderLines([]);
    setOpenModPicker(null);
    setShowNewCust(false);
    setCustForm(EMPTY_CUST);
    setCustError(null);
  };

  const custField = (key) => ({
    value:     custForm[key],
    onChange:  (e) => setCustForm((p) => ({ ...p, [key]: e.target.value })),
    className: 'oe-input',
  });

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

  const handleCustomerChange = (customerId) => {
    const customer = customers.find((c) => c.id === parseInt(customerId));
    setOrderForm((p) => ({
      ...p,
      customer_id:      customerId,
      delivery_address: customer?.address || p.delivery_address,
    }));
  };

  const handleOrderSubmit = async (e) => {
    e.preventDefault();
    if (activeLines.length === 0) {
      setOrderFormError('Add at least one item from the menu below.');
      return;
    }
    setOrderSubmitting(true); setOrderFormError(null);
    try {
      await api.createOrder({
        customer_id:      parseInt(orderForm.customer_id) || null,
        delivery_address: orderForm.delivery_address,
        notes:            orderForm.notes,
        payment_method:   orderForm.payment_method,
        scheduled_date:   localDatetimeToUtcIso(orderForm.scheduled_date),
        items: activeLines.map((line) => ({
          item_id:   line.itemId,
          quantity:  line.quantity,
          modifiers: line.modifiers.map((m) => ({
            modifier_id:      m.modifier_id,
            name:             m.name,
            price_adjustment: m.price_adjustment,
          })),
        })),
      });
      showToast('Order placed successfully');
      closeOrderForm();
    } catch (err) {
      setOrderFormError(err.message);
    } finally {
      setOrderSubmitting(false);
    }
  };

  const oeField = (key) => ({
    value:    orderForm[key],
    onChange: (e) => setOrderForm((p) => ({ ...p, [key]: e.target.value })),
    className: 'oe-input',
  });

  /* ── Modifier CRUD ── */
  const toggleModPanel = (itemId) => {
    setExpandedModItems((prev) => ({ ...prev, [itemId]: !prev[itemId] }));
    setEditingMod(null);
    setModForms((p) => ({ ...p, [itemId]: EMPTY_MOD_FORM }));
    setModErrors((p) => ({ ...p, [itemId]: null }));
  };

  const startEditMod = (item, mod) => {
    setEditingMod({ itemId: item.id, mod });
    setModForms((p) => ({
      ...p,
      [item.id]: { name: mod.name, price_adjustment: String(mod.price_adjustment) },
    }));
    setModErrors((p) => ({ ...p, [item.id]: null }));
  };

  const cancelEditMod = (itemId) => {
    setEditingMod(null);
    setModForms((p) => ({ ...p, [itemId]: EMPTY_MOD_FORM }));
    setModErrors((p) => ({ ...p, [itemId]: null }));
  };

  const handleModSubmit = async (e, item) => {
    e.preventDefault();
    setModSubmitting(true);
    setModErrors((p) => ({ ...p, [item.id]: null }));
    const form    = modForms[item.id] || EMPTY_MOD_FORM;
    const payload = { name: form.name, price_adjustment: parseFloat(form.price_adjustment) || 0 };
    try {
      if (editingMod && editingMod.itemId === item.id) {
        await api.updateItemModifier(item.id, editingMod.mod.id, payload);
      } else {
        await api.createItemModifier(item.id, payload);
      }
      setEditingMod(null);
      setModForms((p) => ({ ...p, [item.id]: EMPTY_MOD_FORM }));
      load();
    } catch (err) {
      setModErrors((p) => ({ ...p, [item.id]: err.message }));
    } finally {
      setModSubmitting(false);
    }
  };

  const handleDeleteMod = async (item, mod) => {
    try {
      await api.deleteItemModifier(item.id, mod.id);
      load();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const sharedCardProps = (item) => ({
    item,
    timesAdded:       orderLines.filter((l) => l.itemId === item.id).length,
    onAddLine:        addLine,
    onEdit:           openEdit,
    onDelete:         (i) => setDeleteTarget(i),
    onActivate:       handleActivate,
    modPanelOpen:     expandedModItems[item.id],
    onToggleModPanel: () => toggleModPanel(item.id),
    modForm:          modForms[item.id] || EMPTY_MOD_FORM,
    onModFormChange:  (itemId, key, val) =>
      setModForms((p) => ({ ...p, [itemId]: { ...(p[itemId] || EMPTY_MOD_FORM), [key]: val } })),
    modError:         modErrors[item.id],
    modSubmitting,
    editingMod:       editingMod?.itemId === item.id ? editingMod : null,
    onStartEditMod:   startEditMod,
    onCancelEditMod:  cancelEditMod,
    onModSubmit:      handleModSubmit,
    onDeleteMod:      handleDeleteMod,
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

        {/* Header — outside oe-layout, same as oe-header in OrderEdit */}
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

        {/* Item form — outside oe-layout */}
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

        {/* ── oe-layout: direct child of itm-root, exactly like OrderEdit ──
            oe-left  — order form panel + all menu content
            oe-right — sticky summary (position:sticky works because there
                       is no overflow:hidden ancestor between it and itm-root)
        ─────────────────────────────────────────────────────────────────── */}
        <div className="oe-layout">
          <div className="oe-left">

            {/* ── Order form ── */}
            {orderFormOpen && (
              <form id="itm-order-form" onSubmit={handleOrderSubmit}>

                {orderFormError && <div className="oe-error-banner">{orderFormError}</div>}

                {/* Order details section */}
                <div className="oe-section">
                  <div className="oe-section-head">
                    <div className="oe-section-title">Order <em>details</em></div>
                  </div>
                  <div className="oe-section-body">

                    <div className="oe-customer-row">
                      <div className="oe-field">
                        <label className="oe-label">Customer</label>
                        <select className="oe-select" value={orderForm.customer_id}
                          onChange={(e) => handleCustomerChange(e.target.value)}>
                          <option value="">No customer</option>
                          {customers.map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                          ))}
                        </select>
                      </div>
                      <button type="button" className="btn-text"
                        onClick={() => { setShowNewCust((o) => !o); setCustError(null); }}>
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

                    <CustomerOrderHistory
                      customerId={orderForm.customer_id ? parseInt(orderForm.customer_id) : null}
                      variant="oe"
                    />

                    <div className="oe-field">
                      <label className="oe-label">Delivery Address</label>
                      <input {...oeField('delivery_address')} placeholder="Street address" />
                    </div>

                    <div className="oe-field">
                      <label className="oe-label">Notes</label>
                      <input {...oeField('notes')} placeholder="Special instructions" />
                    </div>

                    <div className="oe-field-row">
                      <div className="oe-field">
                        <label className="oe-label">Payment Method</label>
                        <select className="oe-select" value={orderForm.payment_method}
                          onChange={(e) => setOrderForm((p) => ({ ...p, payment_method: e.target.value }))}>
                          <option value="cash">Cash</option>
                          <option value="e-transfer">e-Transfer</option>
                        </select>
                      </div>
                      <div className="oe-field">
                        <label className="oe-label">Scheduled Date</label>
                        <input className="oe-input" type="datetime-local"
                          value={orderForm.scheduled_date}
                          onChange={(e) => setOrderForm((p) => ({ ...p, scheduled_date: e.target.value }))} />
                      </div>
                    </div>

                  </div>
                </div>

                {/* Items in order section */}
                <div className="oe-section">
                  <div className="oe-section-head">
                    <div className="oe-section-title">Items <em>in order</em></div>
                  </div>
                  <div className="oe-section-body">
                    {orderLines.length === 0 ? (
                      <p className="oe-no-items">No items selected — add from the menu below.</p>
                    ) : (
                      <div className="oe-items-list">
                        {orderLines
                          .map((line) => {
                            const item = items.find((i) => i.id === line.itemId);
                            if (!item) return null;
                            return { ...line, item };
                          })
                          .filter(Boolean)
                          .sort((a, b) => {
                            if (a.item.name !== b.item.name) return a.item.name.localeCompare(b.item.name);
                            return a.modifiers.map((m) => m.name).sort().join(',')
                              .localeCompare(b.modifiers.map((m) => m.name).sort().join(','));
                          })
                          .map((line) => {
                            const { item }    = line;
                            const itemMods    = item.modifiers || [];
                            const modAdj      = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
                            const lineTotal   = (item.price + modAdj) * line.quantity;
                            const isPickerOpen = !!openModPickers[line.lineId];
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
                                  {itemMods.length > 0 && (
                                    <button type="button" className="oe-mod-toggle"
                                      onClick={() => setOpenModPickers((p) => ({ ...p, [line.lineId]: !p[line.lineId] }))}>
                                      {isPickerOpen ? 'Hide mods' : `Mods${line.modifiers.length > 0 ? ` (${line.modifiers.length})` : ''}`}
                                    </button>
                                  )}
                                  <button type="button" className="oe-remove-line"
                                    onClick={() => removeLine(line.lineId)} title="Remove line">✕</button>
                                </div>

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

                                {isPickerOpen && (() => {
                                  if (itemMods.length === 0) return (
                                    <div className="oe-mod-picker">
                                      <div className="oe-mod-picker-empty">
                                        No modifiers for this item — add them via the Mods button on the card.
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

              </form>
            )}
            {/* end order form */}

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

            {/* Menu content */}
            {!loading && !error && (
              <>
              <div className="itm-menu-box">
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

                {Object.entries(groupedItems).map(([category, catItems], gi) => (
                  <div key={category} className="itm-category"
                    style={{ animationDelay: `${gi * 60}ms` }}>
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
                          delay={gi * 60 + ii * 40}
                          {...sharedCardProps(item)}
                        />
                      ))}
                    </div>
                  </div>
                ))}

                {unavailableItems.length > 0 && (
                  <div className="itm-unavailable-section">
                    <button className="itm-unavailable-toggle"
                      onClick={() => setShowUnavailable(!showUnavailable)}>
                      <span className="itm-unavailable-toggle-icon">{showUnavailable ? '▼' : '▶'}</span>
                      <span className="itm-unavailable-title">Unavailable Items</span>
                      <span className="itm-unavailable-count">
                        {unavailableItems.length} item{unavailableItems.length !== 1 ? 's' : ''}
                      </span>
                    </button>
                    {showUnavailable && (
                      <div className="itm-grid">
                        {unavailableItems.map((item, ii) => (
                          <ItemCard
                            key={item.id}
                            delay={ii * 40}
                            {...sharedCardProps(item)}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>{/* end itm-menu-box */}
              </>
            )}

          </div>{/* end oe-left */}

          {/* ── Right column: sticky summary — direct child of oe-layout,
               same position as in OrderEdit, so oe-right's position:sticky
               works against the real page scroll with no clipping ancestor */}
          {orderFormOpen && (
            <div className="oe-right">
              <div className="oe-summary">
                <div className="oe-summary-head">
                  <div className="oe-summary-title">Order <em>summary</em></div>
                </div>
                <div className="oe-summary-body">
                  {orderLines.length === 0
                    ? <div className="oe-sum-empty">No items selected</div>
                    : orderLines.filter((l) => l.quantity > 0).map((line) => {
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
                            <span className="oe-sum-amt">
                              {fmtUsd((item.price + modAdj) * line.quantity)}
                            </span>
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
                  <button type="submit" form="itm-order-form" className="btn-order"
                    disabled={orderSubmitting || activeLines.length === 0}>
                    {orderSubmitting ? 'Placing…' : `Place order · ${fmtUsd(total)}`}
                  </button>
                  <button type="button" className="btn-ghost" onClick={closeOrderForm}>Cancel</button>
                </div>
              </div>
            </div>
          )}

        </div>{/* end oe-layout */}

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