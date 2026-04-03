import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { useApi } from '../hooks/useApi';
import type { TaskSummary, PipelineResponse } from '../types/entities';

interface ActivityVersion {
  id: number;
  item_type: string;
  item_id: number;
  event: string;
  whodunnit?: string;
  created_at: string;
}

interface TaskItem {
  id: number;
  name: string;
  bucket?: string;
  category?: string;
  due_at?: string;
}

const ENTITY_SLUG: Record<string, string> = {
  Account: 'accounts',
  Contact: 'contacts',
  Lead: 'leads',
  Opportunity: 'opportunities',
  Campaign: 'campaigns',
};

const EVENT_COLORS: Record<string, string> = {
  create: 'bg-green-100 text-green-700',
  update: 'bg-blue-100 text-blue-700',
  destroy: 'bg-red-100 text-red-700',
};

const BUCKET_LABELS: Record<string, string> = {
  due_asap: 'ASAP',
  overdue: 'Overdue',
  due_today: 'Today',
  due_tomorrow: 'Tomorrow',
  due_this_week: 'This Week',
  due_next_week: 'Next Week',
  due_later: 'Later',
};

const STAGE_COLORS: Record<string, string> = {
  prospecting: 'bg-gray-400',
  analysis: 'bg-blue-400',
  presentation: 'bg-indigo-400',
  proposal: 'bg-purple-400',
  negotiation: 'bg-yellow-400',
  final_review: 'bg-orange-400',
  won: 'bg-green-500',
  lost: 'bg-red-400',
};

function formatCurrency(val: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
  }).format(val);
}

function useLocalStorage<T>(key: string, initial: T): [T, (v: T) => void] {
  const [value, setValue] = useState<T>(() => {
    try {
      const stored = localStorage.getItem(key);
      return stored ? JSON.parse(stored) : initial;
    } catch {
      return initial;
    }
  });
  function set(v: T) {
    setValue(v);
    localStorage.setItem(key, JSON.stringify(v));
  }
  return [value, set];
}

