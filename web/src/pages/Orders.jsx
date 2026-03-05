import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';

function Orders() {
  const [orders, setOrders] = useState([]);
  const [customers, setCustomers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [showCustomerForm, setShowCustomerForm] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [formData, setFormData] = useState({
    customer_id: '',
    delivery_address: '',
    total_amount: '',
    items: '',
    notes: ''
  });
  const [customerData, setCustomerData] = useState({
    name: '',
    phone: '',
    email: '',
    address: ''
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [ordersData, customersData] = await Promise.all([
        api.getOrders(),
        api.getCustomers()
      ]);
      setOrders(ordersData);
      setCustomers(customersData);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const data = {
        ...formData,
        customer_id: parseInt(formData.customer_id) || null,
        total_amount: parseFloat(formData.total_amount)
      };
      if (editingId) {
        await api.updateOrder(editingId, data);
      } else {
        await api.createOrder(data);
      }
      handleCancel();
      loadData();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleCustomerSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.createCustomer(customerData);
      setShowCustomerForm(false);
      setCustomerData({ name: '', phone: '', email: '', address: '' });
      loadData();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleEdit = (order) => {
    setFormData({
      customer_id: order.customer_id ? String(order.customer_id) : '',
      delivery_address: order.delivery_address || '',
      total_amount: String(order.total_amount) || '',
      items: order.items || '',
      notes: order.notes || ''
    });
    setEditingId(order.id);
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingId(null);
    setShowCustomerForm(false);
    setFormData({ customer_id: '', delivery_address: '', total_amount: '', items: '', notes: '' });
  };

  const handleCustomerChange = (customerId) => {
    const customer = customers.find(c => c.id === parseInt(customerId));
    setFormData(prev => ({
      ...prev,
      customer_id: customerId,
      delivery_address: customer?.address || prev.delivery_address
    }));
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this order?')) return;
    try {
      await api.deleteOrder(id);
      loadData();
    } catch (err) {
      alert(err.message);
    }
  };

  const statusColors = {
    pending: '#f59e0b',
    preparing: '#3b82f6',
    ready: '#8b5cf6',
    delivered: '#10b981',
    cancelled: '#ef4444',
  };

  const filteredOrders = orders.filter(order => {
    const matchesSearch = !searchTerm || 
      order.customer_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      order.delivery_address?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = !statusFilter || order.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  if (loading) return <div className="loading">Loading orders...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Orders</h1>
        <button className="btn-primary" onClick={() => showForm ? handleCancel() : setShowForm(true)}>
          {showForm ? 'Cancel' : 'New Order'}
        </button>
      </div>

      <div className="filters">
        <input
          type="text"
          placeholder="Search by customer or address..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="filter-select"
        >
          <option value="">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="preparing">Preparing</option>
          <option value="ready">Ready</option>
          <option value="delivered">Delivered</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="form">
          <h3>{editingId ? 'Edit Order' : 'New Order'}</h3>
          <select
            value={formData.customer_id}
            onChange={(e) => handleCustomerChange(e.target.value)}
          >
            <option value="">Select Customer</option>
            {customers.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          <button type="button" className="btn-secondary" onClick={() => setShowCustomerForm(!showCustomerForm)}>
            {showCustomerForm ? 'Cancel' : '+ New Customer'}
          </button>
          
          {showCustomerForm && (
            <div className="nested-form">
              <input
                type="text"
                placeholder="Customer Name"
                value={customerData.name}
                onChange={(e) => setCustomerData({ ...customerData, name: e.target.value })}
                required
              />
              <input
                type="text"
                placeholder="Phone"
                value={customerData.phone}
                onChange={(e) => setCustomerData({ ...customerData, phone: e.target.value })}
              />
              <input
                type="email"
                placeholder="Email"
                value={customerData.email}
                onChange={(e) => setCustomerData({ ...customerData, email: e.target.value })}
              />
              <input
                type="text"
                placeholder="Address"
                value={customerData.address}
                onChange={(e) => setCustomerData({ ...customerData, address: e.target.value })}
              />
              <button type="submit" className="btn-primary" onClick={handleCustomerSubmit}>Add Customer</button>
            </div>
          )}

          <input
            type="text"
            placeholder="Delivery Address"
            value={formData.delivery_address}
            onChange={(e) => setFormData({ ...formData, delivery_address: e.target.value })}
          />
          <input
            type="number"
            placeholder="Total Amount"
            step="0.01"
            value={formData.total_amount}
            onChange={(e) => setFormData({ ...formData, total_amount: e.target.value })}
            required
          />
          <input
            type="text"
            placeholder="Items (e.g., 2x Pizza)"
            value={formData.items}
            onChange={(e) => setFormData({ ...formData, items: e.target.value })}
          />
          <input
            type="text"
            placeholder="Notes"
            value={formData.notes}
            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
          />
          <div className="form-actions">
            <button type="submit" className="btn-primary">
              {editingId ? 'Update Order' : 'Create Order'}
            </button>
          </div>
        </form>
      )}

      <div className="card-grid">
        {filteredOrders.map((order) => (
          <div key={order.id} className="card">
            <div className="card-header">
              <span className="order-id">Order #{order.id}</span>
              <span 
                className="status-badge"
                style={{ backgroundColor: statusColors[order.status] || '#666' }}
              >
                {order.status}
              </span>
            </div>
            <div className="card-body">
              <p><strong>Customer:</strong> {order.customer_name || 'No Customer'}</p>
              <p><strong>Phone:</strong> {order.customer_phone || 'N/A'}</p>
              <p><strong>Address:</strong> {order.delivery_address}</p>
              <p><strong>Items:</strong> {order.items || 'N/A'}</p>
              <p><strong>Total:</strong> ${order.total_amount}</p>
              {order.notes && <p><strong>Notes:</strong> {order.notes}</p>}
              <p><strong>Created:</strong> {new Date(order.created_at).toLocaleString('en-US', { timeZone: 'America/Los_Angeles' })}</p>
              {order.updated_at && order.updated_at !== order.created_at && (
                <p><strong>Updated:</strong> {new Date(order.updated_at).toLocaleString('en-US', { timeZone: 'America/Los_Angeles' })}</p>
              )}
            </div>
            <div className="card-actions">
              <Link to={`/orders/${order.id}/edit`} className="btn-primary">Edit</Link>
              <select
                value={order.status}
                onChange={async (e) => {
                  try {
                    await api.updateOrder(order.id, { ...order, status: e.target.value });
                    loadData();
                  } catch (err) {
                    alert(err.message);
                  }
                }}
                className="status-select"
              >
                <option value="pending">Pending</option>
                <option value="preparing">Preparing</option>
                <option value="ready">Ready</option>
                <option value="delivered">Delivered</option>
                <option value="cancelled">Cancelled</option>
              </select>
              <button className="btn-danger" onClick={() => handleDelete(order.id)}>Delete</button>
            </div>
          </div>
        ))}
      </div>
      {orders.length === 0 && <p className="empty">No orders yet</p>}
    </div>
  );
}

export default Orders;
