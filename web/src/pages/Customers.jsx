import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import '../index.css';

/* ─── Helpers ────────────────────────────── */

const EMPTY_FORM = { name: '', phone: '', email: '', address: '' };

function fmtDate(iso) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
    timeZone: 'America/Los_Angeles',
  });
}

function initials(name) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0].toUpperCase())
    .join('');
}

/* ─── ConfirmDialog ──────────────────────── */

function ConfirmDialog({ name, onConfirm, onCancel }) {
  return (
    <div className="cst-overlay" onClick={onCancel}>
      <div className="cst-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="cst-dialog-title">Delete customer?</div>
        <div className="cst-dialog-body">
          <strong>{name}</strong> and all associated data will be permanently removed.
          This cannot be undone.
        </div>
        <div className="cst-dialog-actions">
          <button className="btn-ghost" onClick={onCancel}>Cancel</button>
          <button className="btn-primary" style={{ background: 'var(--red)' }} onClick={onConfirm}>
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── CustomerCard ───────────────────────── */

function CustomerCard({ customer, onEdit, onDelete, delay }) {
  return (
    <div className="cst-card" style={{ animationDelay: `${delay}ms` }}>
      <div className="cst-card-top">
        <div className="cst-avatar">{initials(customer.name)}</div>
        <div className="cst-card-info">
          <div className="cst-name">{customer.name}</div>
          <div className="cst-id">ID #{customer.id}</div>
        </div>
      </div>

      <div className="cst-card-body">
        {[
          { label: 'Phone',   value: customer.phone },
          { label: 'Email',   value: customer.email },
          { label: 'Address', value: customer.address },
        ].map(({ label, value }) => (
          <div key={label} className="cst-detail">
            <span className="cst-detail-label">{label}</span>
            <span className={`cst-detail-value${!value ? ' na' : ''}`}>
              {value || 'N/A'}
            </span>
          </div>
        ))}
      </div>

      <div className="cst-card-footer">
        <span className="cst-since">Since {fmtDate(customer.created_at)}</span>
        <div className="cst-card-actions">
          <button className="btn-ghost" style={{ padding: '6px 12px', fontSize: '11px' }}
            onClick={() => onEdit(customer)}>
            Edit
          </button>
          <button className="btn-danger" onClick={() => onDelete(customer)}>
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── Customers ──────────────────────────── */

export default function Customers() {
  const [customers, setCustomers]   = useState([]);
  const [loading, setLoading]       = useState(true);
  const [error, setError]           = useState(null);
  const [formOpen, setFormOpen]     = useState(false);
  const [editingId, setEditingId]   = useState(null);
  const [formData, setFormData]     = useState(EMPTY_FORM);
  const [formError, setFormError]   = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [search, setSearch]         = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null); // { id, name }
  const searchRef = useRef(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getCustomers();
      setCustomers(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  /* Form */
  const openCreate = () => {
    setFormData(EMPTY_FORM);
    setEditingId(null);
    setFormError(null);
    setFormOpen(true);
  };

  const openEdit = (customer) => {
    setFormData({
      name:    customer.name    || '',
      phone:   customer.phone   || '',
      email:   customer.email   || '',
      address: customer.address || '',
    });
    setEditingId(customer.id);
    setFormError(null);
    setFormOpen(true);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const closeForm = () => {
    setFormOpen(false);
    setEditingId(null);
    setFormData(EMPTY_FORM);
    setFormError(null);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setFormError(null);
    try {
      if (editingId) {
        await api.updateCustomer(editingId, formData);
      } else {
        await api.createCustomer(formData);
      }
      closeForm();
      load();
    } catch (err) {
      setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const field = (key) => ({
    value: formData[key],
    onChange: (e) => setFormData((prev) => ({ ...prev, [key]: e.target.value })),
    className: 'cst-input',
  });

  /* Delete */
  const confirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await api.deleteCustomer(deleteTarget.id);
      setDeleteTarget(null);
      load();
    } catch (err) {
      setDeleteTarget(null);
      setError(err.message);
    }
  };

  /* Filter */
  const filtered = customers.filter((c) => {
    if (!search) return true;
    const s = search.toLowerCase();
    return (
      c.name?.toLowerCase().includes(s) ||
      c.email?.toLowerCase().includes(s) ||
      c.phone?.toLowerCase().includes(s)
    );
  });

  /* ── Render ── */
  return (
    <>
      <div className="cst-root">

        {/* Delete confirm */}
        {deleteTarget && (
          <ConfirmDialog
            name={deleteTarget.name}
            onConfirm={confirmDelete}
            onCancel={() => setDeleteTarget(null)}
          />
        )}

        {/* Header */}
        <div className="cst-header">
          <div>
            <h1 className="cst-title">Our <em>Customers</em></h1>
            <div className="cst-meta">
              {customers.length} customer{customers.length !== 1 ? 's' : ''}
              {search && filtered.length !== customers.length && ` · ${filtered.length} shown`}
            </div>
          </div>
          <div className="cst-header-actions">
            {formOpen
              ? <button className="btn-ghost" onClick={closeForm}>✕ Cancel</button>
              : <button className="btn-primary" onClick={openCreate}>+ Add Customer</button>
            }
          </div>
        </div>

        {/* Slide-down form */}
        <div className={`cst-form-wrap ${formOpen ? 'open' : 'closed'}`}>
          <form className="cst-form" onSubmit={handleSubmit}>
            <div className="cst-form-head">
              <div className="cst-form-title">
                {editingId ? <>Edit <em>customer</em></> : <>New <em>customer</em></>}
              </div>
            </div>

            {formError && <div className="cst-form-error">{formError}</div>}

            <div className="cst-form-body">
              <div className="cst-field full">
                <label className="cst-label">Name *</label>
                <input {...field('name')} placeholder="Full name" required />
              </div>
              <div className="cst-field">
                <label className="cst-label">Phone</label>
                <input {...field('phone')} placeholder="e.g. 604-555-0100" />
              </div>
              <div className="cst-field">
                <label className="cst-label">Email</label>
                <input {...field('email')} type="email" placeholder="name@example.com" />
              </div>
              <div className="cst-field full">
                <label className="cst-label">Address</label>
                <input {...field('address')} placeholder="Delivery address" />
              </div>
            </div>

            <div className="cst-form-actions">
              <button type="submit" className="btn-primary" disabled={submitting}>
                {submitting ? 'Saving…' : editingId ? 'Update customer' : 'Create customer'}
              </button>
              <button type="button" className="btn-ghost" onClick={closeForm}>
                Cancel
              </button>
            </div>
          </form>
        </div>

        {/* Loading */}
        {loading && (
          <div className="cst-load">
            <div className="cst-spinner" />
            <span className="cst-load-text">Loading customers…</span>
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="cst-error">
            <p className="cst-error-msg">{error}</p>
            <button className="cst-retry" onClick={load}>Try again</button>
          </div>
        )}

        {/* Content */}
        {!loading && !error && (
          <>
            {/* Search */}
            <div className="cst-search-wrap">
              <span className="cst-search-icon">
                <SearchIcon />
              </span>
              <input
                ref={searchRef}
                className="cst-search"
                type="text"
                placeholder="Search by name, email or phone…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>

            {/* Empty state */}
            {customers.length === 0 && (
              <div className="cst-empty">
                <span style={{ fontSize: 32, opacity: 0.25 }}>✦</span>
                <div className="cst-empty-title">No customers yet</div>
                <div className="cst-empty-sub">Add your first customer to get started.</div>
              </div>
            )}

            {/* No search results */}
            {customers.length > 0 && filtered.length === 0 && (
              <div className="cst-empty">
                <div className="cst-empty-title">No results</div>
                <div className="cst-empty-sub">Try a different name, email, or phone number.</div>
              </div>
            )}

            {/* Grid */}
            <div className="cst-grid">
              {filtered.map((customer, i) => (
                <CustomerCard
                  key={customer.id}
                  customer={customer}
                  delay={i * 40}
                  onEdit={openEdit}
                  onDelete={(c) => setDeleteTarget({ id: c.id, name: c.name })}
                />
              ))}
            </div>
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
