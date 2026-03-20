import { useLocation } from 'react-router-dom';

export function PlaceholderPage() {
  const location = useLocation();
  const name = location.pathname.slice(1) || 'page';
  const title = name.charAt(0).toUpperCase() + name.slice(1);

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">{title}</h1>
      <p className="text-gray-500">This page will be built in Phase 1.3.</p>
    </div>
  );
}
