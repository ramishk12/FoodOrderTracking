import { useState, useEffect } from 'react';
import { api } from '../services/api';

function Items() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [quantities, setQuantities] = useState({});
  const [customers, setCustomers] = useState([]);
  const [showOrderForm, setShowOrderForm] = useState(false);
  const [orderForm, setOrderForm] = useState({
    customer_id: '',
    delivery_address: '',
    notes: ''
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [itemsData, customersData] = await Promise.all([
        api.getItems(),
        api.getCustomers()
      ]);
      setItems(itemsData);
      setCustomers(customersData);
      const initialQuantities = {};
      itemsData.forEach(item => {
        initialQuantities[item.id] = 0;
      });
      setQuantities(initialQuantities);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleQuantityChange = (itemId, value) => {
    const qty = Math.max(0, parseInt(value) || 0);
    setQuantities(prev => ({ ...prev, [itemId]: qty }));
  };

  const getTotalAmount = () => {
    return items.reduce((total, item) => {
      return total + (item.price * (quantities[item.id] || 0));
    }, 0);
  };

  const getSelectedItems = () => {
    return items.filter(item => quantities[item.id] > 0);
  };

  const handleCreateOrder = async (e) => {
    e.preventDefault();
    const selectedItems = getSelectedItems();
    if (selectedItems.length === 0) {
      alert('Please select at least one item');
      return;
    }

    const orderItems = selectedItems.map(item => ({
      item_id: item.id,
      quantity: quantities[item.id]
    }));

    try {
      await api.createOrder({
        customer_id: parseInt(orderForm.customer_id) || null,
        delivery_address: orderForm.delivery_address,
        notes: orderForm.notes,
        items: orderItems
      });
      alert('Order created successfully!');
      setShowOrderForm(false);
      setQuantities(prev => {
        const reset = { ...prev };
        Object.keys(reset).forEach(key => reset[key] = 0);
        return reset;
      });
      setOrderForm({ customer_id: '', delivery_address: '', notes: '' });
    } catch (err) {
      alert(err.message);
    }
  };

  const groupedItems = items.reduce((acc, item) => {
    if (!acc[item.category]) {
      acc[item.category] = [];
    }
    acc[item.category].push(item);
    return acc;
  }, {});

  if (loading) return <div className="loading">Loading items...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Menu Items</h1>
        <button 
          className="btn-primary" 
          onClick={() => setShowOrderForm(!showOrderForm)}
          disabled={getTotalAmount() === 0}
        >
          {showOrderForm ? 'Cancel' : `Order ($${getTotalAmount().toFixed(2)})`}
        </button>
      </div>

      {showOrderForm && (
        <form onSubmit={handleCreateOrder} className="form">
          <h3>Create Order</h3>
          <select
            value={orderForm.customer_id}
            onChange={(e) => setOrderForm({ ...orderForm, customer_id: e.target.value })}
          >
            <option value="">Select Customer</option>
            {customers.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          <input
            type="text"
            placeholder="Delivery Address"
            value={orderForm.delivery_address}
            onChange={(e) => setOrderForm({ ...orderForm, delivery_address: e.target.value })}
            required
          />
          <input
            type="text"
            placeholder="Notes"
            value={orderForm.notes}
            onChange={(e) => setOrderForm({ ...orderForm, notes: e.target.value })}
          />
          <div className="order-summary">
            <h4>Order Summary:</h4>
            {getSelectedItems().map(item => (
              <p key={item.id}>
                {quantities[item.id]}x {item.name} - ${(item.price * quantities[item.id]).toFixed(2)}
              </p>
            ))}
            <p><strong>Total: ${getTotalAmount().toFixed(2)}</strong></p>
          </div>
          <button type="submit" className="btn-primary">Place Order</button>
        </form>
      )}

      {Object.entries(groupedItems).map(([category, categoryItems]) => (
        <div key={category} className="category-section">
          <h2 className="category-title">{category}</h2>
          <div className="card-grid">
            {categoryItems.map((item) => (
              <div key={item.id} className="card">
                <div className="card-header">
                  <h3>{item.name}</h3>
                  <span className="item-price">${item.price.toFixed(2)}</span>
                </div>
                <div className="card-body">
                  <p>{item.description}</p>
                </div>
                <div className="card-actions quantity-selector">
                  <button 
                    className="qty-btn"
                    onClick={() => handleQuantityChange(item.id, (quantities[item.id] || 0) - 1)}
                  >
                    -
                  </button>
                  <input
                    type="number"
                    min="0"
                    value={quantities[item.id] || 0}
                    onChange={(e) => handleQuantityChange(item.id, e.target.value)}
                    className="qty-input"
                  />
                  <button 
                    className="qty-btn"
                    onClick={() => handleQuantityChange(item.id, (quantities[item.id] || 0) + 1)}
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
  );
}

export default Items;
