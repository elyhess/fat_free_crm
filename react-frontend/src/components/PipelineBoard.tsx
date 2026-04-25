import { useState, useRef } from 'react';
import { Link } from 'react-router-dom';
import { useMutation } from '../hooks/useMutation';
import type { Opportunity } from '../types/entities';

const STAGES = [
  { value: 'prospecting', label: 'Prospecting', color: 'bg-gray-100 border-gray-300' },
  { value: 'analysis', label: 'Analysis', color: 'bg-blue-50 border-blue-300' },
  { value: 'presentation', label: 'Presentation', color: 'bg-indigo-50 border-indigo-300' },
  { value: 'proposal', label: 'Proposal', color: 'bg-purple-50 border-purple-300' },
  { value: 'negotiation', label: 'Negotiation', color: 'bg-yellow-50 border-yellow-300' },
  { value: 'final_review', label: 'Final Review', color: 'bg-orange-50 border-orange-300' },
  { value: 'won', label: 'Won', color: 'bg-green-50 border-green-300' },
  { value: 'lost', label: 'Lost', color: 'bg-red-50 border-red-300' },
];

function formatCurrency(val?: number): string {
  if (val == null) return '';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);
}

interface PipelineBoardProps {
  opportunities: Opportunity[];
  onStageChange: () => void;
}

export function PipelineBoard({ opportunities, onStageChange }: PipelineBoardProps) {
  const mutation = useMutation();
  const [dragItem, setDragItem] = useState<number | null>(null);
  const [dragOverStage, setDragOverStage] = useState<string | null>(null);
  const dragCounter = useRef<Record<string, number>>({});

  const grouped: Record<string, Opportunity[]> = {};
  for (const stage of STAGES) {
    grouped[stage.value] = [];
  }
  for (const opp of opportunities) {
    const stage = opp.stage || 'prospecting';
    if (grouped[stage]) {
      grouped[stage].push(opp);
    }
  }

  async function handleDrop(newStage: string) {
    if (dragItem == null) return;
    const opp = opportunities.find((o) => o.id === dragItem);
    if (!opp || opp.stage === newStage) {
      setDragItem(null);
      setDragOverStage(null);
      return;
    }
    try {
      await mutation.put(`/opportunities/${opp.id}`, { stage: newStage });
      onStageChange();
    } catch { /* error in mutation */ }
    setDragItem(null);
    setDragOverStage(null);
  }

  return (
    <div className="flex gap-3 overflow-x-auto pb-4" style={{ minHeight: '400px' }}>
      {STAGES.map((stage) => {
        const items = grouped[stage.value];
        const total = items.reduce((sum, o) => sum + (o.amount || 0), 0);
        const isOver = dragOverStage === stage.value;

        return (
          <div
            key={stage.value}
            className={`flex-shrink-0 w-56 rounded-lg border-2 ${stage.color} ${isOver ? 'ring-2 ring-blue-400' : ''}`}
            onDragOver={(e) => e.preventDefault()}
            onDragEnter={() => {
              dragCounter.current[stage.value] = (dragCounter.current[stage.value] || 0) + 1;
              setDragOverStage(stage.value);
            }}
            onDragLeave={() => {
              dragCounter.current[stage.value] = (dragCounter.current[stage.value] || 0) - 1;
              if (dragCounter.current[stage.value] <= 0) {
                dragCounter.current[stage.value] = 0;
                if (dragOverStage === stage.value) setDragOverStage(null);
              }
            }}
            onDrop={(e) => {
              e.preventDefault();
              dragCounter.current[stage.value] = 0;
              handleDrop(stage.value);
            }}
          >
            <div className="px-3 py-2 border-b border-inherit">
              <h3 className="text-xs font-semibold text-gray-700 uppercase tracking-wide">{stage.label}</h3>
              <div className="flex items-baseline justify-between mt-0.5">
                <span className="text-xs text-gray-500">{items.length} deal{items.length !== 1 ? 's' : ''}</span>
                {total > 0 && <span className="text-xs font-medium text-gray-600">{formatCurrency(total)}</span>}
              </div>
            </div>
            <div className="p-2 space-y-2 min-h-[60px]">
              {items.map((opp) => (
                <div
                  key={opp.id}
                  draggable
                  onDragStart={() => setDragItem(opp.id)}
                  onDragEnd={() => { setDragItem(null); setDragOverStage(null); dragCounter.current = {}; }}
                  className={`bg-white rounded-md shadow-sm border border-gray-200 p-2 cursor-grab active:cursor-grabbing ${
                    dragItem === opp.id ? 'opacity-50' : ''
                  }`}
                >
                  <Link to={`/opportunities/${opp.id}`} className="text-sm font-medium text-blue-600 hover:text-blue-800 block truncate">
                    {opp.name}
                  </Link>
                  <div className="flex items-center justify-between mt-1">
                    {opp.amount != null && <span className="text-xs text-gray-600">{formatCurrency(opp.amount)}</span>}
                    {opp.probability != null && <span className="text-xs text-gray-400">{opp.probability}%</span>}
                  </div>
                  {opp.closes_on && (
                    <p className="text-xs text-gray-400 mt-0.5">{new Date(opp.closes_on).toLocaleDateString()}</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        );
      })}
      {mutation.loading && (
        <div className="fixed bottom-4 right-4 bg-blue-600 text-white text-sm px-3 py-1.5 rounded-md shadow-lg">
          Updating...
        </div>
      )}
    </div>
  );
}
