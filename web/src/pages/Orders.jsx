import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';

function Orders() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [paymentFilter, setPaymentFilter] = useState('');
  const [expandedSections, setExpandedSections] = useState({
    pending: true,
    preparing: true,
    ready: true,
    delivered: false,
    cancelled: false
  });
  const [formData, setFormData] = useState({
    customer_id: '',
    delivery_address: '',
    total_amount: '',
    notes: '',
    scheduled_date: ''
  });
  const [customerData, setCustomerData] = useState({
    name: '',
    phone: '',
    email: '',
    address: ''
  });
  const [editingId, setEditingId] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [showCustomerForm, setShowCustomerForm] = useState(false);
  const [customers, setCustomers] = useState([]);

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
      const amount = parseFloat(formData.total_amount);
      if (isNaN(amount) || amount <= 0) {
        alert('Please enter a valid total amount');
        return;
      }
      const data = {
        ...formData,
        customer_id: parseInt(formData.customer_id) || null,
        total_amount: amount,
        status: 'pending',
        scheduled_date: formData.scheduled_date ? formData.scheduled_date + ':00Z' : null
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
      notes: order.notes,
      payment_method: order.payment_method || '',
      scheduled_date: order.scheduled_date ? order.scheduled_date.split('T')[0] : ''
    });
    setEditingId(order.id);
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingId(null);
    setShowCustomerForm(false);
    setFormData({ customer_id: '', delivery_address: '', total_amount: '', notes: '', scheduled_date: '' });
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

  const statusLabels = {
    pending: 'Pending',
    preparing: 'Preparing',
    ready: 'Ready',
    delivered: 'Delivered',
    cancelled: 'Cancelled'
  };

  const toggleSection = (status) => {
    setExpandedSections(prev => ({
      ...prev,
      [status]: !prev[status]
    }));
  };

  const expandAllSections = () => {
    setExpandedSections({
      pending: true,
      preparing: true,
      ready: true,
      delivered: true,
      cancelled: true
    });
  };

  const collapseAllSections = () => {
    setExpandedSections({
      pending: false,
      preparing: false,
      ready: false,
      delivered: false,
      cancelled: false
    });
  };

  const filteredOrders = orders.filter(order => {
    const matchesSearch = !searchTerm || 
      order.customer_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      order.delivery_address?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = !statusFilter || order.status === statusFilter;
    const matchesPayment = !paymentFilter || order.payment_method === paymentFilter;
    return matchesSearch && matchesStatus && matchesPayment;
  });

  // Group filtered orders by status
  const groupedOrders = {
    pending: filteredOrders.filter(o => o.status === 'pending'),
    preparing: filteredOrders.filter(o => o.status === 'preparing'),
    ready: filteredOrders.filter(o => o.status === 'ready'),
    delivered: filteredOrders.filter(o => o.status === 'delivered'),
    cancelled: filteredOrders.filter(o => o.status === 'cancelled')
  };

  // Filter statuses to show only the filtered status when statusFilter is active
  const visibleStatuses = statusFilter 
    ? Object.keys(groupedOrders).filter(status => status === statusFilter)
    : Object.keys(groupedOrders);

  if (loading) return <div className="loading">Loading orders...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Orders</h1>
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
        <select
          value={paymentFilter}
          onChange={(e) => setPaymentFilter(e.target.value)}
          className="filter-select"
        >
          <option value="">All Payment Methods</option>
          <option value="cash">Cash</option>
          <option value="e-transfer">e-Transfer</option>
        </select>
      </div>

      <div className="section-controls">
        <button className="btn-secondary" onClick={expandAllSections}>Expand All</button>
        <button className="btn-secondary" onClick={collapseAllSections}>Collapse All</button>
      </div>

      <div className="orders-by-status">
        {visibleStatuses.map((status) => (
          <div key={status} className="status-section">
            <button
              className="status-section-header"
              onClick={() => toggleSection(status)}
              style={{ borderLeftColor: statusColors[status] }}
            >
              <span className="section-toggle">
                {expandedSections[status] ? '▼' : '▶'}
              </span>
              <span className="section-title">{statusLabels[status]}</span>
              <span className="section-count">({groupedOrders[status].length})</span>
            </button>
            
            {expandedSections[status] && (
              <div className="status-section-content">
                {groupedOrders[status].length === 0 ? (
                  <p className="empty-section">No {statusLabels[status].toLowerCase()} orders</p>
                ) : (
                  <div className="card-grid">
                    {groupedOrders[status].map((order) => (
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
                          <div className="order-items-list">
                            <strong>Items:</strong>
                            {order.order_items && order.order_items.length > 0 ? (
                              <ul>
                                {order.order_items.map((item, index) => (
                                  <li key={index}>{item.quantity}x {item.item_name}</li>
                                ))}
                              </ul>
                            ) : (
                              <span className="no-items">No items</span>
                            )}
                          </div>
                           <p><strong>Total:</strong> ${order.total_amount}</p>
                            <p><strong>Payment Method:</strong> {order.payment_method === 'e-transfer' ? 'e-Transfer' : 'Cash'}</p>
                            {order.notes && <p><strong>Notes:</strong> {order.notes}</p>}
                             {order.scheduled_date && (
                               <p><strong>Scheduled:</strong> {new Date(order.scheduled_date).toLocaleString('en-US')}</p>
                             )}
                            {order.created_at && <p><strong>Created:</strong> {new Date(order.created_at).toLocaleString('en-US')}</p>}
                            {order.updated_at && order.updated_at !== order.created_at && (
                            <p><strong>Updated:</strong> {new Date(order.updated_at).toLocaleString('en-US')}</p>
                            )}
                        </div>
                        <div className="card-actions">
                          <Link to={`/orders/${order.id}/edit`} className="btn-primary">Edit</Link>
                          <select
                            value={order.status}
                            onChange={async (e) => {
                              try {
                                await api.updateOrder(order.id, {
                                  customer_id: order.customer_id,
                                  delivery_address: order.delivery_address,
                                  status: e.target.value,
                                  total_amount: order.total_amount,
                                  notes: order.notes
                                });
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
                )}
              </div>
            )}
          </div>
        ))}
      </div>
      {filteredOrders.length === 0 && <p className="empty">No orders matching your filters</p>}
    </div>
  );
}

export default Orders;
