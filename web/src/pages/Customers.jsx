import { useState, useEffect } from 'react';
import { api } from '../services/api';

function Customers() {
  const [customers, setCustomers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
    email: '',
    address: ''
  });

  useEffect(() => {
    loadCustomers();
  }, []);

  const loadCustomers = async () => {
    try {
      setLoading(true);
      const data = await api.getCustomers();
      setCustomers(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingId) {
        await api.updateCustomer(editingId, formData);
      } else {
        await api.createCustomer(formData);
      }
      setShowForm(false);
      setEditingId(null);
      setFormData({ name: '', phone: '', email: '', address: '' });
      loadCustomers();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleEdit = (customer) => {
    setFormData({
      name: customer.name,
      phone: customer.phone || '',
      email: customer.email || '',
      address: customer.address || ''
    });
    setEditingId(customer.id);
    setShowForm(true);
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this customer?')) return;
    try {
      await api.deleteCustomer(id);
      loadCustomers();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingId(null);
    setFormData({ name: '', phone: '', email: '', address: '' });
  };

  if (loading) return <div className="loading">Loading customers...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Customers</h1>
        <button className="btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Add Customer'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="form">
          <h3>{editingId ? 'Edit Customer' : 'New Customer'}</h3>
          <input
            type="text"
            placeholder="Name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />
          <input
            type="text"
            placeholder="Phone"
            value={formData.phone}
            onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
          />
          <input
            type="email"
            placeholder="Email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
          />
          <input
            type="text"
            placeholder="Address"
            value={formData.address}
            onChange={(e) => setFormData({ ...formData, address: e.target.value })}
          />
          <div className="form-actions">
            <button type="submit" className="btn-primary">
              {editingId ? 'Update' : 'Create'}
            </button>
            {editingId && (
              <button type="button" className="btn-secondary" onClick={handleCancel}>
                Cancel
              </button>
            )}
          </div>
        </form>
      )}

      <div className="card-grid">
        {customers.map((customer) => (
          <div key={customer.id} className="card">
            <div className="card-header">
              <h3>{customer.name}</h3>
              <span className="customer-id">ID: {customer.id}</span>
            </div>
            <div className="card-body">
              <p><strong>Phone:</strong> {customer.phone || 'N/A'}</p>
              <p><strong>Email:</strong> {customer.email || 'N/A'}</p>
              <p><strong>Address:</strong> {customer.address || 'N/A'}</p>
              <p><strong>Created:</strong> {new Date(customer.created_at).toLocaleDateString()}</p>
            </div>
            <div className="card-actions">
              <button className="btn-primary" onClick={() => handleEdit(customer)}>Edit</button>
              <button className="btn-danger" onClick={() => handleDelete(customer.id)}>Delete</button>
            </div>
          </div>
        ))}
      </div>
      {customers.length === 0 && <p className="empty">No customers yet</p>}
    </div>
  );
}

export default Customers;
