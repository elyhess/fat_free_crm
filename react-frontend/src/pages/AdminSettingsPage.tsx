import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';
import { useAuth } from '../auth/AuthContext';

type Settings = Record<string, unknown>;

export function AdminSettingsPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { data: settings, loading, error, refetch } = useApi<Settings>('/admin/settings');
  const mutation = useMutation();
  const [form, setForm] = useState<Settings>({});
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (settings) {
      setForm(settings);
    }
  }, [settings]);

  if (!user?.admin) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Admin access required.
        <button onClick={() => navigate('/')} className="ml-2 text-blue-600 hover:text-blue-800 underline">Go home</button>
      </div>
    );
  }

  if (loading) return <div className="text-gray-500">Loading...</div>;
  if (error) return <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">{error}</div>;

  function setField(key: string, value: unknown) {
    setForm(prev => ({ ...prev, [key]: value }));
  }

  function setNestedField(parent: string, key: string, value: unknown) {
    setForm(prev => ({
      ...prev,
      [parent]: { ...(prev[parent] as Record<string, unknown> || {}), [key]: value },
    }));
  }

  function setArrayField(key: string, value: string) {
    // Split textarea by newlines into array
    setField(key, value.split('\n').filter(line => line.trim() !== ''));
  }

  function getArrayValue(key: string): string {
    const val = form[key];
    if (Array.isArray(val)) return val.join('\n');
    return '';
  }

  function getNestedValue(parent: string, key: string): string {
    const p = form[parent];
    if (p && typeof p === 'object' && !Array.isArray(p)) {
      return String((p as Record<string, unknown>)[key] ?? '');
    }
    return '';
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaved(false);
    try {
      await mutation.put('/admin/settings', form);
      setSaved(true);
      refetch();
    } catch { /* error in mutation */ }
  }

  const inputClass = 'w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500';
  const labelClass = 'block text-sm font-medium text-gray-700 mb-1';
  const checkboxLabelClass = 'flex items-center gap-2 text-sm text-gray-700';

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Application Settings</h1>
        <button onClick={() => navigate('/')} className="text-sm text-blue-600 hover:text-blue-800">
          &larr; Back to Dashboard
        </button>
      </div>

      {saved && (
        <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-2 rounded mb-4 text-sm">
          Settings updated successfully.
        </div>
      )}
      {mutation.error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-2 rounded mb-4 text-sm">
          {mutation.error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* General */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">General</h2>
          <div className="space-y-4">
            <div>
              <label className={labelClass}>Host</label>
              <input type="text" value={String(form.host ?? '')} onChange={e => setField('host', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Base URL</label>
              <input type="text" value={String(form.base_url ?? '')} onChange={e => setField('base_url', e.target.value)} className={inputClass} placeholder="Leave blank if deployed at root" />
            </div>
            <div>
              <label className={labelClass}>Locale</label>
              <input type="text" value={String(form.locale ?? '')} onChange={e => setField('locale', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Default Access</label>
              <select value={String(form.default_access ?? 'Public')} onChange={e => setField('default_access', e.target.value)} className={inputClass}>
                <option value="Public">Public</option>
                <option value="Private">Private</option>
                <option value="Shared">Shared</option>
              </select>
            </div>
            <div>
              <label className={labelClass}>User Signup</label>
              <select value={String(form.user_signup ?? ':not_allowed')} onChange={e => setField('user_signup', e.target.value)} className={inputClass}>
                <option value=":allowed">Allowed</option>
                <option value=":not_allowed">Not Allowed</option>
                <option value=":needs_approval">Needs Approval</option>
              </select>
            </div>
          </div>
        </div>

        {/* Toggles */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Options</h2>
          <div className="space-y-3">
            {[
              { key: 'per_user_locale', label: 'Allow per-user locale selection' },
              { key: 'compound_address', label: 'Use compound address fields' },
              { key: 'task_calendar_with_time', label: 'Show time in task calendar' },
              { key: 'require_first_names', label: 'Require first names on leads/contacts' },
              { key: 'require_last_names', label: 'Require last names on leads/contacts' },
              { key: 'require_unique_account_names', label: 'Require unique account names' },
              { key: 'comments_visible_on_dashboard', label: 'Show comments on dashboard' },
              { key: 'enforce_international_phone_format', label: 'Enforce international phone format' },
            ].map(({ key, label }) => (
              <label key={key} className={checkboxLabelClass}>
                <input
                  type="checkbox"
                  checked={!!form[key]}
                  onChange={e => setField(key, e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                {label}
              </label>
            ))}
          </div>
        </div>

        {/* Dropdown Options */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Dropdown Options</h2>
          <p className="text-xs text-gray-500 mb-4">One value per line. Values starting with ":" are treated as translatable symbols.</p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {[
              { key: 'account_category', label: 'Account Categories' },
              { key: 'campaign_status', label: 'Campaign Statuses' },
              { key: 'lead_status', label: 'Lead Statuses' },
              { key: 'lead_source', label: 'Lead Sources' },
              { key: 'opportunity_stage', label: 'Opportunity Stages' },
              { key: 'task_category', label: 'Task Categories' },
            ].map(({ key, label }) => (
              <div key={key}>
                <label className={labelClass}>{label}</label>
                <textarea
                  value={getArrayValue(key)}
                  onChange={e => setArrayField(key, e.target.value)}
                  rows={5}
                  className={inputClass}
                />
              </div>
            ))}
          </div>
        </div>

        {/* SMTP */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">SMTP (Outgoing Email)</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Server Address</label>
              <input type="text" value={getNestedValue('smtp', 'address')} onChange={e => setNestedField('smtp', 'address', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Port</label>
              <input type="text" value={getNestedValue('smtp', 'port')} onChange={e => setNestedField('smtp', 'port', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>From Address</label>
              <input type="text" value={getNestedValue('smtp', 'from')} onChange={e => setNestedField('smtp', 'from', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Authentication</label>
              <input type="text" value={getNestedValue('smtp', 'authentication')} onChange={e => setNestedField('smtp', 'authentication', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Username</label>
              <input type="text" value={getNestedValue('smtp', 'user_name')} onChange={e => setNestedField('smtp', 'user_name', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Password</label>
              <input type="password" value={getNestedValue('smtp', 'password')} onChange={e => setNestedField('smtp', 'password', e.target.value)} className={inputClass} />
            </div>
          </div>
        </div>

        {/* Email Dropbox */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Email Dropbox (IMAP)</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>Server</label>
              <input type="text" value={getNestedValue('email_dropbox', 'server')} onChange={e => setNestedField('email_dropbox', 'server', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Port</label>
              <input type="text" value={getNestedValue('email_dropbox', 'port')} onChange={e => setNestedField('email_dropbox', 'port', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Address</label>
              <input type="text" value={getNestedValue('email_dropbox', 'address')} onChange={e => setNestedField('email_dropbox', 'address', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>User</label>
              <input type="text" value={getNestedValue('email_dropbox', 'user')} onChange={e => setNestedField('email_dropbox', 'user', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Password</label>
              <input type="password" value={getNestedValue('email_dropbox', 'password')} onChange={e => setNestedField('email_dropbox', 'password', e.target.value)} className={inputClass} />
            </div>
            <div>
              <label className={labelClass}>Scan Folder</label>
              <input type="text" value={getNestedValue('email_dropbox', 'scan_folder')} onChange={e => setNestedField('email_dropbox', 'scan_folder', e.target.value)} className={inputClass} />
            </div>
            <label className={checkboxLabelClass + ' col-span-2'}>
              <input
                type="checkbox"
                checked={!!getNestedValue('email_dropbox', 'ssl')}
                onChange={e => setNestedField('email_dropbox', 'ssl', e.target.checked)}
                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              Use SSL
            </label>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={mutation.loading}
            className="px-6 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {mutation.loading ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </form>
    </div>
  );
}
