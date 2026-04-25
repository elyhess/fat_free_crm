import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';
import { useAuth } from '../auth/AuthContext';
import { Modal } from '../components/Modal';
import type { CustomFieldGroup, FieldGroupsResponse } from '../components/CustomFields';

const ENTITY_TYPES = ['Account', 'Contact', 'Lead', 'Opportunity', 'Campaign', 'Task'];

const FIELD_TYPES = [
  { value: 'string', label: 'Text (single line)' },
  { value: 'text', label: 'Text (multi-line)' },
  { value: 'email', label: 'Email' },
  { value: 'url', label: 'URL' },
  { value: 'tel', label: 'Phone' },
  { value: 'select', label: 'Dropdown' },
  { value: 'radio_buttons', label: 'Radio Buttons' },
  { value: 'check_boxes', label: 'Checkboxes' },
  { value: 'boolean', label: 'Yes/No' },
  { value: 'date', label: 'Date' },
  { value: 'datetime', label: 'Date & Time' },
  { value: 'integer', label: 'Integer' },
  { value: 'decimal', label: 'Decimal' },
  { value: 'float', label: 'Float' },
];

interface FieldDef {
  id: number;
  name: string;
  label: string;
  as: string;
  hint?: string;
  placeholder?: string;
  required: boolean;
  disabled?: boolean;
  collection?: string;
  position?: number;
}

