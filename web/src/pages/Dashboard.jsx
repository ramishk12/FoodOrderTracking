import { useState, useEffect } from 'react';
import { api } from '../services/api';

function Dashboard() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await api.getDashboardStats();
      setStats(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="loading">Loading dashboard...</div>;
  if (error) return <div className="error">Error: {error}</div>;

  const maxRevenue = Math.max(...(stats?.sales_trend?.map(d => d.revenue) || [1]), 1);

  return (
    <div className="page">
      <div className="page-header">
        <h1>Sales Dashboard</h1>
      </div>

      <div className="dashboard-grid">
        <div className="stat-card">
          <h3>Total Revenue</h3>
          <p className="stat-value">${stats.total_revenue?.toFixed(2)}</p>
          <p className="stat-label">All Time</p>
        </div>

        <div className="stat-card">
          <h3>Monthly Revenue</h3>
          <p className="stat-value">${stats.monthly_revenue?.toFixed(2)}</p>
          <p className="stat-label">This Month</p>
        </div>

        <div className="stat-card">
          <h3>Daily Revenue</h3>
          <p className="stat-value">${stats.daily_revenue?.toFixed(2)}</p>
          <p className="stat-label">Today</p>
        </div>

        <div className="stat-card">
          <h3>Total Orders</h3>
          <p className="stat-value">{stats.total_orders}</p>
          <p className="stat-label">All Time</p>
        </div>

        <div className="stat-card">
          <h3>Monthly Orders</h3>
          <p className="stat-value">{stats.monthly_orders}</p>
          <p className="stat-label">This Month</p>
        </div>

        <div className="stat-card">
          <h3>Daily Orders</h3>
          <p className="stat-value">{stats.daily_orders}</p>
          <p className="stat-label">Today</p>
        </div>

        <div className="stat-card">
          <h3>Average Order Value</h3>
          <p className="stat-value">${stats.average_order_value?.toFixed(2)}</p>
          <p className="stat-label">Per Order</p>
        </div>

        <div className="stat-card">
          <h3>Best Selling Item</h3>
          <p className="stat-value">{stats.best_selling_items?.[0]?.name || 'N/A'}</p>
          <p className="stat-label">{stats.best_selling_items?.[0]?.quantity || 0} sold</p>
        </div>
      </div>

      <div className="dashboard-sections">
        <div className="dashboard-section">
          <h2>Orders by Status</h2>
          <div className="status-breakdown">
            {Object.entries(stats.orders_by_status || {}).filter(([status]) => status && status.length > 0).map(([status, count]) => (
              <div key={status} className="status-row">
                <span className={`status-badge status-${status}`}>{status}</span>
                <span className="status-count">{count}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="dashboard-section">
          <h2>Best Selling Items</h2>
          <div className="items-list">
            {(stats.best_selling_items || []).map((item, index) => (
              <div key={index} className="item-row">
                <span className="item-rank">#{index + 1}</span>
                <span className="item-name">{item.name}</span>
                <span className="item-qty">{item.quantity} sold</span>
                <span className="item-revenue">${item.revenue?.toFixed(2)}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="dashboard-section">
          <h2>Top Customers</h2>
          <div className="customers-list">
            {(stats.top_customers || []).map((customer, index) => (
              <div key={index} className="customer-row">
                <span className="customer-rank">#{index + 1}</span>
                <span className="customer-name">{customer.name}</span>
                <span className="customer-orders">{customer.order_count} orders</span>
                <span className="customer-spent">${customer.total_spent?.toFixed(2)}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="dashboard-section full-width">
          <h2>Sales Trend (Last 30 Days)</h2>
          <div className="sales-chart">
            {(stats.sales_trend || []).map((point, index) => (
              <div key={index} className="chart-bar-container">
                <div 
                  className="chart-bar" 
                  style={{ height: `${(point.revenue / maxRevenue) * 100}%` }}
                  title={`${point.date}: $${point.revenue?.toFixed(2)} (${point.orders} orders)`}
                ></div>
                <span className="chart-label">{new Date(point.date).getDate()}</span>
              </div>
            ))}
          </div>
          <div className="chart-legend">
            <span>30 days ago</span>
            <span>Today</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
