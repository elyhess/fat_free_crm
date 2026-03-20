import { useAuth } from '../auth/AuthContext';
import { useApi } from '../hooks/useApi';
import type { TaskSummary, PipelineResponse } from '../types/entities';

const BUCKET_LABELS: Record<string, string> = {
  due_asap: 'ASAP',
  overdue: 'Overdue',
  due_today: 'Today',
  due_tomorrow: 'Tomorrow',
  due_this_week: 'This Week',
  due_next_week: 'Next Week',
  due_later: 'Later',
};

function formatCurrency(val: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
  }).format(val);
}

export function DashboardPage() {
  const { user } = useAuth();
  const { data: tasks } = useApi<TaskSummary>('/dashboard/tasks');
  const { data: pipeline } = useApi<PipelineResponse>('/dashboard/pipeline');

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h1>

      {/* Welcome + Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
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
            Pending Tasks
          </h2>
          <p className="mt-2 text-3xl font-semibold text-gray-900">
            {tasks?.total_tasks ?? '\u2014'}
          </p>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
            Pipeline
          </h2>
          <p className="mt-2 text-3xl font-semibold text-gray-900">
            {pipeline ? formatCurrency(pipeline.total_weighted) : '\u2014'}
          </p>
          <p className="mt-1 text-sm text-gray-500">
            {pipeline ? `${pipeline.total_count} deals` : ''}
          </p>
        </div>
      </div>

      {/* Task Buckets */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Tasks by Due Date</h2>
          {tasks?.buckets ? (
            <div className="space-y-3">
              {tasks.buckets
                .filter((b) => b.count > 0)
                .map((b) => (
                  <div key={b.bucket} className="flex items-center justify-between">
                    <span className="text-sm text-gray-700">
                      {BUCKET_LABELS[b.bucket] ?? b.bucket}
                    </span>
                    <span
                      className={`text-sm font-medium px-2 py-0.5 rounded ${
                        b.bucket === 'overdue'
                          ? 'bg-red-100 text-red-700'
                          : b.bucket === 'due_asap'
                            ? 'bg-orange-100 text-orange-700'
                            : 'bg-blue-100 text-blue-700'
                      }`}
                    >
                      {b.count}
                    </span>
                  </div>
                ))}
              {tasks.buckets.every((b) => b.count === 0) && (
                <p className="text-sm text-gray-500">No pending tasks.</p>
              )}
            </div>
          ) : (
            <p className="text-sm text-gray-500">Loading...</p>
          )}
        </div>

        {/* Pipeline Stages */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">
            Opportunity Pipeline
          </h2>
          {pipeline?.stages ? (
            <div className="space-y-3">
              {pipeline.stages.length > 0 ? (
                <>
                  {pipeline.stages.map((s) => (
                    <div key={s.stage} className="flex items-center justify-between">
                      <span className="text-sm text-gray-700 capitalize">
                        {s.stage ?? 'Unset'}
                      </span>
                      <div className="text-right">
                        <span className="text-sm font-medium text-gray-900">
                          {formatCurrency(s.total_amount)}
                        </span>
                        <span className="text-xs text-gray-500 ml-2">
                          ({s.count} deal{s.count !== 1 ? 's' : ''})
                        </span>
                      </div>
                    </div>
                  ))}
                  <div className="border-t pt-3 flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-900">
                      Total (weighted)
                    </span>
                    <span className="text-sm font-semibold text-gray-900">
                      {formatCurrency(pipeline.total_weighted)}
                    </span>
                  </div>
                </>
              ) : (
                <p className="text-sm text-gray-500">No open opportunities.</p>
              )}
            </div>
          ) : (
            <p className="text-sm text-gray-500">Loading...</p>
          )}
        </div>
      </div>
    </div>
  );
}
