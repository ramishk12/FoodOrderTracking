import { Link } from 'react-router-dom';
import '../index.css';

const CARDS = [
  {
    to: '/dashboard',
    title: 'Dashboard',
    desc: 'Revenue trends, best-selling items, top customers, and a full 30-day sales picture.',
    tag: 'Analytics',
    featured: true,
    icon: (
      <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/>
        <rect x="3" y="14" width="7" height="7" rx="1"/>
        <path d="M14 17.5h7M17.5 14v7"/>
      </svg>
    ),
  },
  {
    to: '/orders',
    title: 'Orders',
    desc: 'View, search, and manage all orders by status or payment method.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2"/>
        <rect x="9" y="3" width="6" height="4" rx="1"/>
        <path d="M9 12h6M9 16h4"/>
      </svg>
    ),
  },
  {
    to: '/schedule',
    title: 'Schedule',
    desc: 'Upcoming orders grouped by date — today, tomorrow, and this week.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="4" width="18" height="18" rx="2"/>
        <path d="M16 2v4M8 2v4M3 10h18"/>
        <path d="M8 14h.01M12 14h.01M16 14h.01M8 18h.01M12 18h.01"/>
      </svg>
    ),
  },
  {
    to: '/items',
    title: 'Menu',
    desc: 'Browse, create, and toggle availability for menu items by category.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 6h18M3 12h18M3 18h11"/>
        <circle cx="19" cy="18" r="2"/>
        <path d="M19 16v-4"/>
      </svg>
    ),
  },
  {
    to: '/customers',
    title: 'Customers',
    desc: 'Customer profiles, contact details, and full order history.',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="9" cy="7" r="4"/>
        <path d="M3 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/>
        <path d="M16 3.13a4 4 0 0 1 0 7.75M21 21v-2a4 4 0 0 0-3-3.87"/>
      </svg>
    ),
  },
];

export default function Home() {
  return (
    <>
      <div className="home-root">

        {/* Hero */}
        <div className="home-hero">
          <div className="home-rule-top">
            <span className="home-rule-ornament">✦ ✦ ✦</span>
          </div>

          <div className="home-eyebrow">Est. Food Order Tracking System</div>

          <h1 className="home-title">
            Order<br /><em>Management</em>
          </h1>

          <p className="home-subtitle">
            Track deliveries, manage customers, and read your sales.
          </p>

          <div className="home-rule-bottom">
            <span>Select a section to begin</span>
          </div>
        </div>

        {/* Navigation cards */}
        <nav className="home-nav">
          {CARDS.map((card, i) => (
            <Link
              key={card.to}
              to={card.to}
              className={`nav-card${card.featured ? ' nav-card-dashboard' : ''}`}
              style={{ animationDelay: `${0.2 + i * 0.07}s` }}
            >
              <div className="nav-card-inner">
                <div className="nav-icon">{card.icon}</div>
                <div className="nav-text">
                  <div className="nav-card-title">
                    {card.title}
                    <span className="nav-card-title-arrow">→</span>
                  </div>
                  <div className="nav-card-desc">{card.desc}</div>
                </div>
                {card.tag && <span className="nav-card-tag">{card.tag}</span>}
              </div>
            </Link>
          ))}
        </nav>

        {/* Footer */}
        <div className="home-footer">
          FOOD ORDER TRACKING &nbsp;·&nbsp; ALL SECTIONS
        </div>

      </div>
    </>
  );
}