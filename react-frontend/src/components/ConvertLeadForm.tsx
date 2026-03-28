import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Modal } from './Modal';
import { api } from '../api/client';
import type { Lead, Account, PaginatedResult } from '../types/entities';
import { useApi } from '../hooks/useApi';

interface ConvertLeadFormProps {
  lead: Lead;
  open: boolean;
  onClose: () => void;
}

interface ConvertResponse {
  account: Account;
  contact: { id: number };
  opportunity: { id: number };
}

export function ConvertLeadForm({ lead, open, onClose }: ConvertLeadFormProps) {
  const navigate = useNavigate();
  const { data: accountsData } = useApi<PaginatedResult<Account>>('/accounts?per_page=100&sort=name&order=asc');

  const [accountMode, setAccountMode] = useState<'existing' | 'new'>('new');
  const [accountId, setAccountId] = useState<string>('');
  const [accountName, setAccountName] = useState(lead.company ?? '');
  const [oppName, setOppName] = useState('');
  const [oppStage, setOppStage] = useState('prospecting');
  const [oppAmount, setOppAmount] = useState('');
  const [oppProbability, setOppProbability] = useState('');
  const [oppClosesOn, setOppClosesOn] = useState('');
  const [access, setAccess] = useState('Public');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const accounts = accountsData?.data ?? [];

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const body: Record<string, unknown> = {
      account: accountMode === 'existing'
        ? { id: Number(accountId) }
        : { name: accountName },
      opportunity: {
        name: oppName,
        stage: oppStage,
        ...(oppAmount ? { amount: Number(oppAmount) } : {}),
        ...(oppProbability ? { probability: Number(oppProbability) } : {}),
        ...(oppClosesOn ? { closes_on: oppClosesOn } : {}),
      },
      access,
    };

    try {
      const resp = await api.post<ConvertResponse>(`/leads/${lead.id}/convert`, body);
      onClose();
      navigate(`/contacts/${resp.contact.id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Conversion failed';
      try {
        const parsed = JSON.parse(msg);
        setError(parsed.error || msg);
      } catch {
        setError(msg);
      }
    } finally {
      setLoading(false);
    }
  }

  const stages = [
    { value: 'prospecting', label: 'Prospecting' },
    { value: 'analysis', label: 'Analysis' },
    { value: 'presentation', label: 'Presentation' },
    { value: 'proposal', label: 'Proposal' },
    { value: 'negotiation', label: 'Negotiation' },
    { value: 'final_review', label: 'Final Review' },
    { value: 'won', label: 'Won' },
    { value: 'lost', label: 'Lost' },
  ];

  return (
    <Modal open={open} onClose={onClose} title="Convert Lead">
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded text-sm">
            {error}
          </div>
        )}

        {/* Account Section */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-2">Account</h3>
          <div className="flex gap-4 mb-3">
            <label className="flex items-center gap-1.5 text-sm">
              <input
                type="radio"
                checked={accountMode === 'new'}
                onChange={() => setAccountMode('new')}
              />
              Create new
            </label>
            <label className="flex items-center gap-1.5 text-sm">
              <input
                type="radio"
                checked={accountMode === 'existing'}
                onChange={() => setAccountMode('existing')}
              />
              Select existing
            </label>
          </div>
          {accountMode === 'new' ? (
            <input
              type="text"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              placeholder="Account name"
              required
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          ) : (
            <select
              value={accountId}
              onChange={(e) => setAccountId(e.target.value)}
              required
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Select an account...</option>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* Opportunity Section */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-2">Opportunity</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={oppName}
              onChange={(e) => setOppName(e.target.value)}
              placeholder="Opportunity name *"
              required
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div className="grid grid-cols-2 gap-3">
              <select
                value={oppStage}
                onChange={(e) => setOppStage(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {stages.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
              <input
                type="number"
                value={oppAmount}
                onChange={(e) => setOppAmount(e.target.value)}
                placeholder="Amount"
                className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <input
                type="number"
                value={oppProbability}
                onChange={(e) => setOppProbability(e.target.value)}
                placeholder="Probability (%)"
                min="0"
                max="100"
                className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <input
                type="date"
                value={oppClosesOn}
                onChange={(e) => setOppClosesOn(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
        </div>

        {/* Access */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-2">Access</h3>
          <select
            value={access}
            onChange={(e) => setAccess(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="Public">Public</option>
            <option value="Private">Private</option>
            <option value="Shared">Shared</option>
          </select>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
          >
            {loading ? 'Converting...' : 'Convert'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