function CollapsibleSection({ title, storageKey, children }: { title: string; storageKey: string; children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useLocalStorage(`dash_${storageKey}`, false);
  return (
    <div className="bg-white shadow rounded-lg">
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="w-full flex items-center justify-between px-6 py-4 text-left"
      >
        <h2 className="text-lg font-medium text-gray-900">{title}</h2>
        <span className="text-gray-400 text-sm">{collapsed ? '+' : '-'}</span>
      </button>
      {!collapsed && <div className="px-6 pb-6">{children}</div>}
    </div>
  );
}

export function DashboardPage() {
  const { user } = useAuth();
  const { data: tasks } = useApi<TaskSummary>('/dashboard/tasks');
  const { data: pipeline } = useApi<PipelineResponse>('/dashboard/pipeline');
  const { data: activity } = useApi<ActivityVersion[]>('/activity?limit=20');
  const { data: taskItems } = useApi<TaskItem[]>('/tasks?per_page=10&sort=due_at&order=asc');
  const [taskView, setTaskView] = useLocalStorage<'summary' | 'list'>('dash_task_view', 'summary');

  // Compute max stage amount for bar chart scaling
  const maxStageAmount = pipeline?.stages
    ? Math.max(...pipeline.stages.map((s) => s.total_amount), 1)
    : 1;

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

      {/* Tasks and Pipeline */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <CollapsibleSection title="Tasks" storageKey="tasks">
          {/* View toggle */}
          <div className="flex gap-1 mb-4">
            <button
              onClick={() => setTaskView('summary')}
              className={`px-2.5 py-1 text-xs rounded ${
                taskView === 'summary' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              Summary
            </button>
            <button
              onClick={() => setTaskView('list')}
              className={`px-2.5 py-1 text-xs rounded ${
                taskView === 'list' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              List
            </button>
          </div>

          {taskView === 'summary' ? (
            // Summary view — bucket counts
            tasks?.buckets ? (
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
            )
          ) : (
            // List view — actual task items
            <div className="space-y-2">
              {taskItems && (taskItems as unknown as { data?: TaskItem[] }).data
                ? ((taskItems as unknown as { data: TaskItem[] }).data).map((t: TaskItem) => (
                    <Link
                      key={t.id}
                      to={`/tasks/${t.id}`}
                      className="flex items-center justify-between p-2 rounded hover:bg-gray-50 -mx-2"
                    >
                      <div className="min-w-0">
                        <p className="text-sm text-gray-900 truncate">{t.name}</p>
                        {t.category && <span className="text-xs text-gray-500">{t.category}</span>}
                      </div>
                      {t.bucket && (
                        <span className="text-xs text-gray-400 shrink-0 ml-2">
                          {BUCKET_LABELS[t.bucket] ?? t.bucket}
                        </span>
                      )}
                    </Link>
                  ))
                : taskItems && Array.isArray(taskItems)
                  ? (taskItems as TaskItem[]).map((t: TaskItem) => (
                      <Link
                        key={t.id}
                        to={`/tasks/${t.id}`}
                        className="flex items-center justify-between p-2 rounded hover:bg-gray-50 -mx-2"
                      >
                        <div className="min-w-0">
                          <p className="text-sm text-gray-900 truncate">{t.name}</p>
                          {t.category && <span className="text-xs text-gray-500">{t.category}</span>}
                        </div>
                        {t.bucket && (
                          <span className="text-xs text-gray-400 shrink-0 ml-2">
                            {BUCKET_LABELS[t.bucket] ?? t.bucket}
                          </span>
                        )}
                      </Link>
                    ))
                  : <p className="text-sm text-gray-500">Loading...</p>
              }
              <Link to="/tasks" className="text-xs text-blue-600 hover:text-blue-800 block mt-2">
                View all tasks &rarr;
              </Link>
            </div>
          )}
        </CollapsibleSection>

        {/* Pipeline with bar chart */}
        <CollapsibleSection title="Opportunity Pipeline" storageKey="pipeline">
          {pipeline?.stages ? (
            <div className="space-y-3">
              {pipeline.stages.length > 0 ? (
                <>
                  {pipeline.stages.map((s) => {
                    const pct = (s.total_amount / maxStageAmount) * 100;
                    const barColor = STAGE_COLORS[s.stage] ?? 'bg-gray-400';
                    return (
                      <div key={s.stage}>
                        <div className="flex items-center justify-between mb-1">
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
                        <div className="w-full bg-gray-100 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${barColor}`}
                            style={{ width: `${Math.max(pct, 2)}%` }}
                          />
                        </div>
                      </div>
                    );
                  })}
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
        </CollapsibleSection>
      </div>

      {/* Activity Feed */}
      <div className="mt-6">
        <CollapsibleSection title="Recent Activity" storageKey="activity">
          {activity && activity.length > 0 ? (
            <div className="space-y-3">
              {activity.map((v) => {
                const slug = ENTITY_SLUG[v.item_type];
                const colorClass = EVENT_COLORS[v.event] ?? 'bg-gray-100 text-gray-700';
                return (
                  <div key={v.id} className="flex items-start gap-3">
                    <span className={`text-xs font-medium px-2 py-0.5 rounded capitalize shrink-0 ${colorClass}`}>
                      {v.event}
                    </span>
                    <div className="text-sm text-gray-700 min-w-0">
                      <span className="text-gray-500">{v.item_type}</span>
                      {slug ? (
                        <Link to={`/${slug}/${v.item_id}`} className="ml-1 text-blue-600 hover:text-blue-800">
                          #{v.item_id}
                        </Link>
                      ) : (
                        <span className="ml-1">#{v.item_id}</span>
                      )}
                      {v.whodunnit && <span className="text-gray-400 ml-1">by user {v.whodunnit}</span>}
                    </div>
                    <span className="text-xs text-gray-400 shrink-0 ml-auto">
                      {formatRelativeTime(v.created_at)}
                    </span>
                  </div>
                );
              })}
            </div>
          ) : activity ? (
            <p className="text-sm text-gray-500">No recent activity.</p>
          ) : (
            <p className="text-sm text-gray-500">Loading...</p>
          )}
        </CollapsibleSection>
      </div>
    </div>
  );
}

function formatRelativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(iso).toLocaleDateString();
}
