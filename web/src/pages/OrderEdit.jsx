import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';

function OrderEdit() {
  // Get order ID from URL params
  const { id } = useParams();
  const navigate = useNavigate();
  
  // Local state for order data
  const [order, setOrder] = useState(null);
  const [customers, setCustomers] = useState([]);
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // Form state for order details
  const [formData, setFormData] = useState({
    customer_id: '',
    delivery_address: '',
    notes: '',
    payment_method: 'cash',
    scheduled_date: ''
  });
  
  // Order items state - maps item_id to quantity
  const [orderItems, setOrderItems] = useState({});
  
  // Show/hide customer creation form
  const [showCustomerForm, setShowCustomerForm] = useState(false);
  const [customerData, setCustomerData] = useState({
    name: '',
    phone: '',
    email: '',
    address: ''
  });

  // Customer order history
  const [customerOrders, setCustomerOrders] = useState([]);
  const [showOrderHistory, setShowOrderHistory] = useState(false);

  // Load order, customers, and menu items on mount
  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    try {
      setLoading(true);
      
      // Fetch order with items, all customers, and all menu items
      const [orderData, customersData, itemsData] = await Promise.all([
        api.getOrder(id),
        api.getCustomers(),
        api.getItems()
      ]);
      
      setOrder(orderData);
      setCustomers(customersData);
      setItems(itemsData);
      
      // Initialize form data with existing order values
      // Convert UTC scheduled_date to local timezone for datetime-local input
      let localScheduledDate = '';
      if (orderData.scheduled_date) {
        // Parse UTC timestamp from database (e.g., "2026-03-06T14:30:26.052863Z")
        const utcDate = new Date(orderData.scheduled_date);
        // Convert to local timezone and format as datetime-local (YYYY-MM-DDTHH:mm)
        const year = utcDate.getFullYear();
        const month = String(utcDate.getMonth() + 1).padStart(2, '0');
        const day = String(utcDate.getDate()).padStart(2, '0');
        const hours = String(utcDate.getHours()).padStart(2, '0');
        const minutes = String(utcDate.getMinutes()).padStart(2, '0');
        localScheduledDate = `${year}-${month}-${day}T${hours}:${minutes}`;
      }
      
      setFormData({
        customer_id: orderData.customer_id ? String(orderData.customer_id) : '',
        delivery_address: orderData.delivery_address || '',
        notes: orderData.notes || '',
        payment_method: orderData.payment_method || 'cash',
        scheduled_date: localScheduledDate
      });
      
      // Convert order items array to object for easy quantity management
      // Key is item_id, value is quantity
      const itemQuantities = {};
      if (orderData.order_items && orderData.order_items.length > 0) {
        orderData.order_items.forEach(oi => {
          itemQuantities[oi.item_id] = oi.quantity;
        });
      }
      setOrderItems(itemQuantities);
      
      // Load customer's order history if customer is assigned
      if (orderData.customer_id) {
        try {
          const orderId = parseInt(id);
          const orders = await api.getOrdersByCustomer(orderData.customer_id);
          setCustomerOrders(orders.filter(o => o.id !== orderId && !isNaN(orderId)));
          setShowOrderHistory(true);
        } catch (err) {
          console.error('Error fetching customer orders:', err);
        }
      }
      
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Calculate total amount based on items and quantities
  const calculateTotal = () => {
    let total = 0;
    Object.entries(orderItems).forEach(([itemId, qty]) => {
      const item = items.find(i => i.id === parseInt(itemId));
      if (item && qty > 0) {
        total += item.price * qty;
      }
    });
    return total;
  };

  // Handle quantity change for an item
  const handleQuantityChange = (itemId, value) => {
    const qty = Math.max(0, parseInt(value) || 0);
    setOrderItems(prev => ({
      ...prev,
      [itemId]: qty
    }));
  };

  // Handle customer selection change - auto-fill delivery address
  const handleCustomerChange = async (customerId) => {
    const customer = customers.find(c => c.id === parseInt(customerId));
    setFormData(prev => ({
      ...prev,
      customer_id: customerId,
      delivery_address: customer?.address || prev.delivery_address
    }));

    // Fetch customer's order history
    if (customerId) {
      try {
        const orderId = parseInt(id);
        const orders = await api.getOrdersByCustomer(customerId);
        setCustomerOrders(orders.filter(o => o.id !== orderId && !isNaN(orderId)));
        setShowOrderHistory(true);
      } catch (err) {
        console.error('Error fetching customer orders:', err);
      }
    } else {
      setCustomerOrders([]);
      setShowOrderHistory(false);
    }
  };

  // Create new customer and select them
  const handleCustomerSubmit = async (e) => {
    e.preventDefault();
    try {
      const newCustomer = await api.createCustomer(customerData);
      setShowCustomerForm(false);
      setCustomerData({ name: '', phone: '', email: '', address: '' });
      
      // Select the new customer
      handleCustomerChange(newCustomer.id);
      
      // Refresh customers list
      const customersData = await api.getCustomers();
      setCustomers(customersData);
    } catch (err) {
      alert(err.message);
    }
  };

  // Get selected items (quantity > 0)
  const getSelectedItems = () => {
    return Object.entries(orderItems)
      .filter(([_, qty]) => qty > 0)
      .map(([itemId, quantity]) => {
        const item = items.find(i => i.id === parseInt(itemId));
        return {
          item_id: parseInt(itemId),
          quantity,
          price: item?.price || 0,
          name: item?.name || 'Unknown'
        };
      });
  };

  // Save order with updated items
  const handleSave = async () => {
    try {
      const selectedItems = getSelectedItems();
      
       // Convert datetime-local (in user's local timezone) to UTC for storage
       // datetime-local format: "2026-03-06T14:30" represents user's local time
       // We need to convert to UTC before sending to backend
       let scheduledDateISO = null;
       if (formData.scheduled_date) {
         // Parse the local datetime and convert to UTC
         const localDate = new Date(formData.scheduled_date);
         // Get UTC components
         const year = localDate.getUTCFullYear();
         const month = String(localDate.getUTCMonth() + 1).padStart(2, '0');
         const day = String(localDate.getUTCDate()).padStart(2, '0');
         const hours = String(localDate.getUTCHours()).padStart(2, '0');
         const minutes = String(localDate.getUTCMinutes()).padStart(2, '0');
         const seconds = '00';
         scheduledDateISO = `${year}-${month}-${day}T${hours}:${minutes}:${seconds}Z`;
       }
      
      // Update order with new data
      await api.updateOrder(parseInt(id), {
        customer_id: parseInt(formData.customer_id) || null,
        delivery_address: formData.delivery_address,
        notes: formData.notes,
        payment_method: formData.payment_method,
        total_amount: calculateTotal(),
        scheduled_date: scheduledDateISO,
        items: selectedItems // Backend will handle updating order_items table
      });
      
      alert('Order updated successfully!');
      navigate('/orders');
    } catch (err) {
      alert(err.message);
    }
  };

  // Cancel and go back to orders list
  const handleCancel = () => {
    navigate('/orders');
  };

  // Group items by category for display
  const groupedItems = items.reduce((acc, item) => {
    if (!acc[item.category]) {
      acc[item.category] = [];
    }
    acc[item.category].push(item);
    return acc;
  }, {});

  if (loading) return <div className="loading">Loading order...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Edit Order #{id}</h1>
        <div className="header-buttons">
          <button className="btn-secondary" onClick={handleCancel}>Cancel</button>
          <button className="btn-primary" onClick={handleSave}>Save Order</button>
        </div>
      </div>

      {/* Order Details Section */}
      <div className="form-section">
        <h2>Order Details</h2>
        
        <div className="form-group">
          <label>Customer</label>
          <select
            value={formData.customer_id}
            onChange={(e) => handleCustomerChange(e.target.value)}
          >
            <option value="">Select Customer</option>
            {customers.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          
          <button 
            type="button" 
            className="btn-secondary btn-small"
            onClick={() => setShowCustomerForm(!showCustomerForm)}
          >
            {showCustomerForm ? 'Cancel' : '+ New Customer'}
          </button>
        </div>

        {showCustomerForm && (
          <div className="nested-form">
            <h4>New Customer</h4>
            <input
              type="text"
              placeholder="Customer Name"
              value={customerData.name}
              onChange={(e) => setCustomerData({ ...customerData, name: e.target.value })}
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
            <button type="button" className="btn-primary" onClick={handleCustomerSubmit}>
              Add Customer
            </button>
          </div>
        )}

        {showOrderHistory && customerOrders.length > 0 && (
          <div className="customer-order-history">
            <button 
              type="button"
              className="collapsible"
              onClick={() => setShowOrderHistory(!showOrderHistory)}
            >
              Order History ({customerOrders.length} {customerOrders.length === 1 ? 'order' : 'orders'})
            </button>
            {showOrderHistory && (
              <div className="order-history-content">
                {customerOrders.map(order => (
                  <div key={order.id} className="history-order-item">
                    <div className="history-order-info">
                      <span className="order-id">Order #{order.id}</span>
                      <span className={`status-badge status-${order.status}`}>{order.status}</span>
                    </div>
                    {order.order_items && order.order_items.length > 0 && (
                      <div className="history-order-items">
                        {order.order_items.map(item => `${item.quantity}x ${item.item_name}`).join(', ')}
                      </div>
                    )}
                    <div className="history-order-details">
                      <span>{order.created_at ? new Date(order.created_at).toLocaleDateString() : 'N/A'}</span>
                      <span className="order-total">${order.total_amount ? order.total_amount.toFixed(2) : '0.00'}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        <div className="form-group">
          <label>Delivery Address</label>
          <input
            type="text"
            placeholder="Delivery Address"
            value={formData.delivery_address}
            onChange={(e) => setFormData({ ...formData, delivery_address: e.target.value })}
          />
        </div>

        <div className="form-group">
          <label>Notes</label>
          <input
            type="text"
            placeholder="Order Notes"
            value={formData.notes}
            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
          />
        </div>

        <div className="form-group">
          <label>Payment Method</label>
          <select
            value={formData.payment_method}
            onChange={(e) => setFormData({ ...formData, payment_method: e.target.value })}
          >
            <option value="cash">Cash</option>
            <option value="e-transfer">e-Transfer</option>
          </select>
        </div>

        <div className="form-group">
          <label>Scheduled Date & Time</label>
          <input
            type="datetime-local"
            value={formData.scheduled_date}
            onChange={(e) => setFormData({ ...formData, scheduled_date: e.target.value })}
          />
        </div>
      </div>

      {/* Current Order Items Section */}
      <div className="form-section">
        <h2>Order Items</h2>
        
        {Object.keys(orderItems).some(key => orderItems[key] > 0) ? (
          <div className="current-items">
            <h4>Items in Order:</h4>
            {getSelectedItems().map(selected => (
              <div key={selected.item_id} className="order-item-row">
                <span className="item-name">{selected.name}</span>
                <div className="quantity-controls">
                  <button 
                    type="button"
                    className="qty-btn"
                    onClick={() => handleQuantityChange(selected.item_id, selected.quantity - 1)}
                  >
                    -
                  </button>
                  <input
                    type="number"
                    min="0"
                    value={selected.quantity}
                    onChange={(e) => handleQuantityChange(selected.item_id, e.target.value)}
                    className="qty-input"
                  />
                  <button 
                    type="button"
                    className="qty-btn"
                    onClick={() => handleQuantityChange(selected.item_id, selected.quantity + 1)}
                  >
                    +
                  </button>
                </div>
                <span className="item-subtotal">
                  ${(selected.price * selected.quantity).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="no-items">No items in order yet. Add items from the menu below.</p>
        )}

        <div className="order-total">
          <strong>Total: ${calculateTotal().toFixed(2)}</strong>
        </div>
        <div className="order-payment-method">
          <strong>Payment Method:</strong> {formData.payment_method === 'e-transfer' ? 'e-Transfer' : 'Cash'}
        </div>
      </div>

      {/* Add Items Section - Grouped by Category */}
      <div className="form-section">
        <h2>Add Items to Order</h2>
        
        {Object.entries(groupedItems).map(([category, categoryItems]) => (
          <div key={category} className="category-section">
            <h3 className="category-title">{category}</h3>
            <div className="items-grid">
              {categoryItems.map((item) => (
                <div 
                  key={item.id} 
                  className={`menu-item-card ${orderItems[item.id] > 0 ? 'selected' : ''}`}
                >
                  <div className="item-info">
                    <span className="item-name">{item.name}</span>
                    <span className="item-price">${item.price.toFixed(2)}</span>
                  </div>
                  <div className="quantity-controls">
                    <button 
                      type="button"
                      className="qty-btn"
                      onClick={() => handleQuantityChange(item.id, (orderItems[item.id] || 0) - 1)}
                    >
                      -
                    </button>
                    <input
                      type="number"
                      min="0"
                      value={orderItems[item.id] || 0}
                      onChange={(e) => handleQuantityChange(item.id, e.target.value)}
                      className="qty-input"
                    />
                    <button 
                      type="button"
                      className="qty-btn"
                      onClick={() => handleQuantityChange(item.id, (orderItems[item.id] || 0) + 1)}
                    >
                      +
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default OrderEdit;
