import { useState, useEffect } from 'react';
import { api } from '../services/api';
import CustomerOrderHistory from '../components/CustomerOrderHistory';
import '../index.css';

/* ─── Constants ──────────────────────────── */

const EMPTY_ITEM_FORM = { name: '', description: '', price: '', category: '' };
const EMPTY_ORDER_FORM = {
  customer_id: '', delivery_address: '', notes: '',
  payment_method: 'cash', scheduled_date: '',
};
const EMPTY_MOD_FORM = { name: '', price_adjustment: '0' };

/* ─── Helpers ────────────────────────────── */

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
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

/* ─── ConfirmDialog ──────────────────────────────── */

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
  item, qty, onQtyChange, onEdit, onDelete, onActivate, delay,
  modPanelOpen, onToggleModPanel,
  modForm, onModFormChange,
  modError, modSubmitting,
  editingMod, onStartEditMod, onCancelEditMod,
  onModSubmit, onDeleteMod,
}) {
  const unavailable = !item.available;

  return (
    <div
      className={`itm-card${qty > 0 ? ' selected' : ''}${unavailable ? ' unavailable' : ''}`}
      style={{ animationDelay: `${delay * 50}ms` }}
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
          <>
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
          </>
        )}
      </div>

      {unavailable ? (
        <div className="itm-card-footer">
          <button className="itm-activate-btn" onClick={() => onActivate(item)}>
            Make Available
          </button>
          <div className="itm-card-actions">
            <button className="itm-action-btn" onClick={() => onEdit(item)}>Edit</button>
            <button className="itm-action-btn" onClick={() => onDelete(item)}>Delete</button>
          </div>
        </div>
      ) : (
        <div className="itm-card-footer">
          <div className="itm-card-actions">
            <button className="itm-action-btn" onClick={() => onEdit(item)}>Edit</button>
            <button className="itm-action-btn" onClick={() => onDelete(item)}>Delete</button>
          </div>
          <button
            className={`itm-mod-toggle-btn${modPanelOpen ? ' active' : ''}`}
            onClick={onToggleModPanel}
          >
            Mods ({item.modifiers?.length || 0})
          </button>
        </div>
      )}

      {modPanelOpen && (
        <div className="itm-mod-panel">
          <div className="itm-mod-panel-inner">
            <div className="itm-mod-header">
              <span className="itm-mod-title">Modifiers for <em>{item.name}</em></span>
              <button className="itm-mod-close" onClick={onToggleModPanel}>✕</button>
            </div>

            {item.modifiers?.length > 0 ? (
              <div className="itm-mod-list">
                {item.modifiers.map((mod) => (
                  <div key={mod.id} className="itm-mod-row">
                    {editingMod?.id === mod.id ? (
                      <div className="itm-mod-edit-form">
                        <input
                          className="itm-mod-edit-input"
                          value={editingMod.name}
                          onChange={(e) => onModFormChange({ ...editingMod, name: e.target.value })}
                        />
                        <input
                          className="itm-mod-price-input"
                          type="number"
                          step="0.01"
                          value={editingMod.price_adjustment}
                          onChange={(e) => onModFormChange({ ...editingMod, price_adjustment: parseFloat(e.target.value) || 0 })}
                        />
                        <button className="btn-primary" onClick={() => onModSubmit(item.id)} disabled={modSubmitting}>
                          Save
                        </button>
                        <button className="btn-ghost" onClick={onCancelEditMod}>Cancel</button>
                      </div>
                    ) : (
                      <>
                        <span className="itm-mod-name">{mod.name}</span>
                        <span className={`itm-mod-price${mod.price_adjustment > 0 ? ' positive' : mod.price_adjustment < 0 ? ' negative' : ''}`}>
                          {mod.price_adjustment > 0 ? '+' : ''}{fmtUsd(mod.price_adjustment)}
                        </span>
                        <div className="itm-mod-actions">
                          <button className="itm-mod-edit" onClick={() => onStartEditMod(mod)}>Edit</button>
                          <button className="itm-mod-delete" onClick={() => onDeleteMod(item.id, mod)}>Delete</button>
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="itm-mod-empty">No modifiers yet</div>
            )}

            <div className="itm-mod-add-form">
              <div className="itm-mod-add-title">Add new modifier</div>
              {modError && <div className="itm-mod-error">{modError}</div>}
              <div className="itm-mod-add-inputs">
                <input
                  className="itm-mod-name-input"
                  placeholder="Name (e.g., Extra Cheese)"
                  value={modForm.name}
                  onChange={(e) => onModFormChange({ ...modForm, name: e.target.value })}
                />
                <input
                  className="itm-mod-price-input"
                  type="number"
                  step="0.01"
                  placeholder="+$0.00"
                  value={modForm.price_adjustment}
                  onChange={(e) => onModFormChange({ ...modForm, price_adjustment: parseFloat(e.target.value) || 0 })}
                />
              </div>
              <button className="btn-primary" onClick={() => onModSubmit(item.id)} disabled={modSubmitting}>
                {modSubmitting ? 'Adding...' : 'Add Modifier'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

/* ─── Main Component ───────────────────────── */

export default function Items() {
  const [items, setItems] = useState([]);
  const [customers, setCustomers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Item form state
  const [itemFormOpen, setItemFormOpen] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const [itemForm, setItemForm] = useState(EMPTY_ITEM_FORM);
  const [itemFormError, setItemFormError] = useState(null);
  const [itemSubmitting, setItemSubmitting] = useState(false);

  // Order form state
  const [orderFormOpen, setOrderFormOpen] = useState(false);
  const [orderForm, setOrderForm] = useState(EMPTY_ORDER_FORM);
  const [orderFormError, setOrderFormError] = useState(null);
  const [orderSubmitting, setOrderSubmitting] = useState(false);

  // Line items for current order - each line can have its own modifiers
  const [orderLines, setOrderLines] = useState([]);
  const [openModPicker, setOpenModPicker] = useState(null); // lineId with open picker

  const [search, setSearch] = useState('');
  const [catFilter, setCatFilter] = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [toast, setToast] = useState(null);
  const [showUnavailable, setShowUnavailable] = useState(false);

  // Modifier state — track which item has its modifier panel expanded
  const [expandedModItem, setExpandedModItem] = useState(null);
  const [modForms, setModForms] = useState({});
  const [modErrors, setModErrors] = useState({});
  const [modSubmitting, setModSubmitting] = useState(false);
  const [editingMod, setEditingMod] = useState(null);

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

  // Calculate totals from orderLines
  const totalAmount = orderLines.reduce((sum, line) => {
    if (line.quantity <= 0) return sum;
    const modAdj = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
    return sum + (line.item.price + modAdj) * line.quantity;
  }, 0);

  const totalQty = orderLines.reduce((sum, line) => sum + line.quantity, 0);

  /* ── Helpers ── */
  const showToast = (message, type = 'success') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  };

  const resetQty = () => {
    setOrderLines([]);
  };

  // Add a new line for an item
  const addLine = (item) => {
    setOrderLines((prev) => [...prev, {
      lineId: `line-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      itemId: item.id,
      item: item,
      quantity: 1,
      modifiers: [],
    }]);
  };

  // Set quantity for a specific line
  const setLineQty = (lineId, raw) => {
    const qty = Math.max(0, parseInt(raw) || 0);
    setOrderLines((prev) => prev.map((l) => l.lineId === lineId ? { ...l, quantity: qty } : l));
  };

  // Remove a line
  const removeLine = (lineId) => {
    setOrderLines((prev) => prev.filter((l) => l.lineId !== lineId));
    if (openModPicker === lineId) setOpenModPicker(null);
  };

  // Toggle modifier for a line
  const toggleLineModifier = (lineId, mod) => {
    setOrderLines((prev) => prev.map((l) => {
      if (l.lineId !== lineId) return l;
      const has = l.modifiers.some((m) => m.id === mod.id);
      return {
        ...l,
        modifiers: has
          ? l.modifiers.filter((m) => m.id !== mod.id)
          : [...l.modifiers, { id: mod.id, name: mod.name, price_adjustment: mod.price_adjustment }],
      };
    }));
  };

  // Build lines grouped for summary display
  const groupedOrderLines = orderLines
    .filter((l) => l.quantity > 0)
    .sort((a, b) => (a.item.name || '').localeCompare(b.item.name || ''));

  /* ── Load ── */
  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const [itemsData, customersData] = await Promise.all([
          api.getItems(),
          api.getCustomers(),
        ]);
        setItems(itemsData || []);
        setCustomers(customersData || []);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  /* ── Auto-open order form when items selected ── */
  useEffect(() => {
    if (totalQty > 0 && !orderFormOpen) setOrderFormOpen(true);
  }, [totalQty]);

  /* ── Item form handlers ── */
  const openCreate = () => {
    setEditingItem(null);
    setItemForm(EMPTY_ITEM_FORM);
    setItemFormError(null);
    setItemFormOpen(true);
    setOrderFormOpen(false);
  };

  const openEdit = (item) => {
    setEditingItem(item);
    setItemForm({
      name: item.name,
      description: item.description || '',
      price: item.price?.toString() || '',
      category: item.category || '',
    });
    setItemFormError(null);
    setItemFormOpen(true);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const closeItemForm = () => {
    setItemFormOpen(false);
    setEditingItem(null);
    setItemForm(EMPTY_ITEM_FORM);
    setItemFormError(null);
  };

  const handleItemSubmit = async (e) => {
    e.preventDefault();
    const payload = {
      name: itemForm.name.trim(),
      description: itemForm.description.trim() || null,
      price: parseFloat(itemForm.price) || 0,
      category: itemForm.category.trim() || null,
    };
    if (!payload.name) {
      setItemFormError('Name is required');
      return;
    }

    setItemSubmitting(true);
    setItemFormError(null);
    try {
      if (editingItem) {
        await api.updateItem(editingItem.id, payload);
        showToast('Item updated');
      } else {
        await api.createItem(payload);
        showToast('Item created');
      }
      const itemsData = await api.getItems();
      setItems(itemsData || []);
      closeItemForm();
    } catch (err) {
      setItemFormError(err.message);
    } finally {
      setItemSubmitting(false);
    }
  };

  const handleDelete = async () => {
    try {
      await api.deleteItem(deleteTarget.id);
      showToast('Item deleted');
      const itemsData = await api.getItems();
      setItems(itemsData || []);
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      setDeleteTarget(null);
    }
  };

  const handleActivate = async (item) => {
    try {
      await api.activateItem(item.id);
      showToast('Item is now available');
      const itemsData = await api.getItems();
      setItems(itemsData || []);
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  /* ── Modifier form handlers ── */
  const toggleModPanel = (itemId) => {
    setExpandedModItem(expandedModItem === itemId ? null : itemId);
    setModForms((p) => ({ ...p, [itemId]: EMPTY_MOD_FORM }));
    setModErrors((p) => ({ ...p, [itemId]: null }));
  };

  const handleModFormChange = (itemId, form) => {
    setModForms((p) => ({ ...p, [itemId]: form }));
  };

  const startEditMod = (mod) => {
    setEditingMod(mod);
    setModForms((p) => ({ ...p, [mod.item_id]: { name: mod.name, price_adjustment: mod.price_adjustment } }));
  };

  const cancelEditMod = () => {
    setEditingMod(null);
  };

  const handleModSubmit = async (itemId) => {
    const form = modForms[itemId];
    if (!form?.name?.trim()) {
      setModErrors((p) => ({ ...p, [itemId]: 'Name is required' }));
      return;
    }

    setModSubmitting(true);
    setModErrors((p) => ({ ...p, [itemId]: null }));
    try {
      if (editingMod) {
        await api.updateItemModifier(itemId, editingMod.id, {
          name: form.name.trim(),
          price_adjustment: parseFloat(form.price_adjustment) || 0,
        });
        showToast('Modifier updated');
      } else {
        await api.createItemModifier(itemId, {
          name: form.name.trim(),
          price_adjustment: parseFloat(form.price_adjustment) || 0,
        });
        showToast('Modifier added');
      }
      const itemsData = await api.getItems();
      setItems(itemsData || []);
      setEditingMod(null);
      setModForms((p) => ({ ...p, [itemId]: EMPTY_MOD_FORM }));
    } catch (err) {
      setModErrors((p) => ({ ...p, [itemId]: err.message }));
    } finally {
      setModSubmitting(false);
    }
  };

  const handleDeleteMod = async (itemId, mod) => {
    try {
      await api.deleteItemModifier(itemId, mod.id);
      showToast('Modifier deleted');
      const itemsData = await api.getItems();
      setItems(itemsData || []);
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  /* ── Customer change ── */
  const handleCustomerChange = async (customerId) => {
    const customer = customers.find((c) => c.id === parseInt(customerId));
    setOrderForm((p) => ({
      ...p,
      customer_id: customerId,
      delivery_address: customer?.address || p.delivery_address,
    }));
  };

  /* ── Order submit ── */
  const handleOrderSubmit = async (e) => {
    e.preventDefault();
    if (orderLines.filter((l) => l.quantity > 0).length === 0) {
      setOrderFormError('Add at least one item');
      return;
    }

    setOrderSubmitting(true);
    setOrderFormError(null);
    try {
      await api.createOrder({
        customer_id: parseInt(orderForm.customer_id) || null,
        delivery_address: orderForm.delivery_address,
        status: 'pending',
        notes: orderForm.notes,
        payment_method: orderForm.payment_method,
        scheduled_date: localDatetimeToUtcIso(orderForm.scheduled_date),
        items: orderLines
          .filter((l) => l.quantity > 0)
          .map((line) => ({
            item_id: line.itemId,
            quantity: line.quantity,
            modifiers: line.modifiers.map((m) => ({
              modifier_id: m.id,
              name: m.name,
              price_adjustment: m.price_adjustment,
            })),
          })),
      });
      showToast('Order placed successfully');
      setOrderFormOpen(false);
      setOrderForm(EMPTY_ORDER_FORM);
      setOrderLines([]);
    } catch (err) {
      setOrderFormError(err.message);
    } finally {
      setOrderSubmitting(false);
    }
  };

  if (loading) return (
    <div className="itm-root">
      <div className="itm-load">
        <div className="itm-spinner" />
        <span className="itm-load-text">Loading…</span>
      </div>
    </div>
  );

  if (error) return (
    <div className="itm-root">
      <div className="itm-error-page">
        <p className="itm-error-msg">{error}</p>
        <button className="itm-retry" onClick={() => window.location.reload()}>Try again</button>
      </div>
    </div>
  );

  return (
    <div className="itm-root">
      {toast && <Toast message={toast.message} type={toast.type} />}
      {deleteTarget && (
        <ConfirmDialog
          title={`Delete ${deleteTarget.name}?`}
          body="This action cannot be undone."
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}

      {/* Header */}
      <div className="itm-header">
        <div>
          <h1 className="itm-title">Menu</h1>
          <p className="itm-subtitle">{availableItems.length} items</p>
        </div>
        <div className="itm-header-actions">
          <button className="btn-ghost" onClick={() => setShowUnavailable((p) => !p)}>
            {showUnavailable ? 'Hide' : 'Show'} Unavailable ({unavailableItems.length})
          </button>
          <button className="btn-primary" onClick={openCreate}>+ Add Item</button>
        </div>
      </div>

      {/* Item Form Modal */}
      {itemFormOpen && (
        <div className="itm-overlay" onClick={closeItemForm}>
          <div className="itm-form" onClick={(e) => e.stopPropagation()}>
            <div className="itm-form-head">
              <h2 className="itm-form-title">{editingItem ? 'Edit Item' : 'New Item'}</h2>
              <button className="itm-form-close" onClick={closeItemForm}>✕</button>
            </div>
            <form onSubmit={handleItemSubmit}>
              {itemFormError && <div className="itm-form-error">{itemFormError}</div>}
              <div className="itm-form-body">
                <div className="itm-field">
                  <label className="itm-label">Name *</label>
                  <input
                    className="itm-input"
                    value={itemForm.name}
                    onChange={(e) => setItemForm({ ...itemForm, name: e.target.value })}
                    placeholder="Pizza Margherita"
                    autoFocus
                  />
                </div>
                <div className="itm-field">
                  <label className="itm-label">Description</label>
                  <textarea
                    className="itm-textarea"
                    value={itemForm.description}
                    onChange={(e) => setItemForm({ ...itemForm, description: e.target.value })}
                    placeholder="Fresh tomatoes, mozzarella, basil..."
                  />
                </div>
                <div className="itm-field-row">
                  <div className="itm-field">
                    <label className="itm-label">Price</label>
                    <input
                      className="itm-input"
                      type="number"
                      step="0.01"
                      value={itemForm.price}
                      onChange={(e) => setItemForm({ ...itemForm, price: e.target.value })}
                      placeholder="12.99"
                    />
                  </div>
                  <div className="itm-field">
                    <label className="itm-label">Category</label>
                    <input
                      className="itm-input"
                      value={itemForm.category}
                      onChange={(e) => setItemForm({ ...itemForm, category: e.target.value })}
                      placeholder="Pizza"
                      list="categories"
                    />
                    <datalist id="categories">
                      {categories.map((c) => (
                        <option key={c} value={c} />
                      ))}
                    </datalist>
                  </div>
                </div>
              </div>
              <div className="itm-form-actions">
                <button type="button" className="btn-ghost" onClick={closeItemForm}>Cancel</button>
                <button type="submit" className="btn-primary" disabled={itemSubmitting}>
                  {itemSubmitting ? 'Saving...' : editingItem ? 'Update' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Order Form Modal */}
      {orderFormOpen && (
        <div className="itm-overlay" onClick={() => setOrderFormOpen(false)}>
          <div className="ord-form" onClick={(e) => e.stopPropagation()}>
            <div className="ord-form-head">
              <h2 className="ord-form-title">New Order</h2>
              <button className="ord-form-close" onClick={() => setOrderFormOpen(false)}>✕</button>
            </div>
            <form onSubmit={handleOrderSubmit}>
              {orderFormError && <div className="ord-form-error">{orderFormError}</div>}
              <div className="ord-form-body">
                <div className="ord-field">
                  <label className="ord-label">Customer</label>
                  <select
                    className="ord-select"
                    value={orderForm.customer_id}
                    onChange={(e) => handleCustomerChange(e.target.value)}
                  >
                    <option value="">No customer</option>
                    {customers.map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>

                <CustomerOrderHistory
                  customerId={orderForm.customer_id ? parseInt(orderForm.customer_id) : null}
                  variant="ord"
                  showTitle={true}
                />

                <div className="ord-field">
                  <label className="ord-label">Delivery Address</label>
                  <input
                    className="ord-input"
                    value={orderForm.delivery_address}
                    onChange={(e) => setOrderForm({ ...orderForm, delivery_address: e.target.value })}
                    placeholder="Street address"
                  />
                </div>

                <div className="ord-field-row">
                  <div className="ord-field">
                    <label className="ord-label">Payment</label>
                    <select
                      className="ord-select"
                      value={orderForm.payment_method}
                      onChange={(e) => setOrderForm({ ...orderForm, payment_method: e.target.value })}
                    >
                      <option value="cash">Cash</option>
                      <option value="e-transfer">e-Transfer</option>
                    </select>
                  </div>
                  <div className="ord-field">
                    <label className="ord-label">Schedule</label>
                    <input
                      className="ord-input"
                      type="datetime-local"
                      value={orderForm.scheduled_date}
                      onChange={(e) => setOrderForm({ ...orderForm, scheduled_date: e.target.value })}
                    />
                  </div>
                </div>

                <div className="ord-field">
                  <label className="ord-label">Notes</label>
                  <input
                    className="ord-input"
                    value={orderForm.notes}
                    onChange={(e) => setOrderForm({ ...orderForm, notes: e.target.value })}
                    placeholder="Special instructions"
                  />
                </div>

                {/* Order lines in this order */}
                <div className="ord-items-section">
                  <div className="ord-items-title">Items in order</div>
                  {groupedOrderLines.length === 0 ? (
                    <div className="ord-items-empty">No items — add from menu below</div>
                  ) : (
                    <div className="ord-items-list">
                      {groupedOrderLines.map((line) => {
                        const modAdj = line.modifiers.reduce((s, m) => s + (m.price_adjustment || 0), 0);
                        const lineTotal = (line.item.price + modAdj) * line.quantity;
                        const isPickerOpen = openModPicker === line.lineId;
                        return (
                          <div key={line.lineId} className="ord-line">
                            <div className="ord-line-main">
                              <span className="ord-line-name">{line.item.name}</span>
                              <div className="ord-line-qty">
                                <button type="button" className="ord-qty-btn" onClick={() => setLineQty(line.lineId, line.quantity - 1)}>−</button>
                                <input type="number" min="0" className="ord-qty-input" value={line.quantity} onChange={(e) => setLineQty(line.lineId, e.target.value)} />
                                <button type="button" className="ord-qty-btn" onClick={() => setLineQty(line.lineId, line.quantity + 1)}>+</button>
                              </div>
                              <span className="ord-line-price">{fmtUsd(lineTotal)}</span>
                              <button type="button" className="ord-mod-toggle" onClick={() => setOpenModPicker(isPickerOpen ? null : line.lineId)}>
                                {isPickerOpen ? 'Hide' : `Mods${line.modifiers.length > 0 ? ` (${line.modifiers.length})` : ''}`}
                              </button>
                              <button type="button" className="ord-remove-line" onClick={() => removeLine(line.lineId)}>✕</button>
                            </div>
                            {line.modifiers.length > 0 && (
                              <div className="ord-mods-applied">
                                {line.modifiers.map((m) => (
                                  <span key={m.id} className="ord-mod-pill">
                                    {m.name}
                                    {m.price_adjustment !== 0 && <span className="ord-mod-adj">{m.price_adjustment > 0 ? '+' : ''}{fmtUsd(m.price_adjustment)}</span>}
                                  </span>
                                ))}
                              </div>
                            )}
                            {isPickerOpen && line.item.modifiers?.length > 0 && (
                              <div className="ord-mod-picker">
                                {line.item.modifiers.map((mod) => {
                                  const active = line.modifiers.some((m) => m.id === mod.id);
                                  return (
                                    <button key={mod.id} type="button" className={`ord-mod-option${active ? ' active' : ''}`} onClick={() => toggleLineModifier(line.lineId, mod)}>
                                      {mod.name}
                                      {mod.price_adjustment !== 0 && <span className="ord-mod-adj">{mod.price_adjustment > 0 ? '+' : ''}{fmtUsd(mod.price_adjustment)}</span>}
                                    </button>
                                  );
                                })}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>

                <div className="ord-total">
                  <span className="ord-total-label">Total</span>
                  <span className="ord-total-value">{fmtUsd(totalAmount)}</span>
                </div>
              </div>
              <div className="ord-form-actions">
                <button type="button" className="btn-ghost" onClick={() => { setOrderFormOpen(false); resetQty(); }}>Cancel</button>
                <button type="submit" className="btn-primary" disabled={orderSubmitting || groupedOrderLines.length === 0}>
                  {orderSubmitting ? 'Placing...' : `Place Order · ${fmtUsd(totalAmount)}`}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="itm-filters">
        <input
          className="itm-search"
          placeholder="Search menu..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select className="itm-cat-select" value={catFilter} onChange={(e) => setCatFilter(e.target.value)}>
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {/* Menu Items */}
      {filteredItems.map((item, i) => (
        <ItemCard
          key={item.id}
          item={item}
          qty={0}
          onQtyChange={() => {}}
          onEdit={openEdit}
          onDelete={(item) => setDeleteTarget(item)}
          onActivate={handleActivate}
          delay={i}
          modPanelOpen={expandedModItem === item.id}
          onToggleModPanel={() => toggleModPanel(item.id)}
          modForm={modForms[item.id] || EMPTY_MOD_FORM}
          onModFormChange={(form) => handleModFormChange(item.id, form)}
          modError={modErrors[item.id]}
          modSubmitting={modSubmitting}
          editingMod={editingMod}
          onStartEditMod={startEditMod}
          onCancelEditMod={cancelEditMod}
          onModSubmit={() => handleModSubmit(item.id)}
          onDeleteMod={(modId, mod) => handleDeleteMod(item.id, mod)}
        />
      ))}

      {/* Unavailable Items */}
      {showUnavailable && unavailableItems.length > 0 && (
        <>
          <div className="itm-unavailable-header">Unavailable Items</div>
          {unavailableItems.map((item, i) => (
            <ItemCard
              key={item.id}
              item={item}
              qty={0}
              onQtyChange={() => {}}
              onEdit={openEdit}
              onDelete={(item) => setDeleteTarget(item)}
              onActivate={handleActivate}
              delay={i}
              modPanelOpen={expandedModItem === item.id}
              onToggleModPanel={() => toggleModPanel(item.id)}
              modForm={modForms[item.id] || EMPTY_MOD_FORM}
              onModFormChange={(form) => handleModFormChange(item.id, form)}
              modError={modErrors[item.id]}
              modSubmitting={modSubmitting}
              editingMod={editingMod}
              onStartEditMod={startEditMod}
              onCancelEditMod={cancelEditMod}
              onModSubmit={() => handleModSubmit(item.id)}
              onDeleteMod={(modId, mod) => handleDeleteMod(item.id, mod)}
            />
          ))}
        </>
      )}

      {/* Sticky order bar */}
      {totalQty > 0 && (
        <div className="itm-order-bar">
          <div className="itm-order-bar-left">
            <span className="itm-order-bar-qty">
              {totalQty} item{totalQty !== 1 ? 's' : ''} · {fmtUsd(totalAmount)}
            </span>
            <button className="itm-order-bar-clear" onClick={resetQty}>Clear</button>
          </div>
          <button className="itm-order-bar-btn" onClick={() => setOrderFormOpen(true)}>
            Review Order →
          </button>
        </div>
      )}
    </div>
  );
}
