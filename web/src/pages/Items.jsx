import { useState, useEffect } from 'react';
import { api } from '../services/api';

function Items() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [quantities, setQuantities] = useState({});
  const [customers, setCustomers] = useState([]);
  const [showOrderForm, setShowOrderForm] = useState(false);
  const [showItemForm, setShowItemForm] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [orderForm, setOrderForm] = useState({
    customer_id: '',
    delivery_address: '',
    notes: '',
    scheduled_date: ''
  });
  const [itemForm, setItemForm] = useState({
    name: '',
    description: '',
    price: '',
    category: ''
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

  const handleCustomerChange = (customerId) => {
    const customer = customers.find(c => c.id === parseInt(customerId));
    setOrderForm(prev => ({
      ...prev,
      customer_id: customerId,
      delivery_address: customer?.address || prev.delivery_address
    }));
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
      // Convert datetime-local to ISO string treating it as UTC
      // datetime-local format: "2026-03-05T10:30" should be treated as "2026-03-05T10:30Z"
      let scheduledDateISO = null;
      if (orderForm.scheduled_date) {
        scheduledDateISO = orderForm.scheduled_date + 'Z';
      }
      
      await api.createOrder({
        customer_id: parseInt(orderForm.customer_id) || null,
        delivery_address: orderForm.delivery_address,
        notes: orderForm.notes,
        scheduled_date: scheduledDateISO,
        items: orderItems
      });
      alert('Order created successfully!');
      setShowOrderForm(false);
      setQuantities(prev => {
        const reset = { ...prev };
        Object.keys(reset).forEach(key => reset[key] = 0);
        return reset;
      });
      setOrderForm({ customer_id: '', delivery_address: '', notes: '', scheduled_date: '' });
    } catch (err) {
      alert(err.message);
    }
  };

  const handleItemSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingItem) {
        await api.updateItem(editingItem.id, {
          ...itemForm,
          price: parseFloat(itemForm.price),
          available: true
        });
      } else {
        await api.createItem({
          ...itemForm,
          price: parseFloat(itemForm.price),
          available: true
        });
      }
      setShowItemForm(false);
      setEditingItem(null);
      setItemForm({ name: '', description: '', price: '', category: '' });
      loadData();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleEditItem = (item) => {
    setItemForm({
      name: item.name,
      description: item.description || '',
      price: String(item.price),
      category: item.category || ''
    });
    setEditingItem(item);
    setShowItemForm(true);
  };

  const handleDeleteItem = async (id) => {
    if (!confirm('Delete this item?')) return;
    try {
      await api.deleteItem(id);
      loadData();
    } catch (err) {
      alert(err.message);
    }
  };

  const handleCancelItem = () => {
    setShowItemForm(false);
    setEditingItem(null);
    setItemForm({ name: '', description: '', price: '', category: '' });
  };

  const filteredItems = items.filter(item => {
    const search = searchTerm.toLowerCase();
    const matchesSearch = !searchTerm || 
      item.name?.toLowerCase().includes(search) ||
      item.description?.toLowerCase().includes(search);
    const matchesCategory = !categoryFilter || item.category === categoryFilter;
    return matchesSearch && matchesCategory;
  });

  const groupedItems = filteredItems.reduce((acc, item) => {
    if (!acc[item.category]) {
      acc[item.category] = [];
    }
    acc[item.category].push(item);
    return acc;
  }, {});

  const categories = [...new Set(items.map(item => item.category))];

  if (loading) return <div className="loading">Loading items...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  return (
    <div className="page">
      <div className="page-header">
        <h1>Menu Items</h1>
        <div className="header-buttons">
          <button 
            className="btn-secondary"
            onClick={() => setShowItemForm(!showItemForm)}
          >
            {showItemForm ? 'Cancel' : '+ Add Item'}
          </button>
          <button 
            className="btn-primary" 
            onClick={() => setShowOrderForm(!showOrderForm)}
            disabled={getTotalAmount() === 0}
          >
            {showOrderForm ? 'Cancel' : `Order ($${getTotalAmount().toFixed(2)})`}
          </button>
        </div>
      </div>

      <div className="filters">
        <input
          type="text"
          placeholder="Search items..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="search-input"
        />
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="filter-select"
        >
          <option value="">All Categories</option>
          {categories.map(cat => (
            <option key={cat} value={cat}>{cat}</option>
          ))}
        </select>
      </div>

      {showItemForm && (
        <form onSubmit={handleItemSubmit} className="form">
          <h3>{editingItem ? 'Edit Item' : 'Add New Item'}</h3>
          <input
            type="text"
            placeholder="Item Name"
            value={itemForm.name}
            onChange={(e) => setItemForm({ ...itemForm, name: e.target.value })}
            required
          />
          <input
            type="text"
            placeholder="Description"
            value={itemForm.description}
            onChange={(e) => setItemForm({ ...itemForm, description: e.target.value })}
          />
          <input
            type="number"
            placeholder="Price"
            step="0.01"
            value={itemForm.price}
            onChange={(e) => setItemForm({ ...itemForm, price: e.target.value })}
            required
          />
          <input
            type="text"
            placeholder="Category"
            value={itemForm.category}
            onChange={(e) => setItemForm({ ...itemForm, category: e.target.value })}
            required
          />
          <div className="form-actions">
            <button type="submit" className="btn-primary">
              {editingItem ? 'Update Item' : 'Add Item'}
            </button>
            {editingItem && (
              <button type="button" className="btn-secondary" onClick={handleCancelItem}>
                Cancel
              </button>
            )}
          </div>
        </form>
      )}

      {showOrderForm && (
        <form onSubmit={handleCreateOrder} className="form">
          <h3>Create Order</h3>
          <select
            value={orderForm.customer_id}
            onChange={(e) => handleCustomerChange(e.target.value)}
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
          />
          <input
            type="text"
            placeholder="Notes"
            value={orderForm.notes}
            onChange={(e) => setOrderForm({ ...orderForm, notes: e.target.value })}
          />
          <input
            type="datetime-local"
            value={orderForm.scheduled_date}
            onChange={(e) => setOrderForm({ ...orderForm, scheduled_date: e.target.value })}
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
                <div className="card-actions">
                  <button className="btn-primary" onClick={() => handleEditItem(item)}>Edit</button>
                  <button className="btn-danger" onClick={() => handleDeleteItem(item.id)}>Delete</button>
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
