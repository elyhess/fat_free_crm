import { EntityList } from '../components/EntityList';
import type { Task } from '../types/entities';

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'bucket', label: 'Due' },
  { key: 'priority', label: 'Priority' },
  { key: 'category', label: 'Category' },
  {
    key: 'completed_at',
    label: 'Status',
    render: (t: Task) =>
      t.completed_at ? (
        <span className="text-green-600 font-medium">Completed</span>
      ) : (
        <span className="text-yellow-600 font-medium">Pending</span>
      ),
  },
  {
    key: 'due_at',
    label: 'Due At',
    render: (t: Task) =>
      t.due_at ? new Date(t.due_at).toLocaleString() : '',
  },
];

export function TasksPage() {
  return (
    <EntityList<Task>
      title="Tasks"
      endpoint="/tasks"
      columns={columns}
      getRowKey={(t) => t.id}
    />
  );
}
