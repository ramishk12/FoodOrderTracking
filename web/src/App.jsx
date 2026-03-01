import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom'
import Home from './pages/Home'
import Orders from './pages/Orders'
import Customers from './pages/Customers'

function Navbar() {
  const location = useLocation()
  
  const links = [
    { path: '/', label: 'Home' },
    { path: '/orders', label: 'Orders' },
    { path: '/customers', label: 'Customers' },
  ]

  return (
    <nav className="navbar">
      <div className="nav-brand">Food Order Tracking</div>
      <div className="nav-links">
        {links.map((link) => (
          <Link
            key={link.path}
            to={link.path}
            className={location.pathname === link.path ? 'active' : ''}
          >
            {link.label}
          </Link>
        ))}
      </div>
    </nav>
  )
}

function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/orders" element={<Orders />} />
          <Route path="/customers" element={<Customers />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

export default App
