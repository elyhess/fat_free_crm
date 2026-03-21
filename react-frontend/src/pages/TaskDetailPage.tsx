import { EntityDetailPage } from './EntityDetailPage';
import { taskFields } from '../config/entityFields';
import type { Task } from '../types/entities';
import type { FieldDef } from '../components/EntityForm';

const detailFields = [
  { key: 'name', label: 'Name' },
  { key: 'priority', label: 'Priority' },
  { key: 'category', label: 'Category' },
  { key: 'bucket', label: 'Due' },
  { key: 'due_at', label: 'Due Date', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'completed_at', label: 'Completed', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'background_info', label: 'Notes' },
  { key: 'created_at', label: 'Created', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
  { key: 'updated_at', label: 'Updated', render: (v: unknown) => v ? new Date(v as string).toLocaleDateString() : '' },
];

export function TaskDetailPage() {
  return (
    <EntityDetailPage<Task>
      entityName="Task"
      entitySlug="tasks"
      endpoint="/tasks"
      fields={detailFields}
      formFields={taskFields as FieldDef[]}
      getTitle={(t) => t.name}
    />
  );
}
