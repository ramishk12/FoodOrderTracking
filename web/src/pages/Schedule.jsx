import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api';

function Schedule() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [days, setDays] = useState(7);

  useEffect(() => {
    loadOrders();
  }, [days]);

  const loadOrders = async () => {
    try {
      setLoading(true);
      const data = await api.getScheduledOrders(days);
      console.log('Loaded orders:', data);
      setOrders(data || []);
    } catch (err) {
      console.error('Error loading orders:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const groupOrdersByDate = () => {
    const groups = {};
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    orders.forEach(order => {
      if (!order.scheduled_date) return;
      
      const scheduledDate = new Date(order.scheduled_date);
      scheduledDate.setHours(0, 0, 0, 0);
      
      const diffDays = Math.floor((scheduledDate - today) / (1000 * 60 * 60 * 24));
      
      let label;
      if (diffDays < 0) {
        label = 'Overdue';
      } else if (diffDays === 0) {
        label = 'Today';
      } else if (diffDays === 1) {
        label = 'Tomorrow';
      } else if (diffDays <= 7) {
        label = 'This Week';
      } else {
        label = scheduledDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
      }

      if (!groups[label]) {
        groups[label] = [];
      }
      groups[label].push(order);
    });

    return groups;
  };

  const getOrderDateStatus = (scheduledDate) => {
    if (!scheduledDate) return 'normal';
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const date = new Date(scheduledDate);
    date.setHours(0, 0, 0, 0);
    
    const diffDays = Math.floor((date - today) / (1000 * 60 * 60 * 24));
    if (diffDays < 0) return 'overdue';
    if (diffDays === 0) return 'today';
    return 'normal';
  };

  const formatScheduledDate = (scheduledDate) => {
    if (!scheduledDate) return '';
    const date = new Date(scheduledDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);
    
    const dateOnly = new Date(date);
    dateOnly.setHours(0, 0, 0, 0);
    
    const timeStr = date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
    
    if (dateOnly.getTime() === today.getTime()) {
      return `Today at ${timeStr}`;
    } else if (dateOnly.getTime() === tomorrow.getTime()) {
      return `Tomorrow at ${timeStr}`;
    } else {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + ` at ${timeStr}`;
    }
  };

  const statusColors = {
    pending: '#f59e0b',
    preparing: '#3b82f6',
    ready: '#8b5cf6',
    delivered: '#10b981',
    cancelled: '#ef4444',
  };

  if (loading) return <div className="loading">Loading schedule...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  const groupedOrders = groupOrdersByDate();

  return (
    <div className="page">
      <div className="page-header">
        <h1>Order Schedule</h1>
        <div className="header-buttons">
          <select 
            value={days} 
            onChange={(e) => setDays(Number(e.target.value))}
            className="day-select"
          >
            <option value={3}>3 Days</option>
            <option value={7}>7 Days</option>
            <option value={14}>14 Days</option>
            <option value={30}>30 Days</option>
          </select>
        </div>
      </div>

      <div style={{ padding: '1rem' }}>
        <p>Total orders: {orders.length}</p>
        <p>Days filter: {days}</p>
      </div>

      {orders.length === 0 ? (
        <p className="empty">No orders scheduled for the next {days} days. Create an order and set a scheduled date.</p>
      ) : (
        <div className="schedule-groups">
          {Object.entries(groupedOrders).map(([dateLabel, dateOrders]) => (
            <div key={dateLabel} className={`schedule-group ${dateLabel === 'Overdue' ? 'overdue-group' : ''} ${dateLabel === 'Today' ? 'today-group' : ''}`}>
              <h2 className="schedule-date">
                {dateLabel}
                <span className="order-count">{dateOrders.length} order{dateOrders.length !== 1 ? 's' : ''}</span>
              </h2>
              <div className="schedule-orders">
                {dateOrders.map(order => (
                  <div key={order.id} className={`schedule-order-card ${getOrderDateStatus(order.scheduled_date)}`}>
                    <div className="schedule-order-header">
                      <span className="order-id">Order #{order.id}</span>
                      <span className="scheduled-time">{formatScheduledDate(order.scheduled_date)}</span>
                    </div>
                    <div className="schedule-order-header">
                      <span 
                        className="status-badge" 
                        style={{ backgroundColor: statusColors[order.status] || '#666' }}
                      >
                        {order.status}
                      </span>
                    </div>
                    <div className="schedule-order-customer">
                      {order.customer_name || 'No customer'}
                    </div>
                    <div className="schedule-order-items">
                      {order.order_items && order.order_items.length > 0 
                        ? order.order_items.map(item => `${item.quantity}x ${item.item_name}`).join(', ')
                        : 'No items'}
                    </div>
                    <div className="schedule-order-footer">
                      <span className="order-total">${order.total_amount}</span>
                      <Link to={`/orders/${order.id}/edit`} className="btn-primary btn-small">
                        Edit
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default Schedule;