export function AdminFieldsPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [selectedEntity, setSelectedEntity] = useState('Account');
  const { data, loading, refetch } = useApi<FieldGroupsResponse>(`/field_groups?entity=${selectedEntity}`);
  const mutation = useMutation();

  const [showCreate, setShowCreate] = useState(false);
  const [editField, setEditField] = useState<FieldDef | null>(null);
  const [createGroupId, setCreateGroupId] = useState<number | null>(null);

  if (!user?.admin) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Admin access required.
        <button onClick={() => navigate('/')} className="ml-2 text-blue-600 hover:text-blue-800 underline">Go home</button>
      </div>
    );
  }

  const fieldGroups = data?.field_groups ?? [];

  async function handleCreateField(values: Record<string, unknown>) {
    try {
      await mutation.post('/admin/fields', values);
      setShowCreate(false);
      setCreateGroupId(null);
      refetch();
    } catch { /* */ }
  }

  async function handleUpdateField(id: number, values: Record<string, unknown>) {
    try {
      await mutation.put(`/admin/fields/${id}`, values);
      setEditField(null);
      refetch();
    } catch { /* */ }
  }

  async function handleDeleteField(id: number) {
    if (!confirm('Delete this field? Data in this column will be lost.')) return;
    try {
      await mutation.del(`/admin/fields/${id}`);
      refetch();
    } catch { /* */ }
  }

  const inputClass = 'w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500';
  const labelClass = 'block text-sm font-medium text-gray-700 mb-1';

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Custom Fields</h1>
        <button onClick={() => navigate('/')} className="text-sm text-blue-600 hover:text-blue-800">
          &larr; Back to Dashboard
        </button>
      </div>

      {/* Entity type tabs */}
      <div className="flex gap-1 mb-6 border-b border-gray-200">
        {ENTITY_TYPES.map(type => (
          <button
            key={type}
            onClick={() => setSelectedEntity(type)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              selectedEntity === type
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {type}s
          </button>
        ))}
      </div>

      {loading && <div className="text-gray-500">Loading...</div>}

      {/* Field groups */}
      {fieldGroups.map((group: CustomFieldGroup) => (
        <div key={group.id} className="bg-white shadow rounded-lg p-6 mb-4">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-medium text-gray-900">{group.label}</h2>
            <button
              onClick={() => { setCreateGroupId(group.id); mutation.reset(); setShowCreate(true); }}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Add Field
            </button>
          </div>

          {group.fields.filter(f => f.name.startsWith('cf_')).length === 0 ? (
            <p className="text-sm text-gray-500">No custom fields. Click "Add Field" to create one.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="pb-2 font-medium">Label</th>
                  <th className="pb-2 font-medium">Column</th>
                  <th className="pb-2 font-medium">Type</th>
                  <th className="pb-2 font-medium">Required</th>
                  <th className="pb-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {group.fields
                  .filter(f => f.name.startsWith('cf_'))
                  .map((field: FieldDef) => (
                    <tr key={field.id} className="border-b border-gray-100 last:border-0">
                      <td className="py-2">{field.label}</td>
                      <td className="py-2 text-gray-500 font-mono text-xs">{field.name}</td>
                      <td className="py-2">{FIELD_TYPES.find(t => t.value === field.as)?.label ?? field.as}</td>
                      <td className="py-2">{field.required ? 'Yes' : 'No'}</td>
                      <td className="py-2 text-right">
                        <button
                          onClick={() => { mutation.reset(); setEditField(field); }}
                          className="text-blue-600 hover:text-blue-800 text-xs mr-3"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDeleteField(field.id)}
                          className="text-red-600 hover:text-red-800 text-xs"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          )}
        </div>
      ))}

      {/* Create Field Modal */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Add Custom Field">
        <FieldForm
          onSubmit={(values) => handleCreateField({ ...values, field_group_id: createGroupId })}
          onCancel={() => setShowCreate(false)}
          loading={mutation.loading}
          error={mutation.error}
          inputClass={inputClass}
          labelClass={labelClass}
        />
      </Modal>

      {/* Edit Field Modal */}
      <Modal open={!!editField} onClose={() => setEditField(null)} title="Edit Custom Field">
        {editField && (
          <FieldForm
            initialValues={editField}
            onSubmit={(values) => handleUpdateField(editField.id, values)}
            onCancel={() => setEditField(null)}
            loading={mutation.loading}
            error={mutation.error}
            inputClass={inputClass}
            labelClass={labelClass}
            isEdit
          />
        )}
      </Modal>
    </div>
  );
}

interface FieldFormProps {
  initialValues?: Partial<FieldDef>;
  onSubmit: (values: Record<string, unknown>) => void;
  onCancel: () => void;
  loading: boolean;
  error: string | null;
  inputClass: string;
  labelClass: string;
  isEdit?: boolean;
}

function FieldForm({ initialValues, onSubmit, onCancel, loading, error, inputClass, labelClass, isEdit }: FieldFormProps) {
  const [label, setLabel] = useState(initialValues?.label ?? '');
  const [as, setAs] = useState(initialValues?.as ?? 'string');
  const [hint, setHint] = useState(initialValues?.hint ?? '');
  const [placeholder, setPlaceholder] = useState(initialValues?.placeholder ?? '');
  const [required, setRequired] = useState(initialValues?.required ?? false);
  const [collection, setCollection] = useState(initialValues?.collection ?? '');

  const needsCollection = ['select', 'radio_buttons', 'check_boxes'].includes(as);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const values: Record<string, unknown> = { label, hint, placeholder, required };
    if (!isEdit) {
      values.as = as;
    }
    if (needsCollection) {
      values.collection = collection;
    }
    onSubmit(values);
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded text-sm">{error}</div>
      )}
      <div>
        <label className={labelClass}>Label</label>
        <input type="text" value={label} onChange={e => setLabel(e.target.value)} className={inputClass} required />
      </div>
      {!isEdit && (
        <div>
          <label className={labelClass}>Field Type</label>
          <select value={as} onChange={e => setAs(e.target.value)} className={inputClass}>
            {FIELD_TYPES.map(t => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
        </div>
      )}
      {needsCollection && (
        <div>
          <label className={labelClass}>Options (pipe-separated)</label>
          <input
            type="text"
            value={collection}
            onChange={e => setCollection(e.target.value)}
            className={inputClass}
            placeholder="Option 1|Option 2|Option 3"
          />
        </div>
      )}
      <div>
        <label className={labelClass}>Hint</label>
        <input type="text" value={hint} onChange={e => setHint(e.target.value)} className={inputClass} />
      </div>
      <div>
        <label className={labelClass}>Placeholder</label>
        <input type="text" value={placeholder} onChange={e => setPlaceholder(e.target.value)} className={inputClass} />
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-700">
        <input
          type="checkbox"
          checked={required}
          onChange={e => setRequired(e.target.checked)}
          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        />
        Required
      </label>
      <div className="flex justify-end gap-2 pt-2">
        <button type="button" onClick={onCancel} className="px-4 py-2 text-sm text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">
          Cancel
        </button>
        <button type="submit" disabled={loading} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
          {loading ? 'Saving...' : isEdit ? 'Update' : 'Create'}
        </button>
      </div>
    </form>
  );
}
