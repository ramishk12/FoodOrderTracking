import { useState, useEffect } from 'react';
import { api } from '../services/api';

function fmtUsd(n) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency', currency: 'USD',
    minimumFractionDigits: 2, maximumFractionDigits: 2,
  }).format(n ?? 0);
}

function fmtDate(iso) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  });
}

function groupOrderItems(orderItems) {
  const groups = [];
  const seen = new Map();

  for (const oi of orderItems) {
    const modKey = (oi.modifiers || [])
      .map((m) => m.modifier_id ?? m.modifier_name)
      .sort()
      .join(',');
    const key = `${oi.item_id}::${modKey}`;

    if (seen.has(key)) {
      groups[seen.get(key)].quantity += oi.quantity;
    } else {
      seen.set(key, groups.length);
      groups.push({
        item_id: oi.item_id,
        item_name: oi.item_name,
        quantity: oi.quantity,
        modifiers: oi.modifiers || [],
      });
    }
  }
  return groups;
}

function OrderHistoryRow({ order, variant = 'ord' }) {
  const prefix = variant === 'oe' ? 'oe-hist' : 'ord-hist';

  return (
    <div className={`${prefix}-row`}>
      <div className={`${prefix}-top`}>
        <span className={`${prefix}-id`}>#{order.id}</span>
        <span className={`${prefix}-date`}>{fmtDate(order.created_at)}</span>
      </div>
      <div className={`${prefix}-items`}>
        {order.order_items?.length > 0 ? (
          <div>
            {groupOrderItems(order.order_items)
              .sort((a, b) => (a.item_name || '').localeCompare(b.item_name || ''))
              .map((i, idx) => (
                <div key={idx}>
                  {i.quantity}× {i.item_name}
                  {i.modifiers?.length > 0 && (
                    <span className={`${prefix}-mod`}>
                      {' '}({i.modifiers.map((m) => m.modifier_name).join(', ')})
                    </span>
                  )}
                </div>
              ))}
          </div>
        ) : 'No items'}
      </div>
      <div className={`${prefix}-footer`}>
        <span className={`${prefix}-total`}>{fmtUsd(order.total_amount)}</span>
        <span className={`${prefix}-status`}>{order.status}</span>
      </div>
    </div>
  );
}

export default function CustomerOrderHistory({ customerId, variant = 'ord', showTitle = true, excludeId = null }) {
  const [orderHistory, setOrderHistory] = useState([]);
  const [loading, setLoading] = useState(false);
  const prefix = variant === 'oe' ? 'oe-hist' : 'ord-hist';

  useEffect(() => {
    if (!customerId) {
      setOrderHistory([]);
      return;
    }

    const fetchHistory = async () => {
      setLoading(true);
      try {
        const orders = await api.getOrdersByCustomer(customerId);
        const filtered = excludeId 
          ? (orders || []).filter((o) => o.id !== excludeId)
          : (orders || []);
        setOrderHistory(filtered);
      } catch {
        setOrderHistory([]);
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
  }, [customerId]);

  if (!customerId) return null;

  return (
    <div className={`${prefix}`}>
      {showTitle && (
        <div className={`${prefix}-title`}>
          Order history
          {variant === 'oe' && orderHistory.length > 0 && (
            <> · {orderHistory.length} previous order{orderHistory.length !== 1 ? 's' : ''}</>
          )}
        </div>
      )}
      {loading ? (
        <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--muted)' }}>
          Loading…
        </div>
      ) : orderHistory.length === 0 ? (
        <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--rule)', fontStyle: 'italic' }}>
          No previous orders
        </div>
      ) : (
        orderHistory.map((order) => (
          <OrderHistoryRow key={order.id} order={order} variant={variant} />
        ))
      )}
    </div>
  );
}
