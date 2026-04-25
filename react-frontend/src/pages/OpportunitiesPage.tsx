import { useState } from 'react';
import { EntityList, type FilterDef, type Column } from '../components/EntityList';
import { PipelineBoard } from '../components/PipelineBoard';
import { useApi } from '../hooks/useApi';
import { opportunityFields } from '../config/entityFields';
import type { Opportunity, PaginatedResult } from '../types/entities';

const STAGE_OPTIONS = [
  { value: 'prospecting', label: 'Prospecting' },
  { value: 'analysis', label: 'Analysis' },
  { value: 'presentation', label: 'Presentation' },
  { value: 'proposal', label: 'Proposal' },
  { value: 'negotiation', label: 'Negotiation' },
  { value: 'final_review', label: 'Final Review' },
  { value: 'won', label: 'Won' },
  { value: 'lost', label: 'Lost' },
];

const filterDefs: FilterDef[] = [
  { key: 'name', label: 'Name', type: 'text' },
  { key: 'stage', label: 'Stage', type: 'select', operator: 'eq', options: STAGE_OPTIONS },
];

function formatCurrency(val?: number): string {
  if (val == null) return '';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
}

const columns: Column<Opportunity>[] = [
  { key: 'name', label: 'Name' },
  {
    key: 'stage',
    label: 'Stage',
    inlineEdit: { type: 'select', options: STAGE_OPTIONS },
  },
  {
    key: 'amount',
    label: 'Amount',
    render: (o: Opportunity) => formatCurrency(o.amount),
  },
  {
    key: 'probability',
    label: 'Probability',
    render: (o: Opportunity) => (o.probability != null ? `${o.probability}%` : ''),
  },
  {
    key: 'closes_on',
    label: 'Closes On',
    render: (o: Opportunity) =>
      o.closes_on ? new Date(o.closes_on).toLocaleDateString() : '',
  },
];

export function OpportunitiesPage() {
  const [view, setView] = useState<'table' | 'board'>('table');
  const { data: boardData, refetch: refetchBoard } = useApi<PaginatedResult<Opportunity>>(
    '/opportunities?per_page=200&sort=amount&order=desc'
  );

  return (
    <div>
      {/* View Toggle */}
      <div className="flex items-center justify-end mb-4 gap-1">
        <button
          onClick={() => setView('table')}
          className={`px-3 py-1.5 text-sm rounded-l-md border ${
            view === 'table'
              ? 'bg-blue-600 text-white border-blue-600'
              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
          }`}
        >
          Table
        </button>
        <button
          onClick={() => setView('board')}
          className={`px-3 py-1.5 text-sm rounded-r-md border ${
            view === 'board'
              ? 'bg-blue-600 text-white border-blue-600'
              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
          }`}
        >
          Board
        </button>
      </div>

      {view === 'table' ? (
        <EntityList<Opportunity>
          title="Opportunities"
          endpoint="/opportunities"
          columns={columns}
          getRowKey={(o) => o.id}
          formFields={opportunityFields}
          detailPath={(o) => `/opportunities/${o.id}`}
          filterFields={filterDefs}
        />
      ) : (
        <div>
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-2xl font-semibold text-gray-900">Opportunities</h1>
          </div>
          {boardData ? (
            <PipelineBoard opportunities={boardData.data} onStageChange={refetchBoard} />
          ) : (
            <div className="text-gray-500">Loading...</div>
          )}
        </div>
      )}
    </div>
  );
}
