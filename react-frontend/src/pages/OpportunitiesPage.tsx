import { EntityList } from '../components/EntityList';
import type { Opportunity } from '../types/entities';

function formatCurrency(val?: number): string {
  if (val == null) return '';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
}

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'stage', label: 'Stage' },
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
  return (
    <EntityList<Opportunity>
      title="Opportunities"
      endpoint="/opportunities"
      columns={columns}
      getRowKey={(o) => o.id}
    />
  );
}
