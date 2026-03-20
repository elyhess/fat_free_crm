import { EntityList } from '../components/EntityList';
import { leadFields } from '../config/entityFields';
import type { Lead } from '../types/entities';

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
    />
  );
}
