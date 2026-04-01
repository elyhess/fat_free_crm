import { EntityList, type FilterDef } from '../components/EntityList';
import { leadFields } from '../config/entityFields';
import type { Lead } from '../types/entities';

const filterDefs: FilterDef[] = [
  { key: 'first_name', label: 'First Name', type: 'text' },
  { key: 'last_name', label: 'Last Name', type: 'text' },
  { key: 'company', label: 'Company', type: 'text' },
  { key: 'status', label: 'Status', type: 'select', operator: 'eq', options: [
    { value: 'new', label: 'New' },
    { value: 'contacted', label: 'Contacted' },
    { value: 'converted', label: 'Converted' },
    { value: 'rejected', label: 'Rejected' },
  ]},
];

const columns = [
  {
    key: 'last_name',
    label: 'Name',
    render: (l: Lead) => `${l.first_name} ${l.last_name}`,
  },
  { key: 'company', label: 'Company' },
  { key: 'status', label: 'Status' },
  { key: 'email', label: 'Email' },
  { key: 'rating', label: 'Rating' },
];

export function LeadsPage() {
  return (
    <EntityList<Lead>
      title="Leads"
      endpoint="/leads"
      columns={columns}
      getRowKey={(l) => l.id}
      formFields={leadFields}
      detailPath={(l) => `/leads/${l.id}`}
      filterFields={filterDefs}
    />
  );
}
