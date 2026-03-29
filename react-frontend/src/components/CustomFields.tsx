import { useApi } from '../hooks/useApi';

interface FieldDef {
  id: number;
  name: string;
  label: string;
  as: string;
  hint?: string;
  required: boolean;
  collection?: string;
}

interface FieldGroup {
  id: number;
  label: string;
  fields: FieldDef[];
}

interface FieldGroupsResponse {
  entity_type: string;
  field_groups: FieldGroup[];
}

interface CustomFieldsDisplayProps {
  entityType: string;
  entitySlug: string;
  entityId: number;
}

export function CustomFieldsDisplay({ entityType, entitySlug, entityId }: CustomFieldsDisplayProps) {
  const { data: fieldDefs } = useApi<FieldGroupsResponse>(`/field_groups?entity=${entityType}`);
  const { data: values } = useApi<Record<string, unknown>>(`/${entitySlug}/${entityId}/custom_fields`);

  if (!fieldDefs || !values) return null;

  const customGroups = fieldDefs.field_groups.filter(g =>
    g.fields.some(f => f.name.startsWith('cf_'))
  );

  if (customGroups.length === 0) return null;

  const hasAnyValue = customGroups.some(g =>
    g.fields.some(f => f.name.startsWith('cf_') && values[f.name] != null && values[f.name] !== '')
  );

  if (!hasAnyValue && Object.keys(values).length === 0) return null;

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <h2 className="text-lg font-medium text-gray-900 mb-4">Custom Fields</h2>
      {customGroups.map(group => (
        <div key={group.id} className="mb-4 last:mb-0">
          {customGroups.length > 1 && (
            <h3 className="text-sm font-medium text-gray-700 mb-2">{group.label}</h3>
          )}
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3">
            {group.fields
              .filter(f => f.name.startsWith('cf_'))
              .map(field => {
                const val = values[field.name];
                const display = formatFieldValue(val, field.as);
                return (
                  <div key={field.name}>
                    <dt className="text-sm font-medium text-gray-500">{field.label}</dt>
                    <dd className="text-sm text-gray-900">{display || <span className="text-gray-400">—</span>}</dd>
                  </div>
                );
              })}
          </dl>
        </div>
      ))}
    </div>
  );
}

function formatFieldValue(val: unknown, fieldType: string): string {
  if (val == null || val === '') return '';
  switch (fieldType) {
    case 'boolean':
      return val ? 'Yes' : 'No';
    case 'date':
      return new Date(val as string).toLocaleDateString();
    case 'datetime':
      return new Date(val as string).toLocaleString();
    case 'decimal':
    case 'float':
      return Number(val).toLocaleString();
    case 'check_boxes':
      if (Array.isArray(val)) return val.filter(Boolean).join(', ');
      return String(val);
    default:
      return String(val);
  }
}

// --- Form component for editing custom fields ---

interface CustomFieldsFormProps {
  entityType: string;
  values: Record<string, unknown>;
  onChange: (values: Record<string, unknown>) => void;
}

export function CustomFieldsForm({ entityType, values, onChange }: CustomFieldsFormProps) {
  const { data: fieldDefs } = useApi<FieldGroupsResponse>(`/field_groups?entity=${entityType}`);

  if (!fieldDefs) return null;

  const customGroups = fieldDefs.field_groups.filter(g =>
    g.fields.some(f => f.name.startsWith('cf_'))
  );

  if (customGroups.length === 0) return null;

  function setValue(name: string, val: unknown) {
    onChange({ ...values, [name]: val });
  }

  const inputClass = 'w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500';
  const labelClass = 'block text-sm font-medium text-gray-700 mb-1';

  return (
    <div className="border-t border-gray-200 pt-4 mt-4">
      <h3 className="text-sm font-medium text-gray-900 mb-3">Custom Fields</h3>
      {customGroups.map(group => (
        <div key={group.id} className="mb-4">
          {customGroups.length > 1 && (
            <h4 className="text-xs font-medium text-gray-600 mb-2">{group.label}</h4>
          )}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {group.fields
              .filter(f => f.name.startsWith('cf_'))
              .map(field => (
                <div key={field.name}>
                  <label className={labelClass}>
                    {field.label}
                    {field.required && <span className="text-red-500 ml-1">*</span>}
                  </label>
                  {renderFieldInput(field, values[field.name], (v) => setValue(field.name, v), inputClass)}
                  {field.hint && <p className="text-xs text-gray-500 mt-1">{field.hint}</p>}
                </div>
              ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function renderFieldInput(
  field: FieldDef,
  value: unknown,
  onChange: (val: unknown) => void,
  inputClass: string,
) {
  const strVal = value != null ? String(value) : '';

  switch (field.as) {
    case 'text':
      return (
        <textarea
          value={strVal}
          onChange={e => onChange(e.target.value)}
          rows={3}
          className={inputClass}
          required={field.required}
        />
      );

    case 'boolean':
      return (
        <input
          type="checkbox"
          checked={!!value}
          onChange={e => onChange(e.target.checked)}
          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        />
      );

    case 'select':
    case 'radio_buttons': {
      const options = field.collection ? field.collection.split('|').map(s => s.trim()) : [];
      return (
        <select
          value={strVal}
          onChange={e => onChange(e.target.value)}
          className={inputClass}
          required={field.required}
        >
          <option value="">— Select —</option>
          {options.map(opt => (
            <option key={opt} value={opt}>{opt}</option>
          ))}
        </select>
      );
    }

    case 'check_boxes': {
      const options = field.collection ? field.collection.split('|').map(s => s.trim()) : [];
      const selected = Array.isArray(value) ? value : [];
      return (
        <div className="space-y-1">
          {options.map(opt => (
            <label key={opt} className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={selected.includes(opt)}
                onChange={e => {
                  const next = e.target.checked
                    ? [...selected, opt]
                    : selected.filter((v: string) => v !== opt);
                  onChange(next);
                }}
                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              {opt}
            </label>
          ))}
        </div>
      );
    }

    case 'date':
      return (
        <input
          type="date"
          value={strVal}
          onChange={e => onChange(e.target.value)}
          className={inputClass}
          required={field.required}
        />
      );

    case 'datetime':
      return (
        <input
          type="datetime-local"
          value={strVal}
          onChange={e => onChange(e.target.value)}
          className={inputClass}
          required={field.required}
        />
      );

    case 'integer':
      return (
        <input
          type="number"
          value={strVal}
          onChange={e => onChange(e.target.value ? parseInt(e.target.value, 10) : '')}
          className={inputClass}
          required={field.required}
          step="1"
        />
      );

    case 'decimal':
    case 'float':
      return (
        <input
          type="number"
          value={strVal}
          onChange={e => onChange(e.target.value ? parseFloat(e.target.value) : '')}
          className={inputClass}
          required={field.required}
          step="any"
        />
      );

    default: // string, email, url, tel
      return (
        <input
          type={field.as === 'email' ? 'email' : field.as === 'url' ? 'url' : field.as === 'tel' ? 'tel' : 'text'}
          value={strVal}
          onChange={e => onChange(e.target.value)}
          className={inputClass}
          required={field.required}
          placeholder={field.hint || ''}
        />
      );
  }
}

// Re-export types for use in other components
export type { FieldDef as CustomFieldDef, FieldGroup as CustomFieldGroup, FieldGroupsResponse };
