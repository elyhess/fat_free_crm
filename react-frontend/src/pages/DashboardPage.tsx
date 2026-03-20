import { useAuth } from '../auth/AuthContext';

export function DashboardPage() {
  const { user } = useAuth();

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
            Welcome
          </h2>
          <p className="mt-2 text-lg text-gray-900">
            {user?.first_name} {user?.last_name}
          </p>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
            Tasks
          </h2>
          <p className="mt-2 text-3xl font-semibold text-gray-900">&mdash;</p>
          <p className="mt-1 text-sm text-gray-500">Coming soon</p>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
            Opportunities
          </h2>
          <p className="mt-2 text-3xl font-semibold text-gray-900">&mdash;</p>
          <p className="mt-1 text-sm text-gray-500">Coming soon</p>
        </div>
      </div>
    </div>
  );
}
