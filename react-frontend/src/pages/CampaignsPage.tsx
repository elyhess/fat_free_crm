import { EntityList, type FilterDef } from '../components/EntityList';
import { campaignFields } from '../config/entityFields';
import type { Campaign } from '../types/entities';

const filterDefs: FilterDef[] = [
  { key: 'name', label: 'Name', type: 'text' },
  { key: 'status', label: 'Status', type: 'select', operator: 'eq', options: [
    { value: 'planned', label: 'Planned' },
    { value: 'started', label: 'Started' },
    { value: 'completed', label: 'Completed' },
    { value: 'on_hold', label: 'On Hold' },
    { value: 'called_off', label: 'Called Off' },
  ]},
];

function formatCurrency(val?: number): string {
  if (val == null) return '';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
}

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  {
    key: 'budget',
    label: 'Budget',
    render: (c: Campaign) => formatCurrency(c.budget),
  },
  { key: 'leads_count', label: 'Leads' },
  { key: 'opportunities_count', label: 'Opportunities' },
  {
    key: 'starts_on',
    label: 'Starts',
    render: (c: Campaign) =>
      c.starts_on ? new Date(c.starts_on).toLocaleDateString() : '',
  },
];

export function CampaignsPage() {
  return (
    <EntityList<Campaign>
      title="Campaigns"
      endpoint="/campaigns"
      columns={columns}
      getRowKey={(c) => c.id}
      formFields={campaignFields}
      detailPath={(c) => `/campaigns/${c.id}`}
      filterFields={filterDefs}
    />
  );
}
