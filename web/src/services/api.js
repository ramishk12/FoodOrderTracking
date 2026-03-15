const API_BASE = '/api';

async function request(endpoint, options = {}) {
  const config = {
    headers: {
      'Content-Type': 'application/json',
      'Cache-Control': 'no-cache',
      ...options.headers,
    },
    ...options,
  };

  if (config.body && typeof config.body === 'object') {
    config.body = JSON.stringify(config.body);
  }

  const method = config.method || 'GET';
  let url = `${API_BASE}${endpoint}`;
  if (method === 'GET') {
    const separator = endpoint.includes('?') ? '&' : '?';
    url += `${separator}t=${Date.now()}`;
  }
  
  const response = await fetch(url, config);
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || 'Request failed');
  }

  return response.json();
}

export const api = {
  getCustomers: () => request('/customers'),
  getCustomer: (id) => request(`/customers/${id}`),
  createCustomer: (data) => request('/customers', { method: 'POST', body: data }),
  updateCustomer: (id, data) => request(`/customers/${id}`, { method: 'PUT', body: data }),
  deleteCustomer: (id) => request(`/customers/${id}`, { method: 'DELETE' }),

  getItems: () => request('/items'),
  getItem: (id) => request(`/items/${id}`),
  createItem: (data) => request('/items', { method: 'POST', body: data }),
  updateItem: (id, data) => request(`/items/${id}`, { method: 'PUT', body: data }),
  deleteItem: (id) => request(`/items/${id}`, { method: 'DELETE' }),
  activateItem: (id) => request(`/items/${id}/activate`, { method: 'POST' }),

  getOrders: () => request('/orders'),
  getScheduledOrders: (days = 7) => request(`/orders/scheduled?days=${days}`),
  getOrdersByCustomer: (customerId) => request(`/orders/customer/${customerId}`),
  getOrder: (id) => request(`/orders/${id}`),
  createOrder: (data) => request('/orders', { method: 'POST', body: data }),
  updateOrder: (id, data) => request(`/orders/${id}`, { method: 'PUT', body: data }),
  deleteOrder: (id) => request(`/orders/${id}`, { method: 'DELETE' }),

  getDashboardStats: () => request('/dashboard'),

  // Item modifiers (item-scoped)
  getItemModifiers: (itemId) => request(`/items/${itemId}/modifiers`),
  createItemModifier: (itemId, data) => request(`/items/${itemId}/modifiers`, { method: 'POST', body: data }),
  updateItemModifier: (itemId, modifierId, data) => request(`/items/${itemId}/modifiers/${modifierId}`, { method: 'PUT', body: data }),
  deleteItemModifier: (itemId, modifierId) => request(`/items/${itemId}/modifiers/${modifierId}`, { method: 'DELETE' }),
};
