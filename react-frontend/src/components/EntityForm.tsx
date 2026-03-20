import { useState } from 'react';

export interface FieldDef {
  key: string;
  label: string;
  type?: 'text' | 'email' | 'tel' | 'number' | 'date' | 'select' | 'textarea';
  required?: boolean;
  options?: { value: string; label: string }[];
  placeholder?: string;
}

interface EntityFormProps {
  fields: FieldDef[];
  initialValues?: Record<string, unknown>;
  onSubmit: (values: Record<string, unknown>) => void;
  onCancel: () => void;
  loading?: boolean;
  error?: string | null;
  submitLabel?: string;
}

export function EntityForm({
  fields,
  initialValues = {},
  onSubmit,
  onCancel,
  loading = false,
  error,
  submitLabel = 'Save',
}: EntityFormProps) {
  const [values, setValues] = useState<Record<string, unknown>>(() => {
    const v: Record<string, unknown> = {};
    for (const f of fields) {
      v[f.key] = initialValues[f.key] ?? '';
    }
    return v;
  });

  function handleChange(key: string, value: unknown) {
    setValues((prev) => ({ ...prev, [key]: value }));
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    // Strip empty strings for optional fields, convert numbers
    const cleaned: Record<string, unknown> = {};
    for (const f of fields) {
      const val = values[f.key];
      if (f.type === 'number' && val !== '' && val != null) {
        cleaned[f.key] = Number(val);
      } else if (val !== '' && val != null) {
        cleaned[f.key] = val;
      }
    }
    onSubmit(cleaned);
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded text-sm">
          {error}
        </div>
      )}

      {fields.map((f) => (
        <div key={f.key}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {f.label}
            {f.required && <span className="text-red-500 ml-0.5">*</span>}
          </label>

          {f.type === 'select' ? (
            <select
              value={String(values[f.key] ?? '')}
              onChange={(e) => handleChange(f.key, e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              required={f.required}
            >
              <option value="">-- Select --</option>
              {f.options?.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          ) : f.type === 'textarea' ? (
            <textarea
              value={String(values[f.key] ?? '')}
              onChange={(e) => handleChange(f.key, e.target.value)}
              placeholder={f.placeholder}
              rows={3}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              required={f.required}
            />
          ) : (
            <input
              type={f.type ?? 'text'}
              value={String(values[f.key] ?? '')}
              onChange={(e) => handleChange(f.key, e.target.value)}
              placeholder={f.placeholder}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              required={f.required}
            />
          )}
        </div>
      ))}

      <div className="flex justify-end gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-sm border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          disabled={loading}
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          disabled={loading}
        >
          {loading ? 'Saving...' : submitLabel}
        </button>
      </div>
    </form>
  );
}
