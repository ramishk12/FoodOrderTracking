import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom'
import Home from './pages/Home'
import Dashboard from './pages/Dashboard'
import Schedule from './pages/Schedule'
import Orders from './pages/Orders'
import OrderEdit from './pages/OrderEdit'
import Customers from './pages/Customers'
import Items from './pages/Items'
import './index.css'

const LINKS = [
  { path: '/',          label: 'Home' },
  { path: '/dashboard', label: 'Dashboard' },
  { path: '/schedule',  label: 'Schedule' },
  { path: '/items',     label: 'Menu' },
  { path: '/orders',    label: 'Orders' },
  { path: '/customers', label: 'Customers' },
]

function OrderEditWrapper() {
  const location = useLocation()
  return <OrderEdit key={location.pathname} />
}

function Navbar() {
  const { pathname } = useLocation()

  const isActive = (path) => path === '/' ? pathname === '/' : pathname.startsWith(path)

  return (
    <nav className="nav">
      <Link className="nav-brand" to="/">
        <span className="nav-brand-word">Food <em>Order</em></span>
        <span className="nav-brand-dot" />
        <span className="nav-brand-word">Tracking</span>
      </Link>

      <div className="nav-divider" />

      <div className="nav-links">
        {LINKS.map(({ path, label }) => (
          <Link
            key={path}
            to={path}
            className={`nav-link${isActive(path) ? ' active' : ''}`}
            onClick={isActive(path) ? () => window.location.reload() : undefined}
          >
            {label}
          </Link>
        ))}
      </div>
    </nav>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <main className="app-main">
        <Routes>
          <Route path="/"                element={<Home />} />
          <Route path="/dashboard"       element={<Dashboard />} />
          <Route path="/schedule"        element={<Schedule />} />
          <Route path="/items"           element={<Items />} />
          <Route path="/orders"          element={<Orders />} />
          <Route path="/orders/:id/edit" element={<OrderEditWrapper />} />
          <Route path="/customers"       element={<Customers />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}