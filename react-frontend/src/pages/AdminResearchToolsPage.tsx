import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useApi } from '../hooks/useApi';
import { useMutation } from '../hooks/useMutation';
import { useAuth } from '../auth/AuthContext';

interface ResearchTool {
  id: number;
  name: string;
  url_template: string;
  enabled: boolean;
}

interface FormState {
  name: string;
  url_template: string;
  enabled: boolean;
}

const emptyForm: FormState = { name: '', url_template: '', enabled: false };

export function AdminResearchToolsPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { data: tools, loading, error, refetch } = useApi<ResearchTool[]>('/admin/research_tools');
  const mutation = useMutation();
  const [editing, setEditing] = useState<number | 'new' | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [formError, setFormError] = useState('');

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

  function startNew() {
    setEditing('new');
    setForm(emptyForm);
    setFormError('');
  }

  function startEdit(tool: ResearchTool) {
    setEditing(tool.id);
    setForm({ name: tool.name, url_template: tool.url_template, enabled: tool.enabled });
    setFormError('');
  }

  function cancel() {
    setEditing(null);
    setForm(emptyForm);
    setFormError('');
  }

  async function save() {
    setFormError('');
    if (!form.name.trim() || !form.url_template.trim()) {
      setFormError('Name and URL template are required');
      return;
    }
    try {
      if (editing === 'new') {
        await mutation.post('/admin/research_tools', form);
      } else {
        await mutation.put(`/admin/research_tools/${editing}`, form);
      }
      setEditing(null);
      setForm(emptyForm);
      refetch();
    } catch {
      setFormError(mutation.error || 'Save failed');
    }
  }

  async function remove(id: number) {
    if (!confirm('Delete this research tool?')) return;
    try {
      await mutation.del(`/admin/research_tools/${id}`);
      refetch();
    } catch {
      // error shown via mutation.error
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Research Tools</h1>
        {editing === null && (
          <button
            onClick={startNew}
            className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
          >
            Add Research Tool
          </button>
        )}
      </div>

      <p className="text-sm text-gray-600">
        Research tools provide quick lookup links for entities. Use <code>{'{query}'}</code> in the URL template as a placeholder.
      </p>

      {mutation.error && editing === null && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
          {mutation.error}
        </div>
      )}

      {editing !== null && (
        <div className="bg-white border border-gray-200 rounded-lg p-6 space-y-4">
          <h2 className="text-lg font-medium text-gray-900">
            {editing === 'new' ? 'New Research Tool' : 'Edit Research Tool'}
          </h2>
          {formError && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
              {formError}
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g. Google Scholar"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">URL Template</label>
            <input
              type="text"
              value={form.url_template}
              onChange={(e) => setForm({ ...form, url_template: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g. https://scholar.google.com/scholar?q={query}"
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="enabled"
              checked={form.enabled}
              onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
              className="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
            />
            <label htmlFor="enabled" className="text-sm text-gray-700">Enabled</label>
          </div>
          <div className="flex gap-3">
            <button
              onClick={save}
              disabled={mutation.loading}
              className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {mutation.loading ? 'Saving...' : 'Save'}
            </button>
            <button
              onClick={cancel}
              className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-200"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">URL Template</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {(!tools || tools.length === 0) && (
              <tr>
                <td colSpan={4} className="px-6 py-8 text-center text-sm text-gray-500">
                  No research tools configured yet.
                </td>
              </tr>
            )}
            {tools?.map((tool) => (
              <tr key={tool.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{tool.name}</td>
                <td className="px-6 py-4 text-sm text-gray-500 font-mono truncate max-w-xs">{tool.url_template}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <span className={`inline-flex px-2 py-0.5 text-xs font-medium rounded-full ${tool.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                    {tool.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-3">
                  <button
                    onClick={() => startEdit(tool)}
                    className="text-blue-600 hover:text-blue-900"
                    disabled={editing !== null}
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => remove(tool.id)}
                    className="text-red-600 hover:text-red-900"
                    disabled={editing !== null}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
