import { Link } from 'react-router-dom'

function Home() {
  return (
    <div className="home">
      <h1>Food Order Tracking</h1>
      <p>Track and manage your food delivery orders</p>
      
      <div className="home-links">
        <Link to="/orders" className="home-card">
          <h3>📦 Orders</h3>
          <p>View and manage all orders</p>
        </Link>
        <Link to="/customers" className="home-card">
          <h3>👥 Customers</h3>
          <p>View and manage customers</p>
        </Link>
      </div>
    </div>
  )
}

export default Home
