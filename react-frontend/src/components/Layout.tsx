import { useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

const navItems = [
  { label: 'Dashboard', path: '/' },
  { label: 'Tasks', path: '/tasks' },
  { label: 'Campaigns', path: '/campaigns' },
  { label: 'Leads', path: '/leads' },
  { label: 'Accounts', path: '/accounts' },
  { label: 'Contacts', path: '/contacts' },
  { label: 'Opportunities', path: '/opportunities' },
];

export function Layout() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    const q = searchQuery.trim();
    if (q) {
      navigate(`/search?q=${encodeURIComponent(q)}`);
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Top bar */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4">
          {/* Upper row: brand, search, user */}
          <div className="flex items-center justify-between h-12">
            <span className="text-lg font-semibold text-gray-900 shrink-0">
              Fat Free CRM
            </span>
            <div className="flex items-center gap-3 min-w-0">
              <form onSubmit={handleSearch} className="hidden sm:block">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search..."
                  className="w-44 px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </form>
              {user?.admin && (
                <div className="hidden lg:flex items-center gap-3">
                  <Link to="/admin/fields" className="text-xs text-gray-500 hover:text-gray-900 whitespace-nowrap">Fields</Link>
                  <Link to="/admin/research-tools" className="text-xs text-gray-500 hover:text-gray-900 whitespace-nowrap">Research Tools</Link>
                  <Link to="/admin/settings" className="text-xs text-gray-500 hover:text-gray-900 whitespace-nowrap">Settings</Link>
                </div>
              )}
              <Link to="/profile" className="flex items-center gap-1.5 text-sm text-gray-600 hover:text-gray-900 shrink-0">
                <img
                  src={user?.avatar_url || `/api/v1/avatars/${user?.id}`}
                  alt=""
                  className="w-6 h-6 rounded-full object-cover bg-gray-200"
                  onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                />
                <span className="hidden sm:inline max-w-[80px] truncate">{user?.username}</span>
                {user?.admin && (
                  <span className="text-xs bg-blue-100 text-blue-700 px-1 py-0.5 rounded hidden sm:inline">
                    admin
                  </span>
                )}
              </Link>
              <button
                onClick={logout}
                className="text-sm text-gray-500 hover:text-gray-700 shrink-0"
              >
                Sign out
              </button>
            </div>
          </div>
          {/* Lower row: navigation links */}
          <div className="flex items-center gap-1 overflow-x-auto pb-2 -mb-px scrollbar-none">
            {navItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`px-3 py-1.5 rounded-md text-sm font-medium whitespace-nowrap ${
                  location.pathname === item.path
                    ? 'bg-gray-100 text-gray-900'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
              >
                {item.label}
              </Link>
            ))}
            {user?.admin && (
              <div className="flex items-center gap-1 lg:hidden ml-2 pl-2 border-l border-gray-200">
                <Link to="/admin/fields" className="px-2 py-1.5 rounded-md text-xs font-medium text-gray-500 hover:text-gray-900 hover:bg-gray-50 whitespace-nowrap">Fields</Link>
                <Link to="/admin/research-tools" className="px-2 py-1.5 rounded-md text-xs font-medium text-gray-500 hover:text-gray-900 hover:bg-gray-50 whitespace-nowrap">Research Tools</Link>
                <Link to="/admin/settings" className="px-2 py-1.5 rounded-md text-xs font-medium text-gray-500 hover:text-gray-900 hover:bg-gray-50 whitespace-nowrap">Settings</Link>
              </div>
            )}
          </div>
        </div>
      </nav>

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
